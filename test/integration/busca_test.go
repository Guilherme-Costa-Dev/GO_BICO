package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"bico/internal/dto"

	"cloud.google.com/go/firestore"
)

func buscarPrestadores(t *testing.T, uid, query string) (*httptest.ResponseRecorder, []dto.PrestadorPublicoDTO) {
	t.Helper()

	rec := chamarComo(t, prestadorH.BuscarPrestadores, http.MethodGet, "/prestadores/busca?"+query, uid, nil)

	var lista []dto.PrestadorPublicoDTO
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
			t.Fatalf("resposta inválida: %v (%s)", err, rec.Body.String())
		}
	}
	return rec, lista
}

func definirNotaMedia(t *testing.T, uid string, nota float64) {
	t.Helper()

	if _, err := db.Collection("prestadores").Doc(uid).Update(ctx, []firestore.Update{
		{Path: "notaMedia", Value: nota},
	}); err != nil {
		t.Fatalf("definir notaMedia de %s: %v", uid, err)
	}
}

// criarGrupoComNotas cria prestadores com o mesmo tipo e notas distintas,
// devolvendo os UIDs na mesma ordem das notas.
func criarGrupoComNotas(t *testing.T, tipo string, notas ...float64) []string {
	t.Helper()

	ids := make([]string, 0, len(notas))
	for _, nota := range notas {
		p := criarPrestadorComTipo(t, "01001000", tipo)
		definirNotaMedia(t, p.UID, nota)
		ids = append(ids, p.UID)
	}
	return ids
}

func TestBuscarPrestadoresFiltroUsername(t *testing.T) {
	p := criarPrestador(t, "01001000")

	rec, lista := buscarPrestadores(t, p.UID, "username="+url.QueryEscape(p.Username))
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 1 {
		t.Fatalf("esperado 1 prestador para o username %q, veio %d", p.Username, len(lista))
	}
	if lista[0].ID != p.UID {
		t.Errorf("id: esperado %q, veio %q", p.UID, lista[0].ID)
	}
}

func TestBuscarPrestadoresFiltroTipoEOrdenacao(t *testing.T) {
	tipo := "teste-bico-tipo-" + sufixoUnico()
	criarGrupoComNotas(t, tipo, 3, 5, 4)

	rec, lista := buscarPrestadores(t, "", "tipo="+url.QueryEscape(tipo)+"&inicio=0&limite=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 3 {
		t.Fatalf("esperado 3 prestadores para o tipo %q, veio %d", tipo, len(lista))
	}

	esperado := []float64{5, 4, 3}
	for i, nota := range esperado {
		if lista[i].NotaMedia != nota {
			t.Errorf("posição %d: esperado notaMedia %v, veio %v", i, nota, lista[i].NotaMedia)
		}
	}
}

func TestBuscarPrestadoresPaginacao(t *testing.T) {
	tipo := "teste-bico-pag-" + sufixoUnico()
	criarGrupoComNotas(t, tipo, 5, 4, 3)

	casos := []struct {
		query          string
		notasEsperadas []float64
	}{
		{"inicio=0&limite=2", []float64{5, 4}},
		{"inicio=2&limite=2", []float64{3}},
		{"inicio=10&limite=2", []float64{}},
	}

	for _, caso := range casos {
		rec, lista := buscarPrestadores(t, "", "tipo="+url.QueryEscape(tipo)+"&"+caso.query)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: esperado 200, veio %d: %s", caso.query, rec.Code, rec.Body.String())
		}
		if len(lista) != len(caso.notasEsperadas) {
			t.Errorf("%s: esperado %d itens, veio %d", caso.query, len(caso.notasEsperadas), len(lista))
			continue
		}
		for i, nota := range caso.notasEsperadas {
			if lista[i].NotaMedia != nota {
				t.Errorf("%s: posição %d: esperado notaMedia %v, veio %v", caso.query, i, nota, lista[i].NotaMedia)
			}
		}
	}
}

func TestBuscarPrestadoresPaginacaoInvalida(t *testing.T) {
	casos := []string{
		"inicio=-1",
		"inicio=abc",
		"limite=-1",
		"limite=101",
		"limite=abc",
	}

	for _, query := range casos {
		rec, _ := buscarPrestadores(t, "uid-inexistente", query)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperado 400, veio %d", query, rec.Code)
		}
	}
}

func TestBuscarPrestadoresRaioInvalido(t *testing.T) {
	c := criarCliente(t)

	casos := []string{
		"busca=proximos&raio=0",
		"busca=proximos&raio=-1",
		"busca=proximos&raio=abc",
	}

	for _, query := range casos {
		rec, _ := buscarPrestadores(t, c.UID, query)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperado 400, veio %d: %s", query, rec.Code, rec.Body.String())
		}
	}
}

func TestBuscarPrestadoresProximosRaio(t *testing.T) {
	c := criarCliente(t)
	tipo := "teste-bico-dist-" + sufixoUnico()

	perto := criarPrestadorComTipo(t, "01001000", tipo) // São Paulo (mesmo CEP do cliente)
	criarPrestadorComTipo(t, "20040020", tipo)          // Rio de Janeiro (~357 km)

	rec, lista := buscarPrestadores(t, c.UID, "busca=proximos&tipo="+url.QueryEscape(tipo)+"&raio=50&inicio=0&limite=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 1 {
		t.Fatalf("esperado só o prestador de São Paulo, veio %d", len(lista))
	}
	if lista[0].ID != perto.UID {
		t.Errorf("id: esperado %q, veio %q", perto.UID, lista[0].ID)
	}
	if lista[0].DistanciaKm > 50 {
		t.Errorf("distanciaKm: esperado <= 50, veio %v", lista[0].DistanciaKm)
	}
}

func TestBuscarPrestadoresProximosSemLocalizacao(t *testing.T) {
	c := criarCliente(t)

	if _, err := db.Collection("clientes").Doc(c.UID).Update(ctx, []firestore.Update{
		{Path: "lat", Value: float64(0)},
		{Path: "lng", Value: float64(0)},
	}); err != nil {
		t.Fatalf("zerar localização do cliente: %v", err)
	}

	rec, _ := buscarPrestadores(t, c.UID, "busca=proximos")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d: %s", rec.Code, rec.Body.String())
	}
}
