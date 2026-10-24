package integration_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"bico/internal/domain"
	"bico/internal/dto"
)

func TestCriarEBuscarPrestador(t *testing.T) {
	p := criarPrestador(t, "01001000")

	rec := chamarComo(t, prestadorH.DadosPrestador, http.MethodGet, "/prestadores", p.UID, nil)
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
	if corpo["username"] != p.Username {
		t.Errorf("username: esperado %q, veio %v", p.Username, corpo["username"])
	}
}

func TestAtualizarPrestadorSubstituicao(t *testing.T) {
	p := criarPrestador(t, "01001000")

	rec := chamarComo(t, prestadorH.AtualizarPrestador, http.MethodPut, "/prestadores", p.UID, map[string]any{
		"nome":             "Prestador Renomeado",
		"cpf":              "98765432100",
		"tiposServico":     []string{p.Tipo, "teste-bico-extra"},
		"localAtuacao":     "Rio de Janeiro",
		"username":         p.Username,
		"cep":              "01001000",
		"fotoPerfil":       "https://exemplo.com/perfil.png",
		"fotoPaginaPerfil": "https://exemplo.com/pagina.png",
		"fotosServicos":    []string{"https://exemplo.com/servico1.png"},
		"sobre":            "sobre atualizado",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	salvo := lerDoc[domain.Prestador](t, "prestadores", p.UID)
	if salvo.Nome != "Prestador Renomeado" {
		t.Errorf("nome: esperado %q, veio %q", "Prestador Renomeado", salvo.Nome)
	}
	if salvo.Cpf != "98765432100" {
		t.Errorf("cpf: esperado %q, veio %q", "98765432100", salvo.Cpf)
	}
	if salvo.LocalAtuacao != "Rio de Janeiro" {
		t.Errorf("localAtuacao: veio %q", salvo.LocalAtuacao)
	}
	if salvo.Sobre != "sobre atualizado" {
		t.Errorf("sobre: veio %q", salvo.Sobre)
	}
	if len(salvo.FotosServicos) != 1 || salvo.FotosServicos[0] != "https://exemplo.com/servico1.png" {
		t.Errorf("fotosServicos: veio %v", salvo.FotosServicos)
	}
	if len(salvo.TiposServico) != 2 {
		t.Errorf("tiposServico: esperado 2 itens, veio %v", salvo.TiposServico)
	}
}

func TestAtualizarPrestadorObrigatorios(t *testing.T) {
	p := criarPrestador(t, "01001000")

	completo := func() map[string]any {
		return map[string]any{
			"nome":             "Prestador Teste",
			"cpf":              "12345678900",
			"tiposServico":     []string{p.Tipo},
			"localAtuacao":     "São Paulo",
			"username":         p.Username,
			"cep":              "01001000",
			"fotoPerfil":       "",
			"fotoPaginaPerfil": "",
			"fotosServicos":    []string{},
			"sobre":            "",
		}
	}

	casos := map[string]func(map[string]any){
		"nome":         func(c map[string]any) { c["nome"] = "" },
		"cpf":          func(c map[string]any) { c["cpf"] = "" },
		"cep":          func(c map[string]any) { c["cep"] = "" },
		"username":     func(c map[string]any) { c["username"] = "" },
		"localAtuacao": func(c map[string]any) { c["localAtuacao"] = "" },
		"tiposServico": func(c map[string]any) { c["tiposServico"] = []string{} },
	}

	for campo, alterar := range casos {
		corpo := completo()
		alterar(corpo)

		rec := chamarComo(t, prestadorH.AtualizarPrestador, http.MethodPut, "/prestadores", p.UID, corpo)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s vazio: esperado 400, veio %d", campo, rec.Code)
		}
	}
}

func TestDeletarPrestador(t *testing.T) {
	p := criarPrestador(t, "01001000")

	rec := chamarComo(t, prestadorH.DeletarPrestador, http.MethodDelete, "/prestadores", p.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	if _, err := db.Collection("prestadores").Doc(p.UID).Get(ctx); err == nil {
		t.Error("documento ainda existe no Firestore")
	}
	if _, err := authClient.GetUser(ctx, p.UID); err == nil {
		t.Error("usuário ainda existe no Firebase Auth")
	}
}

func TestBuscarPrestadoresProximos(t *testing.T) {
	exigirCep(t)
	c := criarCliente(t)
	p := criarPrestador(t, "01001000")

	caminho := "/prestadores/busca?busca=proximos&tipo=" + url.QueryEscape(p.Tipo) + "&raio=5&inicio=0&limite=10"
	rec := chamarComo(t, prestadorH.BuscarPrestadores, http.MethodGet, caminho, c.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	var lista []dto.PrestadorPublicoDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
		t.Fatalf("resposta inválida: %v", err)
	}

	if len(lista) != 1 {
		t.Fatalf("esperado 1 prestador para o tipo %q, veio %d", p.Tipo, len(lista))
	}
	if lista[0].ID != p.UID {
		t.Errorf("id: esperado %q, veio %q", p.UID, lista[0].ID)
	}
	if lista[0].DistanciaKm > 5 {
		t.Errorf("distanciaKm: esperado <= 5, veio %v", lista[0].DistanciaKm)
	}
}

func TestMudarEmailPrestador(t *testing.T) {
	p := criarPrestador(t, "01001000")
	novo := emailUnico()

	rec := chamarComo(t, prestadorH.MudarEmail, http.MethodPut, "/prestadores/email?email="+url.QueryEscape(novo), p.UID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	u, err := authClient.GetUser(ctx, p.UID)
	if err != nil {
		t.Fatalf("buscar usuário no auth: %v", err)
	}
	if u.Email != novo {
		t.Errorf("auth: esperado %q, veio %q", novo, u.Email)
	}

	salvo := lerDoc[domain.Prestador](t, "prestadores", p.UID)
	if salvo.Email != novo {
		t.Errorf("firestore: esperado %q, veio %q", novo, salvo.Email)
	}
}
