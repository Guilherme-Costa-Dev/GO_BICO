package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bico/internal/domain"
	"bico/internal/utils"

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

func (h *PrestadorHandler) CriarPrestador(w http.ResponseWriter, r *http.Request) {
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

	lat, lng, err := utils.buscarCoordPorCep(prestador.LocalAtuacao)
	if err != nil {
		http.Error(w, "Erro ao buscar coordenadas do CEP: "+err.Error(), http.StatusInternalServerError)
		return
	}
	prestador.Lat = lat
	prestador.Lng = lng

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
	uid := r.Context().Value("userUID").(string)

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
	uid := r.Context().Value("userUID").(string)

	var dto AtualizarPrestadorDTO
	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	docRef := h.db.Collection("prestadores").Doc(uid)
	docPrestador, err := docRef.Get(r.Context())
	if err != nil || !docPrestador.Exists() {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	_, err = docRef.Set(r.Context(), dto, firestore.MergeAll)
	if err != nil {
		http.Error(w, "Erro ao atualizar prestador", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Prestador atualizado com sucesso"}`))
}

func (h *PrestadorHandler) DeletarPrestador(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)

	_, err := h.db.Collection("prestadores").Doc(uid).Delete(r.Context())
	if err != nil {
		http.Error(w, "Erro ao deletar prestador no banco de dados", http.StatusInternalServerError)
		return
	}

	err = h.auth.DeleteUser(r.Context(), uid)
	if err != nil {
		http.Error(w, "Erro ao deletar credenciais de autenticação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Prestador deletado com sucesso"}`))
}

func (h *PrestadorHandler) BuscarPrestadores(w http.ResponseWriter, r *http.Request) {
	var dto BuscarPrestadorDTO
	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	inicio := 0
	if dto.Inicio != "" {
		v, err := strconv.Atoi(dto.Inicio)
		if err != nil || v < 0 {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		inicio = v
	}

	limite := 20
	if dto.Limite != "" {
		v, err := strconv.Atoi(dto.Limite)
		if err != nil || v < 0 || v > 100 {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		limite = v
	}

	var proximos bool = dto.Busca == "proximos"
	var latCliente, lngCliente, raio float64

	if proximos {
		if dto.IDCliente == "" {
			http.Error(w, "ID do cliente não fornecido", http.StatusBadRequest)
			return
		}

		docCliente, err := h.db.Collection("clientes").Doc(dto.IDCliente).Get(r.Context())
		if err != nil {
			http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
			return
		}
		var cliente domain.Cliente
		err := docCliente.DataTo(&cliente)
		if err != nil {
			http.Error(w, "Erro ao mapear dados do usuário", http.StatusInternalServerError)
			return
		}

		if cliente.Lat == 0 || cliente.Lng == 0 {
			http.Error(w, "Cliente sem localização cadastrada", http.StatusBadRequest)
			return
		}
		latCliente, lngCliente = cliente.Lat, cliente.Lng

		
		raio = 10
		if dto.Raio != "" {
			v, err := strconv.ParseFloat(dto.Raio, 64)
			if err != nil || v <= 0 {
				http.Error(w, "Raio inválido", http.StatusBadRequest)
				return
			}
			raio = v
		}
	}

	query := h.db.Collection("prestadores").Query

	if dto.Username != "" {
		query = query.Where("username", "==", dto.Username)
	}

	if dto.Tipo != "" {
		query = query.Where("tiposServico", "array-contains", dto.Tipo)
	}

	if !proximos {
		query = query.OrderBy("notaMedia", firestore.Desc)
		query = query.Offset(inicio).Limit(limite)
	}

	docs, err := query.Documents(r.Context()).GetAll()
	if err != nil {
		http.Error(w, "Erro ao buscar prestadores: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var prestadores []PrestadorPublicoDTO
	for _, doc := range docs {
		var p domain.Prestador
		err := doc.DataTo(&p)
		if err == nil {
			prestador := PrestadorPublicoDTO{
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

			if proximos {
				if p.Lat == 0 || p.Lng == 0 {
					continue
				}
				d := calcDistanciaPorCoord(latCliente, lngCliente, p.Lat, p.Lng)
				if d > raio {
					continue
				}
				prestador.DistanciaKm = d
			}

			prestadores = append(prestadores, prestador)
		}
	}

	if proximos {
		sort.Slice(prestadores, func(i, j int) bool {
			return prestadores[i].DistanciaKm < prestadores[j].DistanciaKm
		})

		if inicio > len(prestadores) {
			inicio = len(prestadores)
		}
		fim := inicio + limite
		if fim > len(prestadores) {
			fim = len(prestadores)
		}
		prestadores = prestadores[inicio:fim]
	}

	if prestadores == nil {
		prestadores = []PrestadorPublicoDTO{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prestadores)
}

func (h *PrestadorHandler) MudarEmail(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("userUID").(string)
	email := r.URL.Query().Get("email")

	params := (&auth.UserToUpdate{}).Email(email)

	if email == "" {
		http.Error(w, "E-mail não fornecido", http.StatusBadRequest)
		return
	}

	docCliente, err := h.db.Collection("prestadores").Doc(uid).Get(r.Context())
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

	_, err = h.db.Collection("prestadores").Doc(uid).Update(r.Context(), []firestore.Update{
		{
			Path:  "email",
			Value: email,
		},
	})
	if err != nil {
		http.Error(w, "Erro ao atualizar o email no firestore", http.StatusInternalServerError)

		_, err := h.auth.UpdateUser(ctx, uid, rollbackParams)
		if err != nil {
			http.Error(w, "Erro ao reverter o email do firebase auth", http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "Email alterado com sucesso"}`))
}