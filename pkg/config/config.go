package config

// QRConfig représente les configurations pour générer un QR code
type QRConfig struct {
	// Version du QR code (1-40), détermine la taille de la matrice
	Version int

	// Niveau de correction d'erreur (L: 7%, M: 15%, Q: 25%, H: 30%)
	ErrorCorrectionLevel string

	// Échelle pour l'image de sortie
	Scale int

	// Couleur de fond
	BackgroundColor string

	// Couleur des modules
	ForegroundColor string

	// Chemin du fichier de sortie
	OutputFile string

	// Données à encoder
	Data string
}

// NewDefaultConfig crée une nouvelle configuration avec des valeurs par défaut
func NewDefaultConfig() *QRConfig {
	return &QRConfig{
		Version:              1,
		ErrorCorrectionLevel: "M",
		Scale:                10,
		BackgroundColor:      "white",
		ForegroundColor:      "black",
		OutputFile:           "qrcode.png",
		Data:                 "",
	}
}

// ValidateConfig vérifie si la configuration est valide
func ValidateConfig(cfg *QRConfig) error {
	if cfg.Version < 1 || cfg.Version > 40 {
		return ErrInvalidVersion
	}

	if cfg.Data == "" {
		return ErrEmptyData
	}

	if cfg.ErrorCorrectionLevel != "L" && cfg.ErrorCorrectionLevel != "M" &&
		cfg.ErrorCorrectionLevel != "Q" && cfg.ErrorCorrectionLevel != "H" {
		return ErrInvalidErrorCorrectionLevel
	}

	return nil
}

// Erreurs standard pour la validation des configurations
var (
	ErrInvalidVersion              = NewError("version QR invalide, doit être entre 1 et 40")
	ErrEmptyData                   = NewError("les données ne peuvent pas être vides")
	ErrInvalidErrorCorrectionLevel = NewError("niveau de correction d'erreur invalide, doit être L, M, Q ou H")
)

// Error représente une erreur de configuration
type Error struct {
	Message string
}

// NewError crée une nouvelle erreur avec le message spécifié
func NewError(message string) *Error {
	return &Error{Message: message}
}

// Error implémente l'interface error
func (e *Error) Error() string {
	return e.Message
}
