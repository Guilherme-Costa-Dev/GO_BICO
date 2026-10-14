package handler

type AtualizarClienteDTO struct {
	Nome        string `json:"nome" firestore:"nome,omitempty"`
	Cpf         string `json:"cpf" firestore:"cpf,omitempty"`
	Cep         string `json:"cep" firestore:"cep,omitempty"`
	Numero      string `json:"numero" firestore:"numero,omitempty"`
	Complemento string `json:"complemento" firestore:"complemento,omitempty"`
}