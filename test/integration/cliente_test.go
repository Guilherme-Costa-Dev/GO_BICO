package integration_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"bico/internal/domain"
)

func TestRotasProtegidasSemToken(t *testing.T) {
	rotas := []struct {
		metodo  string
		caminho string
	}{
		{http.MethodGet, "/clientes"},
		{http.MethodPut, "/clientes"},
		{http.MethodDelete, "/clientes"},
		{http.MethodPut, "/clientes/favoritos"},
		{http.MethodGet, "/clientes/favoritos"},
		{http.MethodDelete, "/clientes/favoritos"},
		{http.MethodPut, "/clientes/avaliacoes"},
		{http.MethodPut, "/clientes/email"},
		{http.MethodGet, "/prestadores"},
		{http.MethodPut, "/prestadores"},
		{http.MethodDelete, "/prestadores"},
		{http.MethodGet, "/prestadores/busca"},
		{http.MethodPut, "/prestadores/email"},
	}

	for _, r := range rotas {
		rec := chamarJSON(t, router, r.metodo, r.caminho, "", nil)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: esperado 401, veio %d", r.metodo, r.caminho, rec.Code)
		}
	}
}

func TestTokenInvalido(t *testing.T) {
	rec := chamarJSON(t, router, http.MethodGet, "/clientes", "token-invalido", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, veio %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCriarEBuscarCliente(t *testing.T) {
	c := criarCliente(t)

	rec := chamarComo(t, clienteH.DadosCliente, http.MethodGet, "/clientes", c.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	var corpo map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("resposta não é JSON: %v", err)
	}

	if _, ok := corpo["lat"]; !ok {
		t.Error("resposta sem a chave 'lat' (minúscula)")
	}
	if _, ok := corpo["lng"]; !ok {
		t.Error("resposta sem a chave 'lng' (minúscula)")
	}
	if _, ok := corpo["Lat"]; ok {
		t.Error("resposta com 'Lat' (maiúscula)")
	}
	if _, ok := corpo["Lng"]; ok {
		t.Error("resposta com 'Lng' (maiúscula)")
	}
	if _, ok := corpo["senha"]; ok {
		t.Error("resposta expõe 'senha'")
	}
	if corpo["email"] != c.Email {
		t.Errorf("email: esperado %q, veio %v", c.Email, corpo["email"])
	}
}

func TestAtualizarClienteSubstituicao(t *testing.T) {
	c := criarCliente(t)

	rec := chamarComo(t, clienteH.AtualizarCliente, http.MethodPut, "/clientes", c.UID, map[string]any{
		"nome":        "Novo Nome",
		"cpf":         "98765432100",
		"cep":         "01001000",
		"numero":      "20",
		"complemento": "",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	salvo := lerDoc[domain.Cliente](t, "clientes", c.UID)
	if salvo.Nome != "Novo Nome" {
		t.Errorf("nome: esperado %q, veio %q", "Novo Nome", salvo.Nome)
	}
	if salvo.Cpf != "98765432100" {
		t.Errorf("cpf: esperado %q, veio %q", "98765432100", salvo.Cpf)
	}
	if salvo.Numero != "20" {
		t.Errorf("numero: esperado %q, veio %q", "20", salvo.Numero)
	}
	if salvo.Complemento != "" {
		t.Errorf("complemento deveria ter sido apagado, veio %q", salvo.Complemento)
	}
}

func TestAtualizarClienteObrigatorios(t *testing.T) {
	c := criarCliente(t)

	casos := map[string]map[string]any{
		"nome": {"nome": "", "cpf": "123", "cep": "01001000", "numero": "1", "complemento": ""},
		"cpf":  {"nome": "Teste", "cpf": "", "cep": "01001000", "numero": "1", "complemento": ""},
		"cep":  {"nome": "Teste", "cpf": "123", "cep": "", "numero": "1", "complemento": ""},
	}

	for campo, corpo := range casos {
		rec := chamarComo(t, clienteH.AtualizarCliente, http.MethodPut, "/clientes", c.UID, corpo)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s vazio: esperado 400, veio %d", campo, rec.Code)
		}
	}
}

func TestAtualizarClienteRecalculaCoordenadas(t *testing.T) {
	exigirCep(t)
	c := criarCliente(t)

	antes := lerDoc[domain.Cliente](t, "clientes", c.UID)

	rec := chamarComo(t, clienteH.AtualizarCliente, http.MethodPut, "/clientes", c.UID, map[string]any{
		"nome":        "Cliente Teste",
		"cpf":         "12345678900",
		"cep":         "20040020",
		"numero":      "10",
		"complemento": "apto 1",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	depois := lerDoc[domain.Cliente](t, "clientes", c.UID)
	if depois.Cep != "20040020" {
		t.Errorf("cep: esperado %q, veio %q", "20040020", depois.Cep)
	}
	if depois.Lat == antes.Lat && depois.Lng == antes.Lng {
		t.Errorf("coordenadas não foram recalculadas: %v,%v", depois.Lat, depois.Lng)
	}
}

func TestDeletarCliente(t *testing.T) {
	c := criarCliente(t)

	rec := chamarComo(t, clienteH.DeletarCliente, http.MethodDelete, "/clientes", c.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	if _, err := db.Collection("clientes").Doc(c.UID).Get(ctx); err == nil {
		t.Error("documento ainda existe no Firestore")
	}
	if _, err := authClient.GetUser(ctx, c.UID); err == nil {
		t.Error("usuário ainda existe no Firebase Auth")
	}
}

func TestFavoritos(t *testing.T) {
	c := criarCliente(t)
	p := criarPrestador(t, "01001000")

	rec := chamarComo(t, clienteH.AdicionarFavorito, http.MethodPut, "/clientes/favoritos?id="+p.UID, c.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("adicionar favorito: esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	rec = chamarComo(t, clienteH.ListarFavoritos, http.MethodGet, "/clientes/favoritos", c.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("listar favoritos: esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	var lista struct {
		Favoritos []struct {
			ID string `json:"id"`
		} `json:"favoritos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
		t.Fatalf("listar favoritos: resposta inválida: %v", err)
	}
	encontrado := false
	for _, f := range lista.Favoritos {
		if f.ID == p.UID {
			encontrado = true
		}
	}
	if !encontrado {
		t.Errorf("prestador %s não apareceu na lista de favoritos", p.UID)
	}

	rec = chamarComo(t, clienteH.DeletarFavorito, http.MethodDelete, "/clientes/favoritos?id="+p.UID, c.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("deletar favorito: esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	salvo := lerDoc[domain.Cliente](t, "clientes", c.UID)
	if len(salvo.Favoritos) != 0 {
		t.Errorf("favoritos deveriam estar vazios, veio %v", salvo.Favoritos)
	}
}

func TestAdicionarAvaliacao(t *testing.T) {
	c := criarCliente(t)
	p := criarPrestador(t, "01001000")

	rec := chamarComo(t, clienteH.AdicionarAvaliacao, http.MethodPut, "/clientes/avaliacoes", c.UID, map[string]any{
		"prestadorID": p.UID,
		"nota":        6,
		"comentario":  "fora do intervalo",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("nota 6: esperado 400, veio %d", rec.Code)
	}

	rec = chamarComo(t, clienteH.AdicionarAvaliacao, http.MethodPut, "/clientes/avaliacoes", c.UID, map[string]any{
		"prestadorID": p.UID,
		"nota":        4,
		"comentario":  "muito bom",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("avaliação válida: esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	salvo := lerDoc[domain.Prestador](t, "prestadores", p.UID)
	if len(salvo.Avaliacoes) != 1 {
		t.Fatalf("esperado 1 avaliação, veio %d", len(salvo.Avaliacoes))
	}
	if salvo.NotaMedia != 4 {
		t.Errorf("notaMedia: esperado 4, veio %v", salvo.NotaMedia)
	}
	if salvo.Avaliacoes[0].Comentario != "muito bom" {
		t.Errorf("comentario: veio %q", salvo.Avaliacoes[0].Comentario)
	}
}

func TestMudarEmailCliente(t *testing.T) {
	c := criarCliente(t)
	novo := emailUnico()

	rec := chamarComo(t, clienteH.MudarEmail, http.MethodPut, "/clientes/email?email="+url.QueryEscape(novo), c.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	u, err := authClient.GetUser(ctx, c.UID)
	if err != nil {
		t.Fatalf("buscar usuário no auth: %v", err)
	}
	if u.Email != novo {
		t.Errorf("auth: esperado %q, veio %q", novo, u.Email)
	}

	salvo := lerDoc[domain.Cliente](t, "clientes", c.UID)
	if salvo.Email != novo {
		t.Errorf("firestore: esperado %q, veio %q", novo, salvo.Email)
	}
}
