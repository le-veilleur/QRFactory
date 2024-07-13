package main

import (
	"fmt"
	"log"
	"net/http"

	"QRFactory/internal/api/handlers"
	"QRFactory/internal/api/routes"
)

func main() {
	// Configurer les routes de l'API
	router := routes.SetupRouter()

	// Initialiser les handlers de l'API
	qrHandler := handlers.NewQRHandler()

	// Définir les routes avec les handlers appropriés
	router.HandleFunc("/api/generateQR", qrHandler.GenerateQRCode).Methods("POST")

	// Définir le port d'écoute du serveur HTTP
	port := ":8080"
	fmt.Printf("Server listening on port %s\n", port)

	// Démarrer le serveur HTTP
	log.Fatal(http.ListenAndServe(port, router))
}
