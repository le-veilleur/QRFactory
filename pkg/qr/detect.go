package qr

import (
	"strconv"
	"strings"
)

// detectDataType détermine le type de données à encoder
func DetectDataType(data string) string {
	// Vérifier si c'est une URL
	if strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://") {
		return "byte"
	}

	// Vérifier si c'est numérique
	if _, err := strconv.Atoi(data); err == nil {
		return "numeric"
	}

	// Vérifier si c'est alphanumérique
	if isAlphanumeric(data) {
		return "alphanumeric"
	}

	// Vérifier si c'est du Kanji
	if IsKanji(data) {
		return "kanji"
	}

	// Par défaut, utiliser le mode byte
	return "byte"
}

// isAlphanumeric vérifie si une chaîne est alphanumérique selon les spécifications QR Code
func isAlphanumeric(data string) bool {
	const validChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"
	data = strings.ToUpper(data)

	for _, char := range data {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}
	return true
}

// IsKanji vérifie si une chaîne contient uniquement des caractères Kanji valides
func IsKanji(data string) bool {
	for _, r := range data {
		if !isValidKanjiRange(r) {
			return false
		}
	}
	return true
}

// isValidKanjiRange vérifie si un caractère est dans une plage Kanji valide
func isValidKanjiRange(r rune) bool {
	ranges := [][2]rune{
		{0x4E00, 0x9FFF},   // CJK Unified Ideographs
		{0x3000, 0x303F},   // CJK Symbols and Punctuation
		{0xFF00, 0xFFEF},   // Halfwidth and Fullwidth Forms
		{0x30A0, 0x30FF},   // Katakana
		{0x3040, 0x309F},   // Hiragana
		{0x3400, 0x4DBF},   // Extension A
		{0x20000, 0x2A6DF}, // Extension B
		{0x2A700, 0x2B73F}, // Extension C
		{0x2B740, 0x2B81F}, // Extension D
		{0x2B820, 0x2CEAF}, // Extension E
		{0x2CEB0, 0x2EBEF}, // Extension F
		{0x2F800, 0x2FA1F}, // Compatibility Ideographs Supplement
	}

	for _, rang := range ranges {
		if r >= rang[0] && r <= rang[1] {
			return true
		}
	}
	return false
}
