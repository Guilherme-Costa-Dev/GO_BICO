package main

import (
	"context"
	"log"
	"net/http"

	"bico/internal/config"
	"bico/internal/server"
)

func main() {
	ctx := context.Background()

	db, authClient, err := config.InitFirebase(ctx, "serviceAccountKey.json")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	api := server.New(db, authClient)

	err = http.ListenAndServe(":8080", api)
	if err != nil {
		log.Fatal(err)
	}
}
