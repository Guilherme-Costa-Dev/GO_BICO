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

	// Cria a entrada no Firestore.
	_, err = h.db.Collection("clientes").Doc(userRecord.UID).Set(r.Context(), cliente)
	if err != nil {
		http.Error(w, "Erro ao salvar cliente", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(cliente)
}

func (h *UserHandler) CreatePrestador(w http.ResponseWriter, r *http.Request) {
	var prestador domain.Prestador

	// checa se nao tem erros no JSON recebido e traduz em struct
	err := json.NewDecoder(r.Body).Decode(&prestador)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// cria o usuario no firebase Authentication, separado do banco padrao, para seguranca da senha
	params := (&auth.UserToCreate{}).
		Email(prestador.Email).
		Password(prestador.Senha).
		DisplayName(prestador.Nome)

	userRecord, err := h.auth.CreateUser(r.Context(), params)
	if err != nil {
		http.Error(w, "Erro ao criar credenciais de autenticação", http.StatusInternalServerError)
		return
	}

	// o ID sera igual nos 2 "bancos"
	prestador.ID = userRecord.UID

	// Inicializa o slice de fotos de serviços como vazio (evita que fique como 'null' no JSON de resposta)
	if prestador.FotosServicos == nil {
		prestador.FotosServicos = []string{}
	}

	// Cria a entrada no Firestore.
	_, err = h.db.Collection("prestadores").Doc(userRecord.UID).Set(r.Context(), prestador)
	if err != nil {
		http.Error(w, "Erro ao salvar prestador", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Retorna o objeto criado (com os campos faltantes vazios, que serão atualizados depois no app)
	json.NewEncoder(w).Encode(prestador)
}

func (h *UserHandler) GetDadosUsuario(w http.ResponseWriter, r *http.Request) {
	// Pegamos o ID passado na URL (/DadosUsuario?id=UID_DO_USUARIO)
	uid := r.URL.Query().Get("id")
	if uid == "" {
		http.Error(w, "ID do usuário não fornecido na requisição", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	//Tenta buscar o documento na coleção de clientes
	docCliente, err := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if err == nil && docCliente.Exists() {
		var cliente domain.Cliente
		// Converte os dados do Firestore para a struct Cliente
		if err := docCliente.DataTo(&cliente); err == nil {
			json.NewEncoder(w).Encode(cliente)
			return
		}
	}

	//Se falhou ou não achou nos clientes, tenta buscar nos prestadores
	docPrestador, err := h.db.Collection("prestadores").Doc(uid).Get(r.Context())
	if err == nil && docPrestador.Exists() {
		var prestador domain.Prestador
		// Converte os dados do Firestore para a struct Prestador
		if err := docPrestador.DataTo(&prestador); err == nil {
			json.NewEncoder(w).Encode(prestador)
			return
		}
	}

	//Se não achou em nenhum dos dois
	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}

func (h *UserHandler) AtualizarUsuario(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("id")
	if uid == "" {
		http.Error(w, "ID do usuário não fornecido", http.StatusBadRequest)
		return
	}

	// Usamos map para capturar apenas os campos enviados no JSON
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	//Tenta atualizar se for cliente
	docCliente, _ := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if docCliente.Exists() {
		// Set com firestore.MergeAll atualiza apenas os campos passados no map
		_, err := h.db.Collection("clientes").Doc(uid).Set(r.Context(), updates, firestore.MergeAll)
		if err != nil {
			http.Error(w, "Erro ao atualizar cliente", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "Cliente atualizado com sucesso"}`))
		return
	}

	//Se não for cliente, tenta atualizar se for prestador
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

	//Se não achou em nenhum
	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}
