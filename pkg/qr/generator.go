package qr

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Caractéristiques des différentes versions de QR code
var versionInfo = map[int]struct {
	Size    int // Taille de la matrice
	Modules int // Nombre de modules
	// Ajoutez d'autres informations spécifiques à chaque version si nécessaire
}{
	1:  {21, 21},
	2:  {25, 25},
	3:  {29, 29},
	4:  {33, 33},
	5:  {37, 37},
	6:  {41, 41},
	7:  {45, 45},
	8:  {49, 49},
	9:  {53, 53},
	10: {57, 57},
	11: {61, 61},
	12: {65, 65},
	13: {69, 69},
	14: {73, 73},
	15: {77, 77},
	16: {81, 81},
	17: {85, 85},
	18: {89, 89},
	19: {93, 93},
	20: {97, 97},
	21: {101, 101},
	22: {105, 105},
	23: {109, 109},
	24: {113, 113},
	25: {117, 117},
	26: {121, 121},
	27: {125, 125},
	28: {129, 129},
	29: {133, 133},
	30: {137, 137},
}

var gf_exp [512]byte
var gf_log [256]byte

// EncodeAlphanumeric encode une chaîne alphanumérique en binaire selon les spécifications du QR code
func EncodeAlphanumeric(data string) (string, error) {
	var result string                                                    // Initialisation de la variable résultat qui contiendra la chaîne binaire finale
	alphanumericTable := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:" // Table des caractères alphanumériques autorisés dans les QR codes

	for i := 0; i < len(data); i += 2 { // Boucle sur la chaîne d'entrée par tranches de 2 caractères
		var bitString string // Déclare une variable pour stocker la représentation binaire des caractères actuels
		if i+1 < len(data) { // Vérifie s'il reste au moins deux caractères à traiter
			// Combine deux caractères en une valeur de 11 bits. Le 45 représente les 45 caractères dans alphanumericTable
			val := strings.IndexByte(alphanumericTable, data[i])*45 + strings.IndexByte(alphanumericTable, data[i+1])
			// Formate la valeur combinée en une chaîne binaire de 11 bits
			bitString = fmt.Sprintf("%011b", val)
		} else {
			// Si un seul caractère reste, convertit ce dernier en une valeur de 6 bits
			val := strings.IndexByte(alphanumericTable, data[i])
			// Formate la valeur en une chaîne binaire de 6 bits
			bitString = fmt.Sprintf("%06b", val)
		}
		// Ajoute la chaîne binaire obtenue au résultat final
		result += bitString

		// Log pour chaque itération de la boucle
		fmt.Printf("Processed characters %d-%d: %s (current result: %s)\n", i, i+min(2, len(data)-i)-1, data[i:min(i+2, len(data))], result)
	}

	fmt.Println("Encoding completed! Result:", result)
	return result, nil // Retourne la chaîne binaire finale et une erreur nil (pas d'erreur)
}

// EncodeNumeric encode une chaîne numérique en binaire selon les spécifications du QR code
func EncodeNumeric(data string) (string, error) {
	var result string
	// Traite les données par blocs de 3 chiffres
	for i := 0; i < len(data); i += 3 {
		// Sélectionne un bloc de 3 chiffres ou moins si c'est la fin de la chaîne
		num := data[i:min(i+3, len(data))]
		// Convertit le bloc en entier et gère les erreurs
		intValue, err := toInt(num)
		if err != nil {
			return "", err
		}
		// Convertit le bloc en entier et en binaire sur 10 bits
		bitString := fmt.Sprintf("%010b", intValue)
		// Ajoute la chaîne binaire au résultat
		result += bitString

		// Log après chaque conversion de bloc numérique
		fmt.Printf("Processed numeric block %d-%d: %s (current result: %s)\n", i, i+min(3, len(data)-i)-1, num, result)
	}
	fmt.Println("Encoding completed! Result:", result)
	return result, nil
}

