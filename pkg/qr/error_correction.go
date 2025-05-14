package qr

import (
	"fmt"
)

// InitGaloisField initialise les tables de multiplication du champ de Galois
func InitGaloisField() {
	// Initialisation des tables de multiplication du champ de Galois
	gfExp := make([]byte, 256)
	gfLog := make([]byte, 256)

	x := byte(1)
	for i := 0; i < 256; i++ {
		gfExp[i] = x
		if i < 255 {
			gfLog[x] = byte(i)
		}
		x = GfMultiply(x, 2)
	}
}

// GfMultiply effectue une multiplication dans le champ de Galois
func GfMultiply(x, y byte) byte {
	if x == 0 || y == 0 {
		return 0
	}
	return gfExp[(gfLog[x]+gfLog[y])%255]
}

// GenerateReedSolomon génère les octets de correction d'erreur Reed-Solomon
func GenerateReedSolomon(data []byte, numECBytes int) []byte {
	generator := generateGenerator(numECBytes)
	remainder := make([]byte, numECBytes)

	for _, d := range data {
		factor := d ^ remainder[0]
		copy(remainder, remainder[1:])
		remainder[len(remainder)-1] = 0

		for i := 0; i < len(remainder); i++ {
			remainder[i] ^= GfMultiply(generator[i], factor)
		}
	}

	return remainder
}

// GenerateErrorCorrection génère les codes de correction d'erreur pour les données
func GenerateErrorCorrection(data []byte, level string) []byte {
	var numECBytes int
	switch level {
	case "L":
		numECBytes = 7
	case "M":
		numECBytes = 10
	case "Q":
		numECBytes = 13
	case "H":
		numECBytes = 17
	default:
		numECBytes = 10 // Par défaut niveau M
	}

	return GenerateReedSolomon(data, numECBytes)
}

// generateGenerator génère le polynôme générateur pour Reed-Solomon
func generateGenerator(degree int) []byte {
	generator := make([]byte, degree)
	generator[0] = 1

	for i := 0; i < degree; i++ {
		for j := i; j >= 0; j-- {
			generator[j] = GfMultiply(generator[j], byte(i+1))
			if j > 0 {
				generator[j] ^= generator[j-1]
			}
		}
	}

	return generator
}

// Variables globales pour les tables de Galois
var (
	gfExp = make([]byte, 256)
	gfLog = make([]byte, 256)
)

// Polynômes générateurs pour différents niveaux de correction d'erreur
var GeneratorPolynomials = map[string][]int{
	"L": {1, 1},          // 7% de correction
	"M": {1, 1, 1},       // 15% de correction
	"Q": {1, 1, 1, 1},    // 25% de correction
	"H": {1, 1, 1, 1, 1}, // 30% de correction
}

// AddErrorCorrection ajoute les codes de correction d'erreur aux données
func AddErrorCorrection(data string, ecLevel string, version int) string {
	// Calculer le nombre de mots de code de correction d'erreur nécessaires
	ecWords := calculateECWords(version, ecLevel)

	// Convertir les données binaires en bytes
	dataBytes := make([]byte, 0)
	for i := 0; i < len(data); i += 8 {
		end := i + 8
		if end > len(data) {
			end = len(data)
		}
		byteVal := binaryStringToByte(data[i:end])
		dataBytes = append(dataBytes, byteVal)
	}

	// Générer les mots de code de correction d'erreur
	ecBytes := generateECBytes(dataBytes, ecWords)

	// Convertir les bytes de correction d'erreur en binaire
	var result string
	result = data
	for _, b := range ecBytes {
		result += fmt.Sprintf("%08b", b)
	}

	return result
}

// Convertit une chaîne binaire en byte
func binaryStringToByte(binary string) byte {
	var result byte
	for i := 0; i < len(binary); i++ {
		if binary[i] == '1' {
			result |= 1 << uint(7-i)
		}
	}
	return result
}

// Calcule le nombre de mots de code de correction d'erreur nécessaires
func calculateECWords(version int, ecLevel string) int {
	// Table simplifiée pour les versions 1-5
	ecWordsTable := map[int]map[string]int{
		1: {"L": 7, "M": 10, "Q": 13, "H": 17},
		2: {"L": 10, "M": 16, "Q": 22, "H": 28},
		3: {"L": 15, "M": 26, "Q": 36, "H": 44},
		4: {"L": 20, "M": 36, "Q": 52, "H": 64},
		5: {"L": 26, "M": 48, "Q": 72, "H": 88},
	}

	if words, ok := ecWordsTable[version][ecLevel]; ok {
		return words
	}
	return 10 // Valeur par défaut pour version 1-M
}

// Génère les bytes de correction d'erreur
func generateECBytes(data []byte, numECBytes int) []byte {
	// Initialiser le tableau de correction d'erreur
	ecBytes := make([]byte, numECBytes)

	// Copier les données dans le tableau de correction d'erreur
	for i := 0; i < len(data); i++ {
		feedback := data[i]
		for j := 0; j < len(ecBytes); j++ {
			temp := ecBytes[j]
			ecBytes[j] = byte(int(feedback) ^ int(temp))
			feedback = temp
		}
	}

	return ecBytes
}
