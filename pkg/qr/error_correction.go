package qr

import (
	"fmt"
	"sync"
)

var (
	// Variables globales pour les tables de Galois
	gfExp = make([]byte, 256)
	gfLog = make([]byte, 256)

	// Mutex pour protéger l'accès aux tables de Galois
	gfMutex sync.RWMutex

	// Variable pour suivre si l'initialisation a été effectuée
	gfInitialized bool
)

// InitGaloisField initialise les tables de multiplication du champ de Galois de manière thread-safe
func InitGaloisField() {
	gfMutex.Lock()
	defer gfMutex.Unlock()

	// Vérifier si l'initialisation a déjà été effectuée
	if gfInitialized {
		return
	}

	fmt.Println("Initialisation des tables de Galois...")

	// Initialisation des tables de multiplication du champ de Galois
	// IMPORTANT: Cette initialisation évite d'appeler GfMultiply de manière récursive
	x := byte(1)
	for i := 0; i < 256; i++ {
		gfExp[i] = x

		// Remplir la table de logarithmes
		if i < 255 {
			gfLog[x] = byte(i)
		}

		// Multiplication par 2 dans GF(2^8)
		// Cette implémentation est l'équivalent direct de GfMultiply(x, 2)
		// mais sans appeler la fonction qui pourrait créer une récursion infinie
		if (x & 0x80) != 0 {
			x = (x << 1) ^ 0x1D // Polynôme de réduction: x^8 + x^4 + x^3 + x^2 + 1 (0x1D)
		} else {
			x = x << 1
		}
	}

	gfInitialized = true
	fmt.Println("Initialisation des tables de Galois terminée.")
}

// GfMultiply effectue une multiplication dans le champ de Galois
func GfMultiply(x, y byte) byte {
	// Cas spécial pour zéro
	if x == 0 || y == 0 {
		return 0
	}

	// Si les tables ne sont pas initialisées, on retourne une valeur sûre pour éviter la récursion
	if !gfInitialized {
		// Au lieu d'appeler InitGaloisField() qui pourrait créer une récursion infinie,
		// on retourne une valeur par défaut qui permet de continuer l'exécution
		fmt.Println("ATTENTION: Tentative d'utiliser GfMultiply avant initialisation")
		return 0
	}

	// Utiliser les tables initialisées pour effectuer la multiplication
	gfMutex.RLock()
	defer gfMutex.RUnlock()

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

// Polynômes générateurs pour différents niveaux de correction d'erreur
var GeneratorPolynomials = map[string][]int{
	"L": {1, 1},          // 7% de correction
	"M": {1, 1, 1},       // 15% de correction
	"Q": {1, 1, 1, 1},    // 25% de correction
	"H": {1, 1, 1, 1, 1}, // 30% de correction
}

// AddErrorCorrectionEC ajoute les codes de correction d'erreur aux données
// avec une implémentation plus robuste
func AddErrorCorrectionEC(data string, ecLevel string, version int) string {
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