// EncodeByte encode une chaîne de caractères en binaire selon les spécifications du QR code
func EncodeByte(data string) (string, error) {
	var result string
	for i := 0; i < len(data); i++ {
		result += fmt.Sprintf("%08b", data[i])

		// Log après chaque caractère encodé
		fmt.Printf("Processed character %d: %c (current result: %s)\n", i, data[i], result)
	}
	fmt.Println("Encoding completed! Result:", result)
	return result, nil
}

// EncodeKanji encode une chaîne de caractères Kanji en binaire selon les spécifications du QR code
func EncodeKanji(data string) (string, error) {
	var result string
	for i := 0; i < len(data); {
		// Décodage d'un rune à la fois
		r, size := utf8.DecodeRuneInString(data[i:])
		if r == utf8.RuneError {
			return "", fmt.Errorf("encodage UTF-8 invalide à la position %d", i)
		}
		i += size

		// Vérification et conversion selon les plages de caractères Kanji
		switch {
		case r >= 0x4E00 && r <= 0x9FCF:
			// Plage Kanji Level 1: 0x4E00 - 0x9FCF
			r -= 0x4E00
		case r >= 0x3400 && r <= 0x4DBF:
			// Plage Kanji Level 2: 0x3400 - 0x4DBF
			r -= 0x3400
		case r >= 0x20000 && r <= 0x2A6DF:
			// Plage Kanji Level 3: 0x20000 - 0x2A6DF
			r -= 0x20000
		case r >= 0x2A700 && r <= 0x2B73F:
			// Plage Kanji Level 4: 0x2A700 - 0x2B73F
			r -= 0x2A700
		default:
			return "", fmt.Errorf("caractère hors de portée Kanji: %U", r)
		}

		// Séparation en MSB et LSB pour la conversion binaire
		msb := (r >> 8) & 0xFF
		lsb := r & 0xFF
		code := (msb << 8) | lsb

		// Conversion en format binaire sur 13 bits
		bitString := fmt.Sprintf("%016b", code)
		result += bitString

		// Log après chaque caractère Kanji encodé
		fmt.Printf("Caractère Kanji encodé: %c (résultat actuel: %s)\n", r, result)
	}
	fmt.Println("Encodage terminé ! Résultat:", result)
	return result, nil
}

