package config

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

func InitFirebase(ctx context.Context, serviceAccountPath string) (*firestore.Client, *auth.Client, error) {
	opt := option.WithCredentialsFile(serviceAccountPath)

	appConfig := &firebase.Config{
		ProjectID: "bico-23171",
	}

	app, err := firebase.NewApp(ctx, appConfig, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao inicializar app do firebase: %w", err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao conectar ao firestore: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao conectar ao firebase auth: %w", err)
	}

	log.Println("Conexão com Firebase Firestore e Auth realizada com sucesso!")
	return firestoreClient, authClient, nil
}
