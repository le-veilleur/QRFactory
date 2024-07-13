// qr_handler.go

package handlers

import (
	"QRFactory/pkg/qr"
)

// QRHandler encapsule les dépendances nécessaires pour les opérations QR
type QRHandler struct {
	qrGenerator *qr.Generator // Instance du générateur QR
	// Ajoutez d'autres dépendances ici au besoin
}

// NewQRHandler initialise et retourne une nouvelle instance de QRHandler
func NewQRHandler(qrGenerator *qr.Generator) *QRHandler {
	return &QRHandler{
		qrGenerator: qrGenerator,
		// Initialisez d'autres dépendances ici au besoin
	}
}

// Implementez ici les méthodes de manipulation des QR codes, comme GenerateQRCodeAlphanumeric, GenerateQRCodeNumeric, etc.
