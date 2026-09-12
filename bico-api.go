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

	userHandler := handler.NewUserHandler(db, authClient)

	http.HandleFunc("POST /Cliente", userHandler.CreateCliente)
	http.HandleFunc("POST /Prestador", userHandler.CreatePrestador)
	http.HandleFunc("GET /Usuario", userHandler.GetDadosUsuario)
	http.HandleFunc("PUT /Usuario", userHandler.AtualizarUsuario)
	http.HandleFunc("DELETE /Usuario", userHandler.DeletarUsuario)

	log.Println("Servidor rodando em http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
