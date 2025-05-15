package model

import (
	"image"
	"testing"
	"time"
)

func TestNewQRCode(t *testing.T) {
	data := "test data"
	version := 1
	errLevel := "M"

	qr := NewQRCode(data, version, errLevel)

	if qr.Data != data {
		t.Errorf("NewQRCode().Data = %v, veut %v", qr.Data, data)
	}

	if qr.Version != version {
		t.Errorf("NewQRCode().Version = %v, veut %v", qr.Version, version)
	}

	if qr.ErrorCorrectionLevel != errLevel {
		t.Errorf("NewQRCode().ErrorCorrectionLevel = %v, veut %v", qr.ErrorCorrectionLevel, errLevel)
	}

	expectedSize := version*4 + 17
	if qr.Size != expectedSize {
		t.Errorf("NewQRCode().Size = %v, veut %v", qr.Size, expectedSize)
	}

	// Vérifier que la date de création est à peu près correcte
	now := time.Now()
	if qr.CreatedAt.After(now) || qr.CreatedAt.Before(now.Add(-time.Second)) {
		t.Errorf("NewQRCode().CreatedAt n'est pas dans la plage attendue")
	}
}

func TestQRCode_SetMatrix(t *testing.T) {
	qr := NewQRCode("test", 1, "M")
	matrix := image.NewRGBA(image.Rect(0, 0, 21, 21))

	qr.SetMatrix(matrix)

	if qr.Matrix != matrix {
		t.Errorf("SetMatrix() n'a pas correctement défini la matrice")
	}
}

func TestQRCode_SetBitString(t *testing.T) {
	qr := NewQRCode("test", 1, "M")
	bitString := "101010"

	qr.SetBitString(bitString)

	if qr.BitString != bitString {
		t.Errorf("SetBitString() n'a pas correctement défini la chaîne de bits")
	}
}

func TestQRCode_SetMaskPattern(t *testing.T) {
	qr := NewQRCode("test", 1, "M")
	maskPattern := 3

	qr.SetMaskPattern(maskPattern)

	if qr.MaskPattern != maskPattern {
		t.Errorf("SetMaskPattern() n'a pas correctement défini le motif de masque")
	}
}
