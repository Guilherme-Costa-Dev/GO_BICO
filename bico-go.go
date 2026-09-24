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
	http.HandleFunc("GET /clientes", middleware.AuthMiddleware(authClient, clienteHandler.DadosCliente))
	http.HandleFunc("PUT /clientes", middleware.AuthMiddleware(authClient, clienteHandler.AtualizarCliente))
	http.HandleFunc("DELETE /clientes", middleware.AuthMiddleware(authClient, clienteHandler.DeletarCliente))

	http.HandleFunc("POST /prestadores", prestadorHandler.CreatePrestador)
	http.HandleFunc("GET /prestadores", middleware.AuthMiddleware(authClient, prestadorHandler.DadosPrestador))
	http.HandleFunc("PUT /prestadores", middleware.AuthMiddleware(authClient, prestadorHandler.AtualizarPrestador))	
	http.HandleFunc("DELETE /prestadores", middleware.AuthMiddleware(authClient, prestadorHandler.DeletarPrestador))
	http.HandleFunc("GET /prestadores/lista", prestadorHandler.ListarPrestadores)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
