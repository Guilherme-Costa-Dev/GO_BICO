package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bico/internal/domain"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
)

type PrestadorHandler struct {
	db   *firestore.Client
	auth *auth.Client
}

func NewPrestadorHandler(db *firestore.Client, authClient *auth.Client) *PrestadorHandler {
	return &PrestadorHandler{
		db:   db,
		auth: authClient,
	}
}

func (h *PrestadorHandler) CreatePrestador(w http.ResponseWriter, r *http.Request) {
	var prestador domain.Prestador
	err := json.NewDecoder(r.Body).Decode(&prestador)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(prestador.Email).
		Password(prestador.Senha).
		DisplayName(prestador.Nome)

	userRecord, err := h.auth.CreateUser(r.Context(), params)
	if err != nil {
		http.Error(w, "Erro ao criar credenciais de autenticação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	prestador.ID = userRecord.UID
	if prestador.FotosServicos == nil {
		prestador.FotosServicos = []string{}
	}

	_, err = h.db.Collection("prestadores").Doc(userRecord.UID).Set(r.Context(), prestador)
	if err != nil {
		http.Error(w, "Erro ao salvar prestador", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prestador)
}

func (h *PrestadorHandler) DadosPrestador(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("id")
	if uid == "" {
		http.Error(w, "ID do usuário não fornecido na requisição", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	docPrestador, err := h.db.Collection("prestadores").Doc(uid).Get(r.Context())
	if err == nil && docPrestador.Exists() {
		var prestador domain.Prestador
		if err := docPrestador.DataTo(&prestador); err == nil {
			prestador.ID = uid
			json.NewEncoder(w).Encode(prestador)
			return
		}
	}

	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}

func (h *PrestadorHandler) AtualizarPrestador(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("id")
	if uid == "" {
		http.Error(w, "ID do usuário não fornecido", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	delete(updates, "id")

	docPrestador, _ := h.db.Collection("prestadores").Doc(uid).Get(r.Context())
	if docPrestador.Exists() {
		_, err := h.db.Collection("prestadores").Doc(uid).Set(r.Context(), updates, firestore.MergeAll)
		if err != nil {
			http.Error(w, "Erro ao atualizar prestador", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "Prestador atualizado com sucesso"}`))
		return
	}

	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}

func (h *PrestadorHandler) DeletarPrestador(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("id")
	if uid == "" {
		http.Error(w, "ID do usuário não fornecido", http.StatusBadRequest)
		return
	}

	err := h.auth.DeleteUser(r.Context(), uid)
	if err != nil {
		http.Error(w, "Erro ao deletar credenciais de autenticação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	docPrestador, _ := h.db.Collection("prestadores").Doc(uid).Get(r.Context())
	if docPrestador.Exists() {
		_, err := h.db.Collection("prestadores").Doc(uid).Delete(r.Context())
		if err != nil {
			http.Error(w, "Erro ao deletar prestador", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "Prestador deletado com sucesso"}`))
		return
	}

	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}

func (h *PrestadorHandler) ListarPrestadores(w http.ResponseWriter, r *http.Request) {
	tiposStr := r.URL.Query().Get("tipoServico")
	inicioStr := r.URL.Query().Get("inicio")
	fimStr := r.URL.Query().Get("fim")

	inicio := 0
	fim := 15
	if inicioStr != "" && fimStr != "" {
		var err error
		inicio, err = strconv.Atoi(inicioStr)
		if err != nil {
			http.Error(w, "Parâmetro 'inicio' inválido", http.StatusBadRequest)
			return
		}
		fim, err = strconv.Atoi(fimStr)
		if err != nil {
			http.Error(w, "Parâmetro 'fim' inválido", http.StatusBadRequest)
			return
		}
	}

	query := h.db.Collection("prestadores").Query

	if tiposStr != "" {
		tipos := strings.Split(tiposStr, ",")
		if len(tipos) > 10 {
			tipos = tipos[:10]
		}
		query = query.Where("tiposServico", "array-contains-any", tipos)
	}

	query = query.OrderBy("notaMedia", firestore.Desc)
	query = query.Offset(inicio).Limit(fim)

	docs, err := query.Documents(r.Context()).GetAll()
	if err != nil {
		http.Error(w, "Erro ao buscar prestadores: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var prestadores []domain.Prestador
	for _, doc := range docs {
		var p domain.Prestador
		if err := doc.DataTo(&p); err == nil {
			p.ID = doc.Ref.ID
			prestadores = append(prestadores, p)
		}
	}

	if prestadores == nil {
		prestadores = []domain.Prestador{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prestadores)
}
