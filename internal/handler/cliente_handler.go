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

type FavoritoRequest struct {
	ID string `json:"id"`
}

func (h *ClienteHandler) AdicionarFavorito(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)

	var req FavoritoRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.ID == "" {
		http.Error(w, "ID nao fornecido", http.StatusBadRequest)
		return
	}

	docCliente, err := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if err != nil {
		http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		return
	}
	if docCliente.Exists() {
		_, err := h.db.Collection("clientes").Doc(uid).Update(r.Context(), []firestore.Update{
			{
				Path:  "favoritos",
				Value: firestore.ArrayUnion(req.ID),
			},
		})
		if err != nil {
			http.Error(w, "Erro ao adicionar favorito: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Favorito adicionado com sucesso"}`))
}

func (h *ClienteHandler) ListarFavoritos(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)

	docCliente, err := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if err != nil {
		http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
		return
	}

	var cliente domain.Cliente
	if err := docCliente.DataTo(&cliente); err != nil {
		http.Error(w, "Erro ao mapear dados do usuário", http.StatusInternalServerError)
		return
	}

	prestadores := []PrestadorPublico{}

	if len(cliente.Favoritos) > 0 {
		var docRefs []*firestore.DocumentRef
		for _, id := range cliente.Favoritos {
			docRefs = append(docRefs, h.db.Collection("prestadores").Doc(id))
		}

		docs, err := h.db.GetAll(r.Context(), docRefs)
		if err != nil {
			http.Error(w, "Erro ao buscar dados dos prestadores", http.StatusInternalServerError)
			return
		}

		for _, doc := range docs {
			if !doc.Exists() {
				continue
			}

			var p domain.Prestador
			err := doc.DataTo(&p)
			if err == nil {
				prestador := PrestadorPublico{
					ID:               doc.Ref.ID,
					Nome:             p.Nome,
					TiposServico:     p.TiposServico,
					NotaMedia:        p.NotaMedia,
					Email:            p.Email,
					LocalAtuacao:     p.LocalAtuacao,
					Username:         p.Username,
					FotoPerfil:       p.FotoPerfil,
					FotoPaginaPerfil: p.FotoPaginaPerfil,
					FotosServicos:    p.FotosServicos,
					Sobre:            p.Sobre,
					TotalAvaliacoes:  p.TotalAvaliacoes,
				}
				prestadores = append(prestadores, prestador)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"favoritos": prestadores,
	})
}

func (h *ClienteHandler) DeletarFavorito(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)

	var req FavoritoRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.ID == "" {
		http.Error(w, "ID nao fornecido", http.StatusBadRequest)
		return
	}

	_, err = h.db.Collection("clientes").Doc(uid).Update(r.Context(), []firestore.Update{
		{
			Path:  "favoritos",
			Value: firestore.ArrayRemove(req.ID),
		},
	})

	if err != nil {
		http.Error(w, "Erro ao remover favorito: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Favorito removido com sucesso"}`))
}
