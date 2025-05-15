package config

import "testing"

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.Version != 1 {
		t.Errorf("Version par défaut incorrecte, obtenu: %d, attendu: %d", cfg.Version, 1)
	}

	if cfg.ErrorCorrectionLevel != "M" {
		t.Errorf("Niveau de correction par défaut incorrect, obtenu: %s, attendu: %s", cfg.ErrorCorrectionLevel, "M")
	}

	if cfg.Scale != 10 {
		t.Errorf("Échelle par défaut incorrecte, obtenu: %d, attendu: %d", cfg.Scale, 10)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *QRConfig
		wantErr error
	}{
		{
			name: "configuration valide",
			config: &QRConfig{
				Version:              1,
				ErrorCorrectionLevel: "M",
				Data:                 "test",
			},
			wantErr: nil,
		},
		{
			name: "version invalide (trop basse)",
			config: &QRConfig{
				Version:              0,
				ErrorCorrectionLevel: "M",
				Data:                 "test",
			},
			wantErr: ErrInvalidVersion,
		},
		{
			name: "version invalide (trop haute)",
			config: &QRConfig{
				Version:              41,
				ErrorCorrectionLevel: "M",
				Data:                 "test",
			},
			wantErr: ErrInvalidVersion,
		},
		{
			name: "données vides",
			config: &QRConfig{
				Version:              1,
				ErrorCorrectionLevel: "M",
				Data:                 "",
			},
			wantErr: ErrEmptyData,
		},
		{
			name: "niveau de correction invalide",
			config: &QRConfig{
				Version:              1,
				ErrorCorrectionLevel: "X",
				Data:                 "test",
			},
			wantErr: ErrInvalidErrorCorrectionLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if err != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	message := "test error message"
	err := NewError(message)

	if err.Error() != message {
		t.Errorf("Error.Error() = %v, want %v", err.Error(), message)
	}
}
