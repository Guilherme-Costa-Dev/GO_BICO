package handler

import (
	"encoding/json"
	"net/http"

	"bico/internal/domain"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
)

type UserHandler struct {
	db   *firestore.Client
	auth *auth.Client
}

func NewUserHandler(db *firestore.Client, authClient *auth.Client) *UserHandler {
	return &UserHandler{
		db:   db,
		auth: authClient,
	}
}

// funcao pra criar novo usuario do tipo cliente
func (h *UserHandler) CreateCliente(w http.ResponseWriter, r *http.Request) {
	var cliente domain.Cliente

	//checa se nao tem erros no JSON recebido e traduz em struct
	err := json.NewDecoder(r.Body).Decode(&cliente)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	//cria o usuario no firebase Authentication, separado do banco padrao, para seguranca da senha
	params := (&auth.UserToCreate{}).
		Email(cliente.Email).
		Password(cliente.Senha).
		DisplayName(cliente.Nome)
	userRecord, err := h.auth.CreateUser(r.Context(), params)
	if err != nil {
		http.Error(w, "Erro ao criar credenciais de autenticação", http.StatusInternalServerError)
		return
	}

	//o ID sera igual nos 2 "bancos"
	cliente.ID = userRecord.UID

	//checa se nao deu erro no salvamento do cliente e cria a entrada no bd
	doc, _, err := h.db.Collection("clientes").Add(r.Context(), cliente)
	if err != nil {
		http.Error(w, "Erro ao salvar usuario", http.StatusInternalServerError)
		return
	}

	cliente.ID = doc.ID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(cliente)
}
