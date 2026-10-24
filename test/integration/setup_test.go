package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"bico/internal/config"
	"bico/internal/handler"
	"bico/internal/server"
	"bico/internal/utils"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
)

var (
	ctx        = context.Background()
	db         *firestore.Client
	authClient *auth.Client
	clienteH   *handler.ClienteHandler
	prestadorH *handler.PrestadorHandler
	router     http.Handler
	cepAtivo   bool
)

func TestMain(m *testing.M) {
	var err error

	db, authClient, err = config.InitFirebase(ctx, "../../serviceAccountKey.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao inicializar o firebase: %v\n", err)
		os.Exit(1)
	}

	clienteH = handler.NewClienteHandler(db, authClient)
	prestadorH = handler.NewPrestadorHandler(db, authClient)
	router = server.New(db, authClient)

	// A criação de usuários depende da geocodificação do CEP via BrasilAPI.
	_, _, err = utils.BuscarCoordPorCep("01001000")
	cepAtivo = err == nil

	codigo := m.Run()
	_ = db.Close()
	os.Exit(codigo)
}

// exigirCep evita que uma indisponibilidade da BrasilAPI derrube a suíte.
func exigirCep(t *testing.T) {
	t.Helper()
	if !cepAtivo {
		t.Skip("BrasilAPI indisponível: testes que criam usuários dependem da geocodificação do CEP")
	}
}

func emailUnico() string {
	return fmt.Sprintf("teste-bico-%s@example.com", sufixoUnico())
}

func sufixoUnico() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1_000_000))
}

type clienteTeste struct {
	UID   string
	Email string
}

func criarCliente(t *testing.T) clienteTeste {
	t.Helper()
	exigirCep(t)

	email := emailUnico()
	rec := chamarJSON(t, router, http.MethodPost, "/clientes", "", map[string]any{
		"nome":        "Cliente Teste",
		"cpf":         "12345678900",
		"cep":         "01001000",
		"email":       email,
		"senha":       "senha123456",
		"numero":      "10",
		"complemento": "apto 1",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar cliente: esperado 201, veio %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || resp.ID == "" {
		t.Fatalf("criar cliente: resposta sem id: %s", rec.Body.String())
	}

	uid := resp.ID
	t.Cleanup(func() { limparUsuario(t, "clientes", uid) })

	return clienteTeste{UID: uid, Email: email}
}

type prestadorTeste struct {
	UID      string
	Email    string
	Username string
	Tipo     string
}

func criarPrestador(t *testing.T, cep string) prestadorTeste {
	t.Helper()
	return criarPrestadorComTipo(t, cep, "teste-bico-"+sufixoUnico())
}

func criarPrestadorComTipo(t *testing.T, cep, tipo string) prestadorTeste {
	t.Helper()
	exigirCep(t)

	email := emailUnico()
	username := "teste-bico-" + sufixoUnico()

	rec := chamarJSON(t, router, http.MethodPost, "/prestadores", "", map[string]any{
		"nome":         "Prestador Teste",
		"cpf":          "12345678900",
		"cep":          cep,
		"email":        email,
		"senha":        "senha123456",
		"username":     username,
		"tiposServico": []string{tipo},
		"localAtuacao": "São Paulo",
		"sobre":        "prestador criado por teste",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar prestador: esperado 201, veio %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || resp.ID == "" {
		t.Fatalf("criar prestador: resposta sem id: %s", rec.Body.String())
	}

	uid := resp.ID
	t.Cleanup(func() { limparUsuario(t, "prestadores", uid) })

	return prestadorTeste{UID: uid, Email: email, Username: username, Tipo: tipo}
}

// limparUsuario remove os dois lados (Firestore e Auth). Ignora registros que o
// próprio teste já apagou e nunca falha o teste por causa da limpeza.
func limparUsuario(t *testing.T, colecao, uid string) {
	t.Helper()

	if _, err := db.Collection(colecao).Doc(uid).Delete(ctx); err != nil {
		t.Logf("cleanup: erro ao apagar %s/%s: %v", colecao, uid, err)
	}
	if err := authClient.DeleteUser(ctx, uid); err != nil && !auth.IsUserNotFound(err) {
		t.Logf("cleanup: erro ao apagar usuário %s: %v", uid, err)
	}
}

func corpoJSON(t *testing.T, corpo any) io.Reader {
	t.Helper()
	if corpo == nil {
		return nil
	}
	b, err := json.Marshal(corpo)
	if err != nil {
		t.Fatalf("marshal do corpo: %v", err)
	}
	return bytes.NewReader(b)
}

// chamarJSON exercita o roteador real, incluindo o middleware de autenticação.
func chamarJSON(t *testing.T, h http.Handler, metodo, caminho, token string, corpo any) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(metodo, caminho, corpoJSON(t, corpo))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// chamarComo chama o handler diretamente, injetando o UID como o middleware faria.
// Usado nas rotas protegidas, onde obter um ID token real exigiria a Web API key.
func chamarComo(t *testing.T, h http.HandlerFunc, metodo, caminho, uid string, corpo any) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(metodo, caminho, corpoJSON(t, corpo))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "userUID", uid))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func lerDoc[T any](t *testing.T, colecao, uid string) T {
	t.Helper()

	doc, err := db.Collection(colecao).Doc(uid).Get(ctx)
	if err != nil {
		t.Fatalf("ler %s/%s: %v", colecao, uid, err)
	}

	var v T
	if err := doc.DataTo(&v); err != nil {
		t.Fatalf("converter %s/%s: %v", colecao, uid, err)
	}
	return v
}
