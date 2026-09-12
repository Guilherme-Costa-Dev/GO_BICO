package domain

type UserBase struct {
	ID    string `json:"id" firestore:"-"`
	Nome  string `json:"nome" firestore:"nome"`
	Cpf   string `json:"cpf" firestore:"cpf"`
	Email string `json:"email" firestore:"email"`
	// senha é recebida a fim de salvar no serviço de auth.
	// provavelmente uma implementação de cadastro diretamente pelo frontend seria melhor, mas esta funcionando
	Senha string `json:"senha,omitempty" firestore:"-"`
}

type Cliente struct {
	UserBase
	Cep         string `json:"cep" firestore:"cep"`
	Numero      string `json:"numero" firestore:"numero"`
	Complemento string `json:"complemento" firestore:"complemento"`
}

type Prestador struct {
	UserBase
	Username         string   `json:"username" firestore:"username"`
	TiposServico     []string `json:"tiposServico" firestore:"tiposServico"`
	LocalAtuacao     string   `json:"localAtuacao" firestore:"localAtuacao"`
	FotoPerfil       string   `json:"fotoPerfil" firestore:"fotoPerfil"`
	FotoPaginaPerfil string   `json:"fotoPaginaPerfil" firestore:"fotoPaginaPerfil"`
	FotosServicos    []string `json:"fotosServicos" firestore:"fotosServicos"`
	Sobre            string   `json:"sobre" firestore:"sobre"`
}
