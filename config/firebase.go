package config

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/storage"
	"google.golang.org/api/option"
)

var (
	FirebaseApp     *firebase.App
	FirebaseAuth    *auth.Client
	FirebaseStorage *storage.Client
)

// InitFirebase initializes Firebase with the service account key
func InitFirebase() error {
	// Get the service account key path from environment
	serviceAccountKeyPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY_PATH")
	if serviceAccountKeyPath == "" {
		serviceAccountKeyPath = "firebase-service-account.json"
	}

	// Check if service account key file exists
	if _, err := os.Stat(serviceAccountKeyPath); os.IsNotExist(err) {
		log.Printf("Warning: Firebase service account key not found at %s", serviceAccountKeyPath)
		log.Println("Firebase authentication will be disabled")
		return nil
	}

	// Initialize Firebase app
	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}
	FirebaseApp = app

	// Initialize Firebase Auth
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return err
	}
	FirebaseAuth = authClient

	// Initialize Firebase Storage
	storageClient, err := app.Storage(context.Background())
	if err != nil {
		return err
	}
	FirebaseStorage = storageClient

	log.Println("Firebase initialized successfully")
	return nil
}

// VerifyFirebaseToken verifies a Firebase ID token and returns the user ID
func VerifyFirebaseToken(idToken string) (string, error) {
	if FirebaseAuth == nil {
		return "", nil // Firebase not configured
	}

	token, err := FirebaseAuth.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return "", err
	}

	return token.UID, nil
}

// GetFirebaseStorageBucket returns the default storage bucket
func GetFirebaseStorageBucket() string {
	return os.Getenv("FIREBASE_STORAGE_BUCKET")
}
