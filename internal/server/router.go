package server

import (
	"net/http"

	"bico/internal/handler"
	"bico/internal/middleware"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
)

func New(db *firestore.Client, authClient *auth.Client) http.Handler {
	clienteHandler := handler.NewClienteHandler(db, authClient)
	prestadorHandler := handler.NewPrestadorHandler(db, authClient)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /clientes", clienteHandler.CriarCliente)
	mux.HandleFunc("GET /clientes", middleware.Auth(authClient, clienteHandler.DadosCliente))
	mux.HandleFunc("PUT /clientes", middleware.Auth(authClient, clienteHandler.AtualizarCliente))
	mux.HandleFunc("DELETE /clientes", middleware.Auth(authClient, clienteHandler.DeletarCliente))
	mux.HandleFunc("PUT /clientes/favoritos", middleware.Auth(authClient, clienteHandler.AdicionarFavorito))
	mux.HandleFunc("GET /clientes/favoritos", middleware.Auth(authClient, clienteHandler.ListarFavoritos))
	mux.HandleFunc("DELETE /clientes/favoritos", middleware.Auth(authClient, clienteHandler.DeletarFavorito))
	mux.HandleFunc("PUT /clientes/avaliacoes", middleware.Auth(authClient, clienteHandler.AdicionarAvaliacao))
	mux.HandleFunc("PUT /clientes/email", middleware.Auth(authClient, clienteHandler.MudarEmail))

	mux.HandleFunc("POST /prestadores", prestadorHandler.CriarPrestador)
	mux.HandleFunc("GET /prestadores/busca", middleware.Auth(authClient, prestadorHandler.BuscarPrestadores))
	mux.HandleFunc("GET /prestadores", middleware.Auth(authClient, prestadorHandler.DadosPrestador))
	mux.HandleFunc("PUT /prestadores", middleware.Auth(authClient, prestadorHandler.AtualizarPrestador))
	mux.HandleFunc("DELETE /prestadores", middleware.Auth(authClient, prestadorHandler.DeletarPrestador))
	mux.HandleFunc("PUT /prestadores/email", middleware.Auth(authClient, prestadorHandler.MudarEmail))

	return middleware.Cors(mux)
}
