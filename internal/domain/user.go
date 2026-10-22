package domain

import "time"

type UserBase struct {
	ID    string `json:"id" firestore:"-"`
	Nome  string `json:"nome" firestore:"nome"`
	Cpf   string `json:"cpf" firestore:"cpf"`
	Cep   string `json:"cep" firestore:"cep"`
	Email string `json:"email" firestore:"email"`
	Senha string `json:"senha,omitempty" firestore:"-"`
}

type Avaliacao struct {
	Nome       string    `json:"nome" firestore:"nome"`
	Nota       int       `json:"nota" firestore:"nota"`
	Comentario string    `json:"comentario" firestore:"comentario"`
	Data       time.Time `json:"data" firestore:"data"`
}

type Cliente struct {
	UserBase
	Numero      string   `json:"numero" firestore:"numero"`
	Complemento string   `json:"complemento" firestore:"complemento"`
	Favoritos   []string `json:"favoritos" firestore:"favoritos"`
}

type Prestador struct {
	UserBase
	Username         string      `json:"username" firestore:"username"`
	TiposServico     []string    `json:"tiposServico" firestore:"tiposServico"`
	LocalAtuacao     string      `json:"localAtuacao" firestore:"localAtuacao"`
	FotoPerfil       string      `json:"fotoPerfil" firestore:"fotoPerfil"`
	FotoPaginaPerfil string      `json:"fotoPaginaPerfil" firestore:"fotoPaginaPerfil"`
	FotosServicos    []string    `json:"fotosServicos" firestore:"fotosServicos"`
	Sobre            string      `json:"sobre" firestore:"sobre"`
	NotaMedia        float64     `json:"notaMedia" firestore:"notaMedia"`
	Avaliacoes       []Avaliacao `json:"avaliacoes" firestore:"avaliacoes"`
}
