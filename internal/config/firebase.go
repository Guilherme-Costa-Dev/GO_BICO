package config

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func InitFirestore(ctx context.Context, serviceAccountPath string) (*firestore.Client, error) {
	opt := option.WithCredentialsFile(serviceAccountPath)

	appConfig := &firebase.Config{
		ProjectID: "bico-23171",
	}

	app, err := firebase.NewApp(ctx, appConfig, opt)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar app do firebase: %w", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao firestore: %w", err)
	}

	log.Println("Conexão com Firebase Firestore realizada com sucesso!")
	return client, nil
}
