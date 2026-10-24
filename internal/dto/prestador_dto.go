package dto

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
	Nome             string   `json:"nome" firestore:"nome"`
	TiposServico     []string `json:"tiposServico" firestore:"tiposServico"`
	LocalAtuacao     string   `json:"localAtuacao" firestore:"localAtuacao"`
	Username         string   `json:"username" firestore:"username"`
	Cpf              string   `json:"cpf" firestore:"cpf"`
	Cep              string   `json:"cep" firestore:"cep"`
	FotoPerfil       string   `json:"fotoPerfil" firestore:"fotoPerfil"`
	FotoPaginaPerfil string   `json:"fotoPaginaPerfil" firestore:"fotoPaginaPerfil"`
	FotosServicos    []string `json:"fotosServicos" firestore:"fotosServicos"`
	Sobre            string   `json:"sobre" firestore:"sobre"`
}

type BuscarPrestadorDTO struct {
	Username string `json:"username"`
	Tipo     string `json:"tipo"`
	Inicio   string `json:"inicio"`
	Limite   string `json:"limite"`
	Busca    string `json:"busca"`
	Raio     string `json:"raio"`
}

type CriarPrestadorDTO struct {
	Nome         string   `json:"nome"`
	Cpf          string   `json:"cpf"`
	Cep          string   `json:"cep"`
	Email        string   `json:"email"`
	Senha        string   `json:"senha"`
	Username     string   `json:"username"`
	TiposServico []string `json:"tiposServico"`
	LocalAtuacao string   `json:"localAtuacao"`
}
