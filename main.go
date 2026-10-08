package main

import (
	"context"
	"log"
	"net/http"

	"bico/internal/config"
	"bico/internal/handler"
	"bico/internal/middleware"
)

func main() {
	ctx := context.Background()

	db, authClient, err := config.InitFirebase(ctx, "serviceAccountKey.json")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	clienteHandler := handler.NewClienteHandler(db, authClient)
	prestadorHandler := handler.NewPrestadorHandler(db, authClient)

	http.HandleFunc("POST /clientes", clienteHandler.CreateCliente)
	http.HandleFunc("GET /clientes", middleware.Auth(authClient, clienteHandler.DadosCliente))
	http.HandleFunc("PUT /clientes", middleware.Auth(authClient, clienteHandler.AtualizarCliente))
	http.HandleFunc("DELETE /clientes", middleware.Auth(authClient, clienteHandler.DeletarCliente))
	http.HandleFunc("PUT /clientes/favoritos", middleware.Auth(authClient, clienteHandler.AdicionarFavorito))
	http.HandleFunc("GET /clientes/favoritos", middleware.Auth(authClient, clienteHandler.ListarFavoritos))
	http.HandleFunc("DELETE /clientes/favoritos", middleware.Auth(authClient, clienteHandler.DeletarFavorito))

	http.HandleFunc("POST /prestadores", prestadorHandler.CreatePrestador)
	http.HandleFunc("GET /prestadores/busca", prestadorHandler.BuscarPrestadores)
	http.HandleFunc("GET /prestadores", middleware.Auth(authClient, prestadorHandler.DadosPrestador))
	http.HandleFunc("PUT /prestadores", middleware.Auth(authClient, prestadorHandler.AtualizarPrestador))
	http.HandleFunc("DELETE /prestadores", middleware.Auth(authClient, prestadorHandler.DeletarPrestador))

	handler := middleware.Cors(http.DefaultServeMux)

	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
