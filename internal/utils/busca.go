package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const minimoAvaliacoes = 5.0

func Normalizar(s string) string {
	if s == "" {
		return ""
	}

	decomposto := norm.NFD.String(s)

	var b strings.Builder
	b.Grow(len(decomposto))
	for _, r := range decomposto {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(' ')
	}

	return strings.Join(strings.Fields(b.String()), " ")
}

func ContemTexto(termo string, textos ...string) bool {
	termo = Normalizar(termo)
	if termo == "" {
		return true
	}

	for _, texto := range textos {
		if strings.Contains(Normalizar(texto), termo) {
			return true
		}
	}
	return false
}

func NotaPonderada(nota float64, total int, mediaGlobal float64) float64 {
	if total <= 0 {
		return nota
	}

	v := float64(total)
	return (v*nota + minimoAvaliacoes*mediaGlobal) / (v + minimoAvaliacoes)
}