// toInt convertit une chaîne de caractères en entier
func toInt(s string) (int, error) {
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// min renvoie le minimum de deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Fonction utilitaire pour détecter le type de données à encoder
func detectDataType(data string) string {
	if _, err := strconv.Atoi(data); err == nil {
		return "numeric"
	} else if isAlphanumeric(data) {
		return "alphanumeric"
	} else if isUTF8(data) {
		return "kanji"
	} else {
		return "byte"
	}
}

// Fonction utilitaire pour vérifier si une chaîne est alphanumérique
func isAlphanumeric(data string) bool {
	for _, r := range data {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// Fonction utilitaire pour vérifier si une chaîne est UTF-8
func isUTF8(data string) bool {
	return utf8.ValidString(data)
}

// GenerateQRMatrix génère une matrice QR pour une version spécifique à partir des données spécifiées
func GenerateQRMatrix(version int, data string) *image.RGBA {
	info, ok := versionInfo[version]
	if !ok {
		return nil // Gérer le cas où la version demandée n'est pas définie
	}

	size := info.Size
	qrMatrix := image.NewRGBA(image.Rect(0, 0, size, size))

	// Ajout des motifs de positionnement, de timing et d'alignement
	addPositionPatterns(qrMatrix)
	addTimingPatterns(qrMatrix)
	addAlignmentPatterns(qrMatrix, version)

	// Encodage des données en fonction de leur type
	var encodedData string
	switch detectDataType(data) {
	case "numeric":
		encodedData, _ = EncodeNumeric(data)
	case "alphanumeric":
		encodedData, _ = EncodeAlphanumeric(data)
	case "byte":
		encodedData, _ = EncodeByte(data)
	case "kanji":
		encodedData, _ = EncodeKanji(data)
	default:
		return nil // Gérer le cas où le type de données n'est pas supporté
	}

	// Ajout de la redondance (correction d'erreurs)
	errorCorrectionData := generateErrorCorrection([]byte(encodedData), "L") // Exemple avec niveau de correction "L"
	fullData := encodedData + string(errorCorrectionData)

	// Remplissage de la matrice QR avec les données encodées
	index := 0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if index < len(fullData) && fullData[index] == '1' {
				qrMatrix.Set(x, y, color.Black)
				index++
			} else {
				qrMatrix.Set(x, y, color.White)
			}
		}
	}

	// Appliquer le masquage (sélectionnez le meilleur masque parmi les 8)
	bestMaskedMatrix := applyMask(qrMatrix, 0) // Exemple avec le masque 0

	return bestMaskedMatrix
}

// Ajoute les motifs de positionnement à la matrice QR
func addPositionPatterns(matrix *image.RGBA) {
	size := matrix.Bounds().Max.X
	positions := []struct {
		x, y int
	}{
		{0, 0},
		{size - 7, 0},
		{0, size - 7},
	}

	for _, pos := range positions {
		for i := 0; i < 7; i++ {
			for j := 0; j < 7; j++ {
				if i == 0 || i == 6 || j == 0 || j == 6 || (i >= 2 && i <= 4 && j >= 2 && j <= 4) {
					matrix.Set(pos.x+i, pos.y+j, color.Black)
				} else {
					matrix.Set(pos.x+i, pos.y+j, color.White)
				}
			}
		}
	}
}

// Ajoute les motifs de timing à la matrice QR
func addTimingPatterns(matrix *image.RGBA) {
	size := matrix.Bounds().Max.X
	for i := 8; i < size-8; i++ {
		col := color.RGBA{255, 255, 255, 255} // Blanc par défaut
		if i%2 == 0 {
			col = color.RGBA{0, 0, 0, 255} // Noir sur les cases paires
		}
		matrix.Set(i, 6, col)
		matrix.Set(6, i, col)
	}
}

// Ajoute les motifs d'alignement à la matrice QR
func addAlignmentPatterns(matrix *image.RGBA, version int) {
	positions := getAlignmentPatternPositions(version)
	for _, pos := range positions {
		addAlignmentPattern(matrix, pos.x, pos.y)
	}
}

// Positions des motifs d'alignement pour chaque version
func getAlignmentPatternPositions(version int) []struct{ x, y int } {
	positionMap := map[int][]int{
		1:  {},
		2:  {6, 18},
		3:  {6, 22},
		4:  {6, 26},
		5:  {6, 30},
		6:  {6, 34},
		7:  {6, 22, 38},
		8:  {6, 24, 42},
		9:  {6, 26, 46},
		10: {6, 28, 50},
		11: {6, 30, 54},
		12: {6, 32, 58},
		13: {6, 34, 62},
		14: {6, 26, 46, 66},
		15: {6, 26, 48, 70},
		16: {6, 26, 50, 74},
		17: {6, 30, 54, 78},
		18: {6, 30, 56, 82},
		19: {6, 30, 58, 86},
		20: {6, 34, 62, 90},
		21: {6, 28, 50, 72, 94},
		22: {6, 26, 50, 74, 98},
		23: {6, 30, 54, 78, 102},
		24: {6, 28, 54, 80, 106},
		25: {6, 32, 58, 84, 110},
		26: {6, 30, 58, 86, 114},
		27: {6, 34, 62, 90, 118},
		28: {6, 26, 50, 74, 98, 122},
		29: {6, 30, 54, 78, 102, 126},
		30: {6, 26, 52, 78, 104, 130},
	}

	positions, exists := positionMap[version]
	if !exists {
		return []struct{ x, y int }{}
	}

	var result []struct{ x, y int }
	for _, x := range positions {
		for _, y := range positions {
			// Skip the position if it's at the top-left, top-right, or bottom-left corners
			if (x == 6 && y == 6) || (x == 6 && y == positions[len(positions)-1]) || (x == positions[len(positions)-1] && y == 6) {
				continue
			}
			result = append(result, struct{ x, y int }{x, y})
		}
	}

	return result
}

// Ajoute un motif d'alignement à une position donnée
func addAlignmentPattern(matrix *image.RGBA, x, y int) {
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			if i == -2 || i == 2 || j == -2 || j == 2 || (i == 0 && j == 0) {
				matrix.Set(x+i, y+j, color.Black)
			} else {
				matrix.Set(x+i, y+j, color.White)
			}
		}
	}
}

