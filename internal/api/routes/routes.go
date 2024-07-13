// routes.go

package routes

import (
	"github.com/gin-gonic/gin"
	"QRFactory/internal/api/handlers"
	"QRFactory/pkg/qr" // Assurez-vous d'importer votre package qr ici
)

// SetupRouter configure les routes de l'application
func SetupRouter(qrGenerator *qr.Generator) *gin.Engine {
	r := gin.Default()

	qrHandler := handlers.NewQRHandler(qrGenerator)

	// Endpoint pour générer un QR code à partir de données alphanumériques
	r.GET("/api/qr/alphanumeric", qrHandler.GenerateQRCodeAlphanumeric)

	// Endpoint pour générer un QR code à partir de données numériques
	r.GET("/api/qr/numeric", qrHandler.GenerateQRCodeNumeric)

	// Endpoint pour générer un QR code à partir de données Kanji
	r.GET("/api/qr/kanji", qrHandler.GenerateQRCodeKanji)

	return r
}
