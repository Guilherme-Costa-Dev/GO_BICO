package handler

type PrestadorPublico struct {
	ID               string   `json:"id"`
	Nome             string   `json:"nome"`
	TiposServico     []string `json:"tiposServico"`
	NotaMedia        float64  `json:"notaMedia"`
	Email            string   `json:"email"`
	LocalAtuacao     string   `json:"localAtuacao"`
	Username         string   `json:"username"`
	FotoPerfil       string   `json:"fotoPerfil"`
	FotoPaginaPerfil string   `json:"fotoPaginaPerfil"`
	FotosServicos    []string `json:"fotosServicos"`
	Sobre            string   `json:"sobre"`
	TotalAvaliacoes  int      `json:"totalAvaliacoes"`
}