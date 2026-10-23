package handler

import (
	"bico/internal/domain"
)

type PrestadorPublicoDTO struct {
	ID               string             `json:"id"`
	Nome             string             `json:"nome"`
	TiposServico     []string           `json:"tiposServico"`
	NotaMedia        float64            `json:"notaMedia"`
	Email            string             `json:"email"`
	LocalAtuacao     string             `json:"localAtuacao"`
	Username         string             `json:"username"`
	FotoPerfil       string             `json:"fotoPerfil"`
	FotoPaginaPerfil string             `json:"fotoPaginaPerfil"`
	FotosServicos    []string           `json:"fotosServicos"`
	Sobre            string             `json:"sobre"`
	TotalAvaliacoes  int                `json:"totalAvaliacoes"`
	Avaliacoes       []domain.Avaliacao `json:"avaliacoes"`
	DistanciaKm      float64            `json:"distanciaKm"`
}

type AtualizarPrestadorDTO struct {
	Nome             string   `json:"nome"`
	TiposServico     []string `json:"tiposServico"`
	LocalAtuacao     string   `json:"localAtuacao"`
	Username         string   `json:"username"`
	Cep              string   `json:"cep"`
	FotoPerfil       string   `json:"fotoPerfil"`
	FotoPaginaPerfil string   `json:"fotoPaginaPerfil"`
	FotosServicos    []string `json:"fotosServicos"`
	Sobre            string   `json:"sobre"`
}

type BuscarPrestadorDTO struct {
	Username  string `json:"username"`
	Tipo      string `json:"tipo"`
	Inicio    string `json:"inicio"`
	Limite    string `json:"limite"`
	Busca     string `json:"busca"`
	IDCliente string `json:"idCliente"`
	Raio      string `json:"raio"`
}