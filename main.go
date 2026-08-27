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

	db, err := config.InitFirestore(ctx, "serviceAccountKey.json")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	userHandler := handler.NewUserHandler(db)

	http.HandleFunc("/CadastrarCliente", userHandler.CreateCliente)

	log.Println("Servidor rodando em http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
