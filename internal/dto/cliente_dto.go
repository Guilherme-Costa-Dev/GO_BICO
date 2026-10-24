package dto

type AtualizarClienteDTO struct {
	Nome        string `json:"nome" firestore:"nome"`
	Cpf         string `json:"cpf" firestore:"cpf"`
	Cep         string `json:"cep" firestore:"cep"`
	Numero      string `json:"numero" firestore:"numero"`
	Complemento string `json:"complemento" firestore:"complemento"`
}

type AdicionarAvaliacaoDTO struct {
	PrestadorID string `json:"prestadorID"`
	Nota        int    `json:"nota"`
	Comentario  string `json:"comentario"`
}

type CriarClienteDTO struct {
	Nome        string `json:"nome"`
	Cpf         string `json:"cpf"`
	Cep         string `json:"cep"`
	Email       string `json:"email"`
	Senha       string `json:"senha"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
}
