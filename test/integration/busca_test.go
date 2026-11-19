package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"bico/internal/domain"
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

func definirAvaliacoes(t *testing.T, uid string, nota float64, total int) {
	t.Helper()

	avaliacoes := make([]domain.Avaliacao, total)
	for i := range avaliacoes {
		avaliacoes[i] = domain.Avaliacao{Nome: "Cliente Teste", Nota: int(nota), Data: time.Now()}
	}

	if _, err := db.Collection("prestadores").Doc(uid).Update(ctx, []firestore.Update{
		{Path: "notaMedia", Value: nota},
		{Path: "avaliacoes", Value: avaliacoes},
	}); err != nil {
		t.Fatalf("definir avaliações de %s: %v", uid, err)
	}
}

func definirCampos(t *testing.T, uid string, campos map[string]any) {
	t.Helper()

	updates := make([]firestore.Update, 0, len(campos))
	for path, valor := range campos {
		updates = append(updates, firestore.Update{Path: path, Value: valor})
	}
	if _, err := db.Collection("prestadores").Doc(uid).Update(ctx, updates); err != nil {
		t.Fatalf("definir campos de %s: %v", uid, err)
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

func TestBuscarPrestadoresTermoUsernameCompleto(t *testing.T) {
	p := criarPrestador(t, "01001000")

	rec, lista := buscarPrestadores(t, p.UID, "termo="+url.QueryEscape(p.Username))
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

func TestBuscarPrestadoresPreencheDistanciaSemOrdenar(t *testing.T) {
	c := criarCliente(t) // CEP 01001000 (São Paulo)

	tipo := "teste-bico-semord-" + sufixoUnico()
	perto := criarPrestadorComTipo(t, "01002000", tipo) // São Paulo, ~0,4 km do cliente
	longe := criarPrestadorComTipo(t, "20040020", tipo) // Rio de Janeiro, ~357 km

	definirNotaMedia(t, perto.UID, 3)
	definirNotaMedia(t, longe.UID, 5)

	rec, lista := buscarPrestadores(t, c.UID, "tipo="+url.QueryEscape(tipo)+"&inicio=0&limite=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 2 {
		t.Fatalf("esperado 2 prestadores, veio %d", len(lista))
	}

	if lista[0].ID != longe.UID || lista[1].ID != perto.UID {
		t.Fatalf("ordem inesperada: esperado [%s %s], veio [%s %s]",
			longe.UID, perto.UID, lista[0].ID, lista[1].ID)
	}

	if d := lista[1].DistanciaKm; d < 0 || d > 5 {
		t.Errorf("distância do prestador próximo: esperado [0, 5], veio %v", d)
	}
	if d := lista[0].DistanciaKm; d < 300 || d > 450 {
		t.Errorf("distância do prestador distante: esperado ~357, veio %v", d)
	}
}

func TestBuscarPrestadoresSemLocalizacaoMantemDistanciaZero(t *testing.T) {
	tipo := "teste-bico-semloc-" + sufixoUnico()
	p := criarPrestadorComTipo(t, "01002000", tipo)

	// UID sem documento em "clientes": a busca padrão não pode falhar nem
	// inventar distância, apenas devolver distanciaKm zerado.
	rec, lista := buscarPrestadores(t, p.UID, "tipo="+url.QueryEscape(tipo))
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 1 {
		t.Fatalf("esperado 1 prestador, veio %d", len(lista))
	}
	if lista[0].DistanciaKm != 0 {
		t.Errorf("distanciaKm: esperado 0 sem localização do cliente, veio %v", lista[0].DistanciaKm)
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

func TestBuscarPrestadoresSemTermoOrdenaPorNotaPonderada(t *testing.T) {
	tipo := "teste-bico-ponderada-" + sufixoUnico()

	isolado := criarPrestadorComTipo(t, "01001000", tipo)
	estabelecido := criarPrestadorComTipo(t, "01001000", tipo)
	antigo := criarPrestadorComTipo(t, "01001000", tipo)

	definirAvaliacoes(t, isolado.UID, 5.0, 1)
	definirAvaliacoes(t, estabelecido.UID, 4.8, 10)
	definirAvaliacoes(t, antigo.UID, 4.0, 50)

	rec, lista := buscarPrestadores(t, "", "tipo="+url.QueryEscape(tipo)+"&limite=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 3 {
		t.Fatalf("esperado 3 prestadores, veio %d", len(lista))
	}

	esperado := []string{estabelecido.UID, isolado.UID, antigo.UID}
	for i, uid := range esperado {
		if lista[i].ID != uid {
			t.Errorf("posição %d: esperado %q, veio %q", i, uid, lista[i].ID)
		}
	}
}

func TestBuscarPrestadoresTermoParcial(t *testing.T) {
	tipo := "teste-bico-parcial-" + sufixoUnico()
	p := criarPrestadorComTipo(t, "01001000", tipo)

	definirCampos(t, p.UID, map[string]any{
		"nome":         "João da Silva",
		"username":     "joao.silva." + sufixoUnico(),
		"localAtuacao": "São Paulo",
	})

	rec, lista := buscarPrestadores(t, "", "tipo="+url.QueryEscape(tipo)+"&termo=silva")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 1 {
		t.Fatalf("esperado 1 prestador para 'silva', veio %d", len(lista))
	}
	if lista[0].ID != p.UID {
		t.Errorf("id: esperado %q, veio %q", p.UID, lista[0].ID)
	}
}

func TestBuscarPrestadoresTermoAcentoECaixa(t *testing.T) {
	tipo := "teste-bico-acento-" + sufixoUnico()
	p := criarPrestadorComTipo(t, "01001000", tipo)

	definirCampos(t, p.UID, map[string]any{
		"nome":         "José Antônio",
		"localAtuacao": "São Paulo",
	})

	casos := []string{"JOSE", "jose", "sao", "SÃO"}

	for _, termo := range casos {
		rec, lista := buscarPrestadores(t, "", "tipo="+url.QueryEscape(tipo)+"&termo="+url.QueryEscape(termo))
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: esperado 200, veio %d: %s", termo, rec.Code, rec.Body.String())
		}
		if len(lista) != 1 || lista[0].ID != p.UID {
			t.Errorf("%q: esperado o prestador %q, veio %v", termo, p.UID, lista)
		}
	}
}

func TestBuscarPrestadoresTermoPorTipoServico(t *testing.T) {
	tipo := "teste-bico-especialidade-" + sufixoUnico()
	p := criarPrestadorComTipo(t, "01001000", tipo)
	criarPrestadorComTipo(t, "01001000", "teste-bico-outra-"+sufixoUnico())

	rec, lista := buscarPrestadores(t, "", "termo="+url.QueryEscape(tipo))
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 1 {
		t.Fatalf("esperado 1 prestador, veio %d", len(lista))
	}
	if lista[0].ID != p.UID {
		t.Errorf("id: esperado %q, veio %q", p.UID, lista[0].ID)
	}
}

func TestBuscarPrestadoresPaginacaoComTermo(t *testing.T) {
	tipo := "teste-bico-pag-termo-" + sufixoUnico()
	criarGrupoComNotas(t, tipo, 5, 4, 3)

	casos := []struct {
		query          string
		notasEsperadas []float64
	}{
		{"inicio=0&limite=2", []float64{5, 4}},
		{"inicio=2&limite=2", []float64{3}},
	}

	for _, caso := range casos {
		rec, lista := buscarPrestadores(t, "", "tipo="+url.QueryEscape(tipo)+"&termo=prestador&"+caso.query)
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

func TestBuscarPrestadoresTermoSemResultado(t *testing.T) {
	tipo := "teste-bico-semresultado-" + sufixoUnico()
	criarPrestadorComTipo(t, "01001000", tipo)

	rec, lista := buscarPrestadores(t, "", "tipo="+url.QueryEscape(tipo)+"&termo=zzzznaoexiste")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 0 {
		t.Fatalf("esperado lista vazia, veio %d", len(lista))
	}
	if rec.Body.String() != "[]\n" {
		t.Errorf("esperado array vazio, veio %q", rec.Body.String())
	}
}

func TestBuscarPrestadoresProximosComTermo(t *testing.T) {
	c := criarCliente(t)
	tipo := "teste-bico-dist-termo-" + sufixoUnico()

	perto := criarPrestadorComTipo(t, "01001000", tipo)
	longe := criarPrestadorComTipo(t, "20040020", tipo)

	definirCampos(t, perto.UID, map[string]any{"nome": "Pedreiro Perto"})
	definirCampos(t, longe.UID, map[string]any{"nome": "Pedreiro Longo"})

	rec, lista := buscarPrestadores(t, c.UID,
		"busca=proximos&tipo="+url.QueryEscape(tipo)+"&raio=50&termo=pedreiro&inicio=0&limite=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(lista) != 1 {
		t.Fatalf("esperado só o prestador de São Paulo, veio %d", len(lista))
	}
	if lista[0].ID != perto.UID {
		t.Errorf("id: esperado %q, veio %q", perto.UID, lista[0].ID)
	}
}
