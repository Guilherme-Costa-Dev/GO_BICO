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
		http.Error(w, "Erro ao criar credenciais de autenticação: "+err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Erro ao criar credenciais de autenticação: "+err.Error(), http.StatusInternalServerError)
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
			cliente.ID = uid
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
			prestador.ID = uid
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

	// Remove o ID do mapa para não salvá-lo como um campo redundante dentro do documento
	delete(updates, "id")

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

func (h *UserHandler) DeletarUsuario(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("id")
	if uid == "" {
		http.Error(w, "ID do usuário não fornecido", http.StatusBadRequest)
		return
	}

	// Deleta o usuário do Firebase Authentication
	err := h.auth.DeleteUser(r.Context(), uid)
	if err != nil {
		http.Error(w, "Erro ao deletar credenciais de autenticação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Tenta deletar se for cliente
	docCliente, _ := h.db.Collection("clientes").Doc(uid).Get(r.Context())
	if docCliente.Exists() {
		_, err := h.db.Collection("clientes").Doc(uid).Delete(r.Context())
		if err != nil {
			http.Error(w, "Erro ao deletar cliente", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "Cliente deletado com sucesso"}`))
		return
	}

	// Se não for cliente, tenta deletar se for prestador
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

	// Se não achou em nenhum
	http.Error(w, "Usuário não encontrado", http.StatusNotFound)
}

func (h *UserHandler) ListarPrestadores(w http.ResponseWriter, r *http.Request) {
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

		// firebase tem um limite de apenas 10
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

