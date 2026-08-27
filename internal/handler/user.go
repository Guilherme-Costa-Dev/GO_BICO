package handler

import (
	"encoding/json"
	"net/http"

	"bico/internal/domain"

	"cloud.google.com/go/firestore"
)

type UserHandler struct {
	db *firestore.Client
}

func NewUserHandler(db *firestore.Client) *UserHandler {
	return &UserHandler{
		db: db,
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
