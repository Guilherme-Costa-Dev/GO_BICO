package utils

import (
	"math"
	"testing"
)

func TestNormalizar(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado string
	}{
		{"José", "jose"},
		{"São Paulo", "sao paulo"},
		{"  ELETRICISTA  ", "eletricista"},
		{"João-Silva", "joao silva"},
		{"Ação & Reação!", "acao reacao"},
		{"", ""},
	}

	for _, caso := range casos {
		if obtido := Normalizar(caso.entrada); obtido != caso.esperado {
			t.Errorf("Normalizar(%q): esperado %q, veio %q", caso.entrada, caso.esperado, obtido)
		}
	}
}

func TestContemTexto(t *testing.T) {
	if !ContemTexto("jose", "José Antônio") {
		t.Error("esperado achar 'jose' em 'José Antônio'")
	}
	if !ContemTexto("sao", "São Paulo") {
		t.Error("esperado achar 'sao' em 'São Paulo'")
	}
	if !ContemTexto("silva", "joaosilva") {
		t.Error("esperado achar 'silva' no meio de 'joaosilva'")
	}
	if ContemTexto("pedreiro", "eletricista") {
		t.Error("não deveria achar 'pedreiro' em 'eletricista'")
	}
	if !ContemTexto("", "qualquer coisa") {
		t.Error("termo vazio deveria casar com qualquer texto")
	}
}

func TestNotaPonderada(t *testing.T) {
	const epsilon = 1e-9
	quaseIgual := func(a, b float64) bool { return math.Abs(a-b) < epsilon }

	if nota := NotaPonderada(5, 0, 4); nota != 5 {
		t.Errorf("sem avaliações deveria usar a nota salva, veio %v", nota)
	}

	if nota := NotaPonderada(5, 1, 4.6); !quaseIgual(nota, (5+5*4.6)/6) {
		t.Errorf("nota ponderada com 1 avaliação: veio %v", nota)
	}

	estabelecido := NotaPonderada(4.8, 10, 4.6)
	isolado := NotaPonderada(5, 1, 4.6)
	if estabelecido <= isolado {
		t.Errorf("4,8 com 10 avaliações deveria superar 5,0 com 1: %v <= %v", estabelecido, isolado)
	}
}