// Applique un masque à la matrice QR et retourne la matrice masquée
func applyMask(matrix *image.RGBA, maskPattern int) *image.RGBA {
	size := matrix.Bounds().Max.X
	maskedMatrix := image.NewRGBA(matrix.Bounds())
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			originalColor := matrix.At(x, y)
			shouldInvert := false
			switch maskPattern {
			case 0:
				shouldInvert = (x+y)%2 == 0
			case 1:
				shouldInvert = y%2 == 0
			case 2:
				shouldInvert = x%3 == 0
			case 3:
				shouldInvert = (x+y)%3 == 0
			case 4:
				shouldInvert = (y/2+x/3)%2 == 0
			case 5:
				shouldInvert = ((x*y)%2 + (x*y)%3) == 0
			case 6:
				shouldInvert = (((x*y)%2 + (x*y)%3) % 2) == 0
			case 7:
				shouldInvert = (((x+y)%2 + (x*y)%3) % 2) == 0
			}
			if shouldInvert {
				r, g, b, a := originalColor.RGBA()
				if r == 0 && g == 0 && b == 0 && a == 0xFFFF { // Vérifie si la couleur originale est complètement noire et opaque
					maskedMatrix.Set(x, y, color.White) // Inverser le noir en blanc
				} else { // Autre couleur que le noir complet
					maskedMatrix.Set(x, y, color.Black) // Ne pas inverser, garder la couleur originale ou inverser en noir
				}
			} else {
				maskedMatrix.Set(x, y, originalColor) // Pas d'inversion, utiliser la couleur originale
			}

		}
	}
	return maskedMatrix
}

// Initialisation du champ de Galois
func initGaloisField() {
	var x int = 1 // Changez ici de byte à int
	for i := 0; i < 255; i++ {
		gf_exp[i] = byte(x) // Assurez-vous de convertir à byte ici
		gf_log[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x1D // Polynôme générateur pour GF(2^8)
		}
	}
	for i := 255; i < 512; i++ {
		gf_exp[i] = gf_exp[i-255]
	}
}

// Fonction de multiplication dans le champ de Galois
func gfMultiply(x, y byte) byte {
	if x == 0 || y == 0 {
		return 0
	}
	return gf_exp[gf_log[x]+gf_log[y]]
}

// Génère les syndromes Reed-Solomon
func generateReedSolomon(data []byte, numECBytes int) []byte {
	// Initialise le tableau de correction d'erreurs (syndromes)
	ecBytes := make([]byte, numECBytes)

	// Parcourt chaque octet des données
	for _, dataByte := range data {
		// Calcule le coefficient principal
		leadCoefficient := dataByte ^ ecBytes[0]

		// Décale les syndromes et les met à jour
		for i := 0; i < numECBytes-1; i++ {
			ecBytes[i] = ecBytes[i+1] ^ gfMultiply(leadCoefficient, gf_exp[i])
		}
		// Ajoute l'octet suivant dans le champ de Galois
		ecBytes[numECBytes-1] = gfMultiply(leadCoefficient, gf_exp[numECBytes-1])
	}

	return ecBytes
}

// Gère le niveau de correction d'erreurs et appelle la fonction Reed-Solomon
func generateErrorCorrection(data []byte, level string) []byte {
	// Table des niveaux de correction : L = 7%, M = 15%, Q = 25%, H = 30%
	errorCorrectionLevels := map[string]int{
		"L": 7,
		"M": 15,
		"Q": 25,
		"H": 30,
	}

	// Récupère le nombre de bytes de correction pour le niveau donné
	numECBytes := errorCorrectionLevels[level]

	// Appelle la fonction Reed-Solomon pour générer les bytes de correction
	return generateReedSolomon(data, numECBytes)
}
