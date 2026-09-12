package main

import (
	"context"
	"log"
	"net/http"

	"bico/internal/config"
	"bico/internal/handler"
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

	http.HandleFunc("POST /cliente", clienteHandler.CreateCliente)
	http.HandleFunc("GET /cliente", clienteHandler.DadosCliente)
	http.HandleFunc("PUT /cliente", clienteHandler.AtualizarCliente)
	http.HandleFunc("DELETE /cliente", clienteHandler.DeletarCliente)

	http.HandleFunc("POST /prestador", prestadorHandler.CreatePrestador)
	http.HandleFunc("GET /prestador", prestadorHandler.DadosPrestador)
	http.HandleFunc("PUT /prestador", prestadorHandler.AtualizarPrestador)
	http.HandleFunc("DELETE /prestador", prestadorHandler.DeletarPrestador)
	http.HandleFunc("GET /prestadores", prestadorHandler.ListarPrestadores)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
