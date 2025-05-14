package qr

import (
	"fmt"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// EncodeNumeric encode une chaîne numérique en binaire selon les spécifications du QR code
func EncodeNumeric(data string) (string, error) {
	var result strings.Builder

	// Traiter les groupes de 3 chiffres
	for i := 0; i < len(data); i += 3 {
		end := min(i+3, len(data))
		group := data[i:end]

		// Conserver les zéros au début pour les groupes non complets
		if len(group) < 3 && i > 0 {
			group = strings.Repeat("0", 3-len(group)) + group
		}

		value, err := ToInt(group)
		if err != nil {
			return "", err
		}

		// Déterminer le nombre de bits nécessaires
		var bits int
		switch len(group) {
		case 3:
			bits = 10 // 3 chiffres = 10 bits (max 999)
		case 2:
			bits = 7 // 2 chiffres = 7 bits (max 99)
		case 1:
			bits = 4 // 1 chiffre = 4 bits (max 9)
		}

		// Formater avec le bon nombre de bits
		format := fmt.Sprintf("%%0%db", bits)
		result.WriteString(fmt.Sprintf(format, value))
	}

	return result.String(), nil
}

// EncodeAlphanumeric encode une chaîne alphanumérique en binaire selon les spécifications du QR code
func EncodeAlphanumeric(data string) (string, error) {
	var result strings.Builder
	alphanumericTable := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"

	for i := 0; i < len(data); i += 2 {
		if i+1 < len(data) {
			val := strings.IndexByte(alphanumericTable, data[i])*45 + strings.IndexByte(alphanumericTable, data[i+1])
			result.WriteString(fmt.Sprintf("%011b", val))
		} else {
			val := strings.IndexByte(alphanumericTable, data[i])
			result.WriteString(fmt.Sprintf("%06b", val))
		}
	}
	return result.String(), nil
}

// EncodeByte encode une chaîne de caractères en binaire selon les spécifications du QR code
func EncodeByte(data string) (string, error) {
	var result strings.Builder
	bytes := []byte(data)

	for _, b := range bytes {
		result.WriteString(fmt.Sprintf("%08b", b))
	}
	return result.String(), nil
}

// EncodeKanji encode une chaîne de caractères Kanji en binaire selon les spécifications du QR code
func EncodeKanji(data string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("la chaîne Kanji ne peut pas être vide")
	}

	var result strings.Builder

	// Encoder la chaîne en Shift JIS
	encoder := japanese.ShiftJIS.NewEncoder()
	sjisBytes, _, err := transform.Bytes(encoder, []byte(data))
	if err != nil {
		return "", fmt.Errorf("erreur lors de la conversion en Shift JIS: %v", err)
	}

	if len(sjisBytes)%2 != 0 {
		return "", fmt.Errorf("données Kanji invalides: nombre impair d'octets")
	}

	for i := 0; i < len(sjisBytes); i += 2 {
		msb := uint16(sjisBytes[i])
		lsb := uint16(sjisBytes[i+1])

		// Vérification des plages valides pour Shift JIS
		if !((msb >= 0x81 && msb <= 0x9F) || (msb >= 0xE0 && msb <= 0xEA)) {
			return "", fmt.Errorf("caractère Kanji invalide: premier octet 0x%02X hors plage", msb)
		}
		if !((lsb >= 0x40 && lsb <= 0x7E) || (lsb >= 0x80 && lsb <= 0xFC)) {
			return "", fmt.Errorf("caractère Kanji invalide: second octet 0x%02X hors plage", lsb)
		}

		// Former le mot de 16 bits
		word := (msb << 8) | lsb

		// Soustraire l'offset selon la plage
		var adjusted uint16
		if msb >= 0x81 && msb <= 0x9F {
			adjusted = word - 0x8140
		} else {
			adjusted = word - 0xC140
		}

		// Extraire MSB et LSB ajustés
		adjustedMsb := adjusted >> 8
		adjustedLsb := adjusted & 0xFF

		// Calcul final selon la spécification QR Code
		value := adjustedMsb*0xC0 + adjustedLsb

		// Convertir en binaire sur 13 bits en utilisant des opérations bit à bit
		var bits strings.Builder
		for j := 12; j >= 0; j-- {
			if (value & (1 << uint(j))) != 0 {
				bits.WriteByte('1')
			} else {
				bits.WriteByte('0')
			}
		}
		result.WriteString(bits.String())
	}

	return result.String(), nil
}
