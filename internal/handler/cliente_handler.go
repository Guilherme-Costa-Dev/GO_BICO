package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bico/internal/domain"
	"bico/internal/dto"
	"bico/internal/utils"

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

func (h *ClienteHandler) CriarCliente(w http.ResponseWriter, r *http.Request) {
	var data dto.CriarClienteDTO
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	var faltando bool
	if strings.TrimSpace(data.Nome) == "" {
		faltando = true
	}
	if strings.TrimSpace(data.Cpf) == "" {
		faltando = true
	}
	if strings.TrimSpace(data.Cep) == "" {
		faltando = true
	}
	if strings.TrimSpace(data.Email) == "" {
		faltando = true
	}
	if strings.TrimSpace(data.Senha) == "" {
		faltando = true
	}
	if faltando {
		http.Error(w, "JSON faltando campos obrigatórios", http.StatusBadRequest)
		return
	}

	lat, lng, err := utils.BuscarCoordPorCep(data.Cep)
	if err != nil {
		http.Error(w, "Erro ao buscar coordenadas do CEP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(data.Email).
		Password(data.Senha).
		DisplayName(data.Nome)
	userRecord, err := h.auth.CreateUser(r.Context(), params)
	if err != nil {
		http.Error(w, "Erro ao criar credenciais de autenticação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cliente := domain.Cliente{
		UserBase: domain.UserBase{
			ID:    userRecord.UID,
			Nome:  data.Nome,
			Cpf:   data.Cpf,
			Cep:   data.Cep,
			Lat:   lat,
			Lng:   lng,
			Email: data.Email,
		},
		Numero:      data.Numero,
		Complemento: data.Complemento,
		Favoritos:   []string{},
	}

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

	var data dto.AtualizarClienteDTO
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var faltando bool
	if strings.TrimSpace(data.Nome) == "" {
		faltando = true
	}
	if strings.TrimSpace(data.Cpf) == "" {
		faltando = true
	}
	if strings.TrimSpace(data.Cep) == "" {
		faltando = true
	}
	if faltando {
		http.Error(w, "JSON faltando campos obrigatórios", http.StatusBadRequest)
		return
	}

	docRef := h.db.Collection("clientes").Doc(uid)
	docCliente, err := docRef.Get(r.Context())
	if err != nil || !docCliente.Exists() {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	updates := []firestore.Update{
		{Path: "nome", Value: data.Nome},
		{Path: "cpf", Value: data.Cpf},
		{Path: "cep", Value: data.Cep},
		{Path: "numero", Value: data.Numero},
		{Path: "complemento", Value: data.Complemento},
	}

	if data.Cep != "" {
		lat, lng, err := utils.BuscarCoordPorCep(data.Cep)
		if err != nil {
			http.Error(w, "Erro ao buscar coordenadas do CEP: "+err.Error(), http.StatusInternalServerError)
			return
		}
		updates = append(updates,
			firestore.Update{Path: "lat", Value: lat},
			firestore.Update{Path: "lng", Value: lng})
	}

	_, err = docRef.Update(r.Context(), updates)
	if err != nil {
		http.Error(w, "Erro ao atualizar cliente", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Cliente atualizado com sucesso"}`))
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

func (h *ClienteHandler) AdicionarFavorito(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)
	prestador_ID := r.URL.Query().Get("id")

	if prestador_ID == "" {
		http.Error(w, "ID do prestador nao fornecido", http.StatusBadRequest)
		return
	}

	_, err := h.db.Collection("clientes").Doc(uid).Update(r.Context(), []firestore.Update{
		{
			Path:  "favoritos",
			Value: firestore.ArrayUnion(prestador_ID),
		},
	})

	if err != nil {
		http.Error(w, "Erro ao adicionar favorito: "+err.Error(), http.StatusInternalServerError)
		return
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

	prestadores := []dto.PrestadorPublicoDTO{}

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
				prestador := dto.PrestadorPublicoDTO{
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
					TotalAvaliacoes:  len(p.Avaliacoes),
					Avaliacoes:       p.Avaliacoes,
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
	ID := r.URL.Query().Get("id")

	if ID == "" {
		http.Error(w, "ID nao fornecido", http.StatusBadRequest)
		return
	}

	_, err := h.db.Collection("clientes").Doc(uid).Update(r.Context(), []firestore.Update{
		{
			Path:  "favoritos",
			Value: firestore.ArrayRemove(ID),
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

func (h *ClienteHandler) AdicionarAvaliacao(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)
	var dto dto.AdicionarAvaliacaoDTO

	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		http.Error(w, "Erro ao processar o JSON recebido: "+err.Error(), http.StatusBadRequest)
		return
	}

	if dto.PrestadorID == "" {
		http.Error(w, "ID ou nota do prestador nao fornecido", http.StatusBadRequest)
		return
	}

	if dto.Nota < 0 || dto.Nota > 5 {
		http.Error(w, "Nota invalida", http.StatusBadRequest)
		return
	}

	docCliente, err := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if err != nil {
		http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
		return
	}

	var cliente domain.Cliente
	err = docCliente.DataTo(&cliente)
	if err != nil {
		http.Error(w, "Erro ao mapear dados do usuário", http.StatusInternalServerError)
		return
	}

	avaliacao := domain.Avaliacao{
		Nome:       cliente.Nome,
		Nota:       dto.Nota,
		Comentario: dto.Comentario,
		Data:       time.Now(),
	}

	docPrestador, err := h.db.Collection("prestadores").Doc(dto.PrestadorID).Get(r.Context())
	if err != nil {
		http.Error(w, "Erro ao buscar dados do prestador", http.StatusInternalServerError)
		return
	}
	var prestador domain.Prestador
	err = docPrestador.DataTo(&prestador)
	if err != nil {
		http.Error(w, "Erro ao mapear dados do prestador", http.StatusInternalServerError)
		return
	}

	lenAvaliacoes := float64(len(prestador.Avaliacoes))
	notaMedia := (float64(dto.Nota) + prestador.NotaMedia*lenAvaliacoes) / (lenAvaliacoes + 1.0)
	_, err = h.db.Collection("prestadores").Doc(dto.PrestadorID).Update(r.Context(), []firestore.Update{
		{
			Path:  "avaliacoes",
			Value: firestore.ArrayUnion(avaliacao),
		},
		{
			Path:  "notaMedia",
			Value: notaMedia,
		},
	})
	if err != nil {
		http.Error(w, "Erro ao adicionar avaliacao ao prestador", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Avaliacao adicionada com sucesso"}`))
}

func (h *ClienteHandler) MudarEmail(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)
	email := r.URL.Query().Get("email")

	params := (&auth.UserToUpdate{}).Email(email)

	if email == "" {
		http.Error(w, "E-mail não fornecido", http.StatusBadRequest)
		return
	}

	docCliente, err := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if err != nil {
		http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
		return
	}
	var cliente domain.Cliente
	err = docCliente.DataTo(&cliente)
	if err != nil {
		http.Error(w, "Erro ao mapear dados do usuário", http.StatusInternalServerError)
		return
	}
	rollbackParams := (&auth.UserToUpdate{}).Email(cliente.Email)

	ctx := r.Context()
	_, err = h.auth.UpdateUser(ctx, uid, params)
	if err != nil {
		http.Error(w, "Erro ao atualizar o email no firebase auth", http.StatusInternalServerError)
		return
	}

	_, err = h.db.Collection("clientes").Doc(uid).Update(r.Context(), []firestore.Update{
		{
			Path:  "email",
			Value: email,
		},
	})
	if err != nil {
		_, rollbackErr := h.auth.UpdateUser(ctx, uid, rollbackParams)
		if rollbackErr != nil {
			http.Error(w, "Erro ao atualizar o email no firestore e ao reverter no firebase auth", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Erro ao atualizar o email no firestore", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Email alterado com sucesso"}`))
}
