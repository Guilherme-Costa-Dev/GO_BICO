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

	//Inicia o bd e o auth de senha
	db, authClient, err := config.InitFirebase(ctx, "serviceAccountKey.json")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	//Passa os dois clientes para o handler
	userHandler := handler.NewUserHandler(db, authClient)

	http.HandleFunc("/CadastrarCliente", userHandler.CreateCliente)
	http.HandleFunc("/CadastrarPrestador", userHandler.CreatePrestador)
	http.HandleFunc("/GetDadosUsuario", userHandler.GetDadosUsuario)
	http.HandleFunc("/AtualizarUsuario", userHandler.AtualizarUsuario)

	log.Println("Servidor rodando em http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
