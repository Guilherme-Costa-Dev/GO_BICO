package handler

import (
	"encoding/json"
	"net/http"

	"bico/internal/domain"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
)

type ClienteHandler struct {
	db   *firestore.Client
	auth *auth.Client
}

func NewClienteHandler(db *firestore.Client, authClient *auth.Client) *ClienteHandler {
	return &ClienteHandler{
		db:   db,
		auth: authClient,
	}
}

func (h *ClienteHandler) CreateCliente(w http.ResponseWriter, r *http.Request) {
	var cliente domain.Cliente
	err := json.NewDecoder(r.Body).Decode(&cliente)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(cliente.Email).
		Password(cliente.Senha).
		DisplayName(cliente.Nome)
	userRecord, err := h.auth.CreateUser(r.Context(), params)
	if err != nil {
		http.Error(w, "Erro ao criar credenciais de autenticação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cliente.ID = userRecord.UID
	_, err = h.db.Collection("clientes").Doc(userRecord.UID).Set(r.Context(), cliente)
	if err != nil {
		http.Error(w, "Erro ao salvar cliente", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cliente)
}

func (h *ClienteHandler) DadosCliente(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)

	w.Header().Set("Content-Type", "application/json")
	docCliente, err := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if err == nil && docCliente.Exists() {
		var cliente domain.Cliente
		if err := docCliente.DataTo(&cliente); err == nil {
			cliente.ID = uid
			json.NewEncoder(w).Encode(cliente)
			return
		}
	}

	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}

func (h *ClienteHandler) AtualizarCliente(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	delete(updates, "id")

	docCliente, _ := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if docCliente.Exists() {
		_, err := h.db.Collection("clientes").Doc(uid).Set(r.Context(), updates, firestore.MergeAll)
		if err != nil {
			http.Error(w, "Erro ao atualizar cliente", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "Cliente atualizado com sucesso"}`))
		return
	}

	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}

func (h *ClienteHandler) DeletarCliente(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)

	_, err := h.db.Collection("clientes").Doc(uid).Delete(r.Context())
	if err != nil {
		http.Error(w, "Erro ao deletar cliente no banco de dados", http.StatusInternalServerError)
		return
	}

	err = h.auth.DeleteUser(r.Context(), uid)
	if err != nil {
		http.Error(w, "Erro ao deletar credenciais de autenticação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Cliente deletado com sucesso"}`))
}
