package model

import (
	"image"
	"time"
)

// QRCode représente un code QR généré
type QRCode struct {
	// Identifiant unique du QR code
	ID string

	// Version du QR code (1-40)
	Version int

	// Niveau de correction d'erreur (L, M, Q, H)
	ErrorCorrectionLevel string

	// Données encodées dans le QR code
	Data string

	// Chaîne de bits représentant le QR code
	BitString string

	// Matrice d'image du QR code
	Matrix *image.RGBA

	// Taille du QR code en modules
	Size int

	// Date de création
	CreatedAt time.Time

	// Masque utilisé (0-7)
	MaskPattern int
}

// NewQRCode crée une nouvelle instance de QRCode avec les données spécifiées
func NewQRCode(data string, version int, errLevel string) *QRCode {
	return &QRCode{
		Data:                 data,
		Version:              version,
		ErrorCorrectionLevel: errLevel,
		CreatedAt:            time.Now(),
		Size:                 version*4 + 17, // Taille de la matrice = version*4 + 17
	}
}

// SetMatrix définit la matrice d'image du QR code
func (qr *QRCode) SetMatrix(matrix *image.RGBA) {
	qr.Matrix = matrix
}

// SetBitString définit la chaîne de bits du QR code
func (qr *QRCode) SetBitString(bitString string) {
	qr.BitString = bitString
}

// SetMaskPattern définit le motif de masque utilisé
func (qr *QRCode) SetMaskPattern(maskPattern int) {
	qr.MaskPattern = maskPattern
}
