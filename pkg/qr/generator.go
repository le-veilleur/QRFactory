package qr

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
	"sync"
)

// CalculateMinVersion calcule la version minimale nécessaire pour les données
// Retourne la version minimale ou une erreur si les données sont trop longues
func CalculateMinVersion(data string) (int, error) {
	// Calculer la taille totale nécessaire en bits
	totalBits := 4 + 8 + (len(data) * 8) // Mode (4) + Length (8) + Data (8 par caractère)

	// Table de capacité en bits pour le mode byte avec niveau de correction M
	capacityTable := map[int]int{
		1:  128,  // 16 octets
		2:  224,  // 28 octets
		3:  352,  // 44 octets
		4:  512,  // 64 octets
		5:  688,  // 86 octets
		6:  864,  // 108 octets
		7:  992,  // 124 octets
		8:  1232, // 154 octets
		9:  1456, // 182 octets
		10: 1728, // 216 octets
	}

	// Trouver la première version qui peut contenir les données
	for version := 1; version <= 10; version++ {
		if capacity := capacityTable[version]; capacity >= totalBits {
			return version, nil
		}
	}

	return 0, fmt.Errorf("impossible de stocker les données même avec la version maximale")
}

// CalculateMinVersionForDataType calcule la version minimale nécessaire pour les données selon le type
func CalculateMinVersionForDataType(data string, dataType string) (int, error) {
	var totalBits int

	// Calcule les bits nécessaires selon le type de données
	switch dataType {
	case "numeric":
		// Mode (4) + Length (10 pour version 1-9) + Data (~3.33 bits par caractère)
		dataLength := len(data)
		dataBits := (dataLength / 3) * 10
		if dataLength%3 == 1 {
			dataBits += 4 // 1 chiffre = 4 bits
		} else if dataLength%3 == 2 {
			dataBits += 7 // 2 chiffres = 7 bits
		}
		totalBits = 4 + 10 + dataBits
	case "alphanumeric":
		// Mode (4) + Length (9 pour version 1-9) + Data (~5.5 bits par caractère)
		dataLength := len(data)
		dataBits := (dataLength / 2) * 11
		if dataLength%2 == 1 {
			dataBits += 6 // 1 caractère seul = 6 bits
		}
		totalBits = 4 + 9 + dataBits
	case "kanji":
		// Mode (4) + Length (8 pour version 1-9) + Data (13 bits par caractère)
		totalBits = 4 + 8 + (len(data) * 13)
	default: // "byte" ou autre
		// Mode (4) + Length (8 pour version 1-9) + Data (8 bits par caractère)
		totalBits = 4 + 8 + (len(data) * 8)
	}

	// Table de capacité en bits pour différents modes avec niveau de correction M
	capacityTable := map[int]int{
		1:  128,  // 16 octets
		2:  224,  // 28 octets
		3:  352,  // 44 octets
		4:  512,  // 64 octets
		5:  688,  // 86 octets
		6:  864,  // 108 octets
		7:  992,  // 124 octets
		8:  1232, // 154 octets
		9:  1456, // 182 octets
		10: 1728, // 216 octets
	}

	// Trouver la première version qui peut contenir les données
	for version := 1; version <= 10; version++ {
		if capacity := capacityTable[version]; capacity >= totalBits {
			return version, nil
		}
	}

	return 0, fmt.Errorf("impossible de stocker les données %s de type %s même avec la version maximale", data, dataType)
}

// GenerateQRMatrix génère la matrice QR pour les données fournies
func GenerateQRMatrix(version int, data string, errorCorrectionLevel string) *image.RGBA {
	// Valider le niveau de correction d'erreur
	if errorCorrectionLevel != "L" && errorCorrectionLevel != "M" &&
		errorCorrectionLevel != "Q" && errorCorrectionLevel != "H" {
		// Par défaut, utiliser le niveau M
		errorCorrectionLevel = "M"
		fmt.Printf("Niveau de correction d'erreur invalide, utilisation du niveau M (15%%)\n")
	}

	// Détecter le type de données
	dataType := DetectDataType(data)
	fmt.Printf("Type de données détecté: %s\n", dataType)

	// Calculer la version minimale nécessaire
	minVersion, minVersionErr := CalculateMinVersionForDataType(data, dataType)
	if minVersionErr != nil {
		fmt.Printf("ERREUR: %v\n", minVersionErr)
		return nil
	}

	// Utiliser la version minimale si la version fournie est trop petite
	if version < minVersion {
		fmt.Printf("La version %d est trop petite pour les données. Utilisation de la version %d.\n", version, minVersion)
		version = minVersion
	}

	// Vérifier la validité de la version
	if version < 1 || version > 40 {
		return nil
	}

	// Préparer l'encodage des données
	var encodedData strings.Builder
	var modeIndicator, lengthBits string
	var encodedBits string
	var encodingErr error

	// Configurer selon le type de données
	switch dataType {
	case "numeric":
		modeIndicator = "0001" // Indicateur mode numérique
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%010b", len(data)) // 10 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%012b", len(data)) // 12 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeNumeric(data)
	case "alphanumeric":
		modeIndicator = "0010" // Indicateur mode alphanumérique
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%09b", len(data)) // 9 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%011b", len(data)) // 11 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeAlphanumeric(data)
	case "kanji":
		modeIndicator = "1000" // Indicateur mode kanji
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%08b", len(data)) // 8 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%010b", len(data)) // 10 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeKanji(data)
	default: // "byte" ou autre
		modeIndicator = "0100" // Indicateur mode byte
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%08b", len(data)) // 8 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%016b", len(data)) // 16 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeByte(data)
	}

	if encodingErr != nil {
		fmt.Printf("Erreur d'encodage: %v\n", encodingErr)
		return nil
	}

	// Ajouter l'indicateur de mode
	encodedData.WriteString(modeIndicator)

	// Ajouter la longueur des données
	encodedData.WriteString(lengthBits)

	// Ajouter les données encodées
	encodedData.WriteString(encodedBits)

	// Calculer la capacité disponible
	size := version*4 + 17
	matrix := image.NewRGBA(image.Rect(0, 0, size, size))

	// Initialiser la matrice en blanc
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			matrix.Set(x, y, color.White)
		}
	}

	// Ajouter les motifs de repérage
	AddFinderPatterns(matrix)
	AddSeparators(matrix)
	AddAlignmentPatterns(matrix, version)
	AddTimingPatterns(matrix)

	// Calculer la capacité disponible
	capacity := calculateAvailableCapacity(size)

	// Vérifier si les données encodées dépassent la capacité
	if encodedData.Len() > capacity {
		// Si c'est le cas, essayer avec une version supérieure
		newVersion := version + 1
		fmt.Printf("ATTENTION: Les données encodées (%d bits) dépassent la capacité (%d bits)\n", encodedData.Len(), capacity)
		fmt.Printf("Tentative avec la version %d...\n", newVersion)
		return GenerateQRMatrix(newVersion, data, errorCorrectionLevel) // Appel récursif avec version supérieure
	}

	// Ajouter le terminateur
	for encodedData.Len() < capacity && encodedData.Len() < capacity-4 {
		encodedData.WriteString("0")
	}

	// Ajouter la correction d'erreur
	finalData := AddErrorCorrection(encodedData.String(), errorCorrectionLevel, version)

	fmt.Printf("Niveau de correction: %s\n", errorCorrectionLevel)
	fmt.Printf("Capacité disponible: %d bits\n", capacity)
	fmt.Printf("Longueur des données encodées: %d bits\n", len(finalData))
	fmt.Printf("Données encodées: %s\n", finalData)

	// Placer les données
	PlaceData(matrix, finalData)

	// Appliquer le meilleur masque
	bestScore := math.MaxInt32
	var bestMatrix *image.RGBA
	bestMask := 0

	fmt.Println("Évaluation des masques:")
	for mask := 0; mask < 8; mask++ {
		maskedMatrix := ApplyMask(matrix, mask)
		score := EvaluateMask(maskedMatrix)
		fmt.Printf("  Masque %d: score %d\n", mask, score)

		if score < bestScore {
			bestScore = score
			bestMatrix = maskedMatrix
			bestMask = mask
		}
	}

	fmt.Printf("Meilleur masque sélectionné: %d (score: %d)\n", bestMask, bestScore)

	if bestMatrix != nil {
		matrix = bestMatrix
		// Ajouter les informations de format après avoir appliqué le masque
		AddFormatInfo(matrix, errorCorrectionLevel, bestMask)
	}

	return matrix
}

// EvaluateMask évalue la qualité d'un masque selon les règles de pénalité du QR code
func EvaluateMask(matrix *image.RGBA) int {
	score := 0
	size := matrix.Bounds().Max.X

	// Règle 1: Pénalité pour 5+ modules de même couleur consécutifs
	score += evaluateRule1(matrix, size)

	// Règle 2: Pénalité pour les blocs de couleur 2x2
	score += evaluateRule2(matrix, size)

	// Règle 3: Pénalité pour motifs spécifiques ressemblant aux finder patterns
	score += evaluateRule3(matrix, size)

	// Règle 4: Équilibre entre modules noirs et blancs
	score += evaluateRule4(matrix, size)

	return score
}

// evaluateRule1 calcule la pénalité pour les séquences de modules de même couleur
func evaluateRule1(matrix *image.RGBA, size int) int {
	penalty := 0

	// Vérifier les lignes horizontales
	for y := 0; y < size; y++ {
		count := 1
		color := isBlack(matrix.At(0, y))

		for x := 1; x < size; x++ {
			if isBlack(matrix.At(x, y)) == color {
				count++
			} else {
				if count >= 5 {
					penalty += 3 + (count - 5)
				}
				count = 1
				color = isBlack(matrix.At(x, y))
			}
		}

		// Vérifier la séquence finale de la ligne
		if count >= 5 {
			penalty += 3 + (count - 5)
		}
	}

	// Vérifier les colonnes verticales
	for x := 0; x < size; x++ {
		count := 1
		color := isBlack(matrix.At(x, 0))

		for y := 1; y < size; y++ {
			if isBlack(matrix.At(x, y)) == color {
				count++
			} else {
				if count >= 5 {
					penalty += 3 + (count - 5)
				}
				count = 1
				color = isBlack(matrix.At(x, y))
			}
		}

		// Vérifier la séquence finale de la colonne
		if count >= 5 {
			penalty += 3 + (count - 5)
		}
	}

	return penalty
}

// evaluateRule2 calcule la pénalité pour les blocs 2x2 de même couleur
func evaluateRule2(matrix *image.RGBA, size int) int {
	penalty := 0

	for y := 0; y < size-1; y++ {
		for x := 0; x < size-1; x++ {
			color := isBlack(matrix.At(x, y))
			if isBlack(matrix.At(x+1, y)) == color &&
				isBlack(matrix.At(x, y+1)) == color &&
				isBlack(matrix.At(x+1, y+1)) == color {
				penalty += 3
			}
		}
	}

	return penalty
}

// evaluateRule3 calcule la pénalité pour les motifs finder-like (1:1:3:1:1)
func evaluateRule3(matrix *image.RGBA, size int) int {
	penalty := 0

	// Motif horizontal: noir-blanc-noir-noir-noir-blanc-noir
	pattern1 := []bool{true, false, true, true, true, false, true}
	// Motif vertical équivalent
	pattern2 := []bool{true, false, true, true, true, false, true}

	// Rechercher les motifs horizontaux
	for y := 0; y < size; y++ {
		for x := 0; x <= size-7; x++ {
			match := true
			for i := 0; i < 7; i++ {
				if isBlack(matrix.At(x+i, y)) != pattern1[i] {
					match = false
					break
				}
			}
			if match {
				penalty += 40
			}
		}
	}

	// Rechercher les motifs verticaux
	for x := 0; x < size; x++ {
		for y := 0; y <= size-7; y++ {
			match := true
			for i := 0; i < 7; i++ {
				if isBlack(matrix.At(x, y+i)) != pattern2[i] {
					match = false
					break
				}
			}
			if match {
				penalty += 40
			}
		}
	}

	return penalty
}

// evaluateRule4 calcule la pénalité pour le déséquilibre noir/blanc
func evaluateRule4(matrix *image.RGBA, size int) int {
	blackCount := 0
	totalCount := size * size

	// Compter les modules noirs
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if isBlack(matrix.At(x, y)) {
				blackCount++
			}
		}
	}

	// Calculer le pourcentage de modules noirs
	blackPercentage := (blackCount * 100) / totalCount

	// Calculer la différence avec 50% par tranches de 5%
	fivePercentDeviation := abs((blackPercentage - 50) / 5)

	return fivePercentDeviation * 10
}

// isBlack vérifie si une couleur est noire
func isBlack(c color.Color) bool {
	r, _, _, _ := c.RGBA()
	// Les valeurs RGBA sont normalisées sur 16 bits (0-65535)
	// On considère noir si la valeur est inférieure à la moitié (0x8000)
	return r < 0x8000
}

// AddFinderPatterns ajoute les motifs de positionnement à la matrice QR
func AddFinderPatterns(matrix *image.RGBA) {
	// Positions des motifs de positionnement (en haut à gauche, en haut à droite, en bas à gauche)
	positions := []struct{ x, y int }{
		{0, 0},                         // En haut à gauche
		{matrix.Bounds().Max.X - 7, 0}, // En haut à droite
		{0, matrix.Bounds().Max.Y - 7}, // En bas à gauche
	}

	for _, pos := range positions {
		// Dessiner le carré extérieur 7x7
		for i := 0; i < 7; i++ {
			for j := 0; j < 7; j++ {
				if i == 0 || i == 6 || j == 0 || j == 6 {
					matrix.Set(pos.x+i, pos.y+j, color.Black)
				}
			}
		}

		// Dessiner le carré intérieur 5x5
		for i := 1; i < 6; i++ {
			for j := 1; j < 6; j++ {
				matrix.Set(pos.x+i, pos.y+j, color.White)
			}
		}

		// Dessiner le carré central 3x3
		for i := 2; i < 5; i++ {
			for j := 2; j < 5; j++ {
				matrix.Set(pos.x+i, pos.y+j, color.Black)
			}
		}
	}
}

// AddSeparators ajoute les séparateurs à la matrice QR
func AddSeparators(matrix *image.RGBA) {
	size := matrix.Bounds().Max.X

	// Séparateurs horizontaux
	for x := 0; x < 8; x++ {
		matrix.Set(x, 7, color.White)        // En haut à gauche
		matrix.Set(size-8+x, 7, color.White) // En haut à droite
		matrix.Set(x, size-8, color.White)   // En bas à gauche
	}

	// Séparateurs verticaux
	for y := 0; y < 8; y++ {
		matrix.Set(7, y, color.White)        // En haut à gauche
		matrix.Set(size-8, y, color.White)   // En haut à droite
		matrix.Set(7, size-8+y, color.White) // En bas à gauche
	}
}

// PlaceData place les données dans la matrice QR selon le motif en zigzag
func PlaceData(matrix *image.RGBA, data string) {
	size := matrix.Bounds().Max.X
	dataIndex := 0
	upward := true

	// Vérifier la longueur des données
	fmt.Printf("Données à placer : %s (longueur: %d)\n", data, len(data))

	// Calculer le nombre total de modules disponibles pour les données
	totalModules := 0

	// Pré-calculer toutes les positions valides pour les données
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			if isValidDataPosition(x, y, size) {
				totalModules++
			}
		}
	}

	// Préparer une liste de valeurs à placer
	// Cette liste contient d'abord les données encodées, puis des bits de remplissage (padding)
	// qui suivent le standard QR (alternance de 236 et 17 en décimal, ou 11101100 et 00010001 en binaire)
	var valuesToPlace []byte

	// Ajouter les données encodées
	for i := 0; i < len(data); i++ {
		if data[i] == '1' {
			valuesToPlace = append(valuesToPlace, 1)
		} else if data[i] == '0' {
			valuesToPlace = append(valuesToPlace, 0)
		} else {
			// Ignorer les séparateurs mais compter l'index
			dataIndex++
			continue
		}
	}

	// Compléter avec les bits de remplissage selon le standard QR
	// Remplir avec des motifs 11101100 et 00010001 en alternance
	paddingPattern := []byte{1, 1, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1}
	paddingIndex := 0

	for len(valuesToPlace) < totalModules {
		valuesToPlace = append(valuesToPlace, paddingPattern[paddingIndex%len(paddingPattern)])
		paddingIndex++
	}

	// Remplir efficacement la matrice en zigzag, en commençant par le coin en bas à droite
	for x := size - 1; x >= 0; x -= 2 {
		// Traiter deux colonnes à la fois (la colonne actuelle et celle à sa gauche)
		for dx := 0; dx <= 1 && x-dx >= 0; dx++ {
			currentX := x - dx

			// Sauter la colonne de timing
			if currentX == 6 {
				continue
			}

			if upward {
				// Monter (de bas en haut)
				for y := size - 1; y >= 0; y-- {
					if isValidDataPosition(currentX, y, size) {
						if dataIndex < len(data) && data[dataIndex] != '-' {
							fmt.Printf("Module placé à (%d,%d): %c [Total: %d]\n",
								currentX, y, data[dataIndex], dataIndex+1)
							if data[dataIndex] == '1' {
								matrix.Set(currentX, y, color.Black)
							} else {
								matrix.Set(currentX, y, color.White)
							}
						} else if dataIndex < len(data) {
							// Si c'est '-' ou autre (séparateur), on saute ce bit mais on compte
							fmt.Printf("Module ignoré à (%d,%d): %c [séparateur]\n",
								currentX, y, data[dataIndex])
						} else {
							// Si on a fini de placer les données, utiliser le padding calculé
							valueIndex := dataIndex - (len(data) - len(valuesToPlace))
							if valueIndex >= 0 && valueIndex < len(valuesToPlace) {
								if valuesToPlace[valueIndex] == 1 {
									matrix.Set(currentX, y, color.Black)
								} else {
									matrix.Set(currentX, y, color.White)
								}
							}
						}
						dataIndex++
					}
				}
			} else {
				// Descendre (de haut en bas)
				for y := 0; y < size; y++ {
					if isValidDataPosition(currentX, y, size) {
						if dataIndex < len(data) && data[dataIndex] != '-' {
							fmt.Printf("Module placé à (%d,%d): %c [Total: %d]\n",
								currentX, y, data[dataIndex], dataIndex+1)
							if data[dataIndex] == '1' {
								matrix.Set(currentX, y, color.Black)
							} else {
								matrix.Set(currentX, y, color.White)
							}
						} else if dataIndex < len(data) {
							// Si c'est '-' ou autre (séparateur), on saute ce bit mais on compte
							fmt.Printf("Module ignoré à (%d,%d): %c [séparateur]\n",
								currentX, y, data[dataIndex])
						} else {
							// Si on a fini de placer les données, utiliser le padding calculé
							valueIndex := dataIndex - (len(data) - len(valuesToPlace))
							if valueIndex >= 0 && valueIndex < len(valuesToPlace) {
								if valuesToPlace[valueIndex] == 1 {
									matrix.Set(currentX, y, color.Black)
								} else {
									matrix.Set(currentX, y, color.White)
								}
							}
						}
						dataIndex++
					}
				}
			}
			upward = !upward
		}
	}

	fmt.Printf("Total des modules placés : %d\n", dataIndex)
}

// isValidDataPosition vérifie si une position peut contenir des données
func isValidDataPosition(x, y, size int) bool {
	// Vérifier les motifs de positionnement
	if (x < 9 && y < 9) || // En haut à gauche
		(x > size-9 && y < 9) || // En haut à droite
		(x < 9 && y > size-9) { // En bas à gauche
		return false
	}

	// Vérifier la colonne de timing
	if x == 6 {
		return false
	}

	// Vérifier la ligne de timing
	if y == 6 {
		return false
	}

	// Vérifier les motifs d'alignement
	alignmentPositions := getAlignmentPositions(size)
	for _, pos := range alignmentPositions {
		ax, ay := pos[0], pos[1]
		// Ne pas exclure la zone centrale du motif d'alignement
		if (x == ax-2 || x == ax+2 || y == ay-2 || y == ay+2) &&
			(x >= ax-2 && x <= ax+2 && y >= ay-2 && y <= ay+2) {
			return false
		}
	}

	return true
}

// getAlignmentPositions calcule les positions des motifs d'alignement
func getAlignmentPositions(size int) [][2]int {
	version := (size - 17) / 4
	if version < 2 {
		return [][2]int{}
	}

	// Table officielle des positions d'alignement pour les versions 2 à 7
	alignmentTable := map[int][]int{
		2: {6, 18},
		3: {6, 22},
		4: {6, 26},
		5: {6, 30},
		6: {6, 34},
		7: {6, 22, 38},
	}

	var positions []int
	if pos, ok := alignmentTable[version]; ok {
		positions = pos
	} else {
		// Pour les versions supérieures, calcule dynamiquement
		numAlign := version/7 + 2
		step := 0
		if numAlign > 2 {
			step = (size - 13) / (numAlign - 1)
		}
		positions = make([]int, numAlign)
		positions[0] = 6
		for i := 1; i < numAlign-1; i++ {
			positions[i] = positions[i-1] + step
		}
		positions[numAlign-1] = size - 7
	}

	// Génère les couples (x, y) pour chaque motif d'alignement
	var result [][2]int
	for _, x := range positions {
		for _, y := range positions {
			// On évite les coins où il y a déjà un finder pattern
			if !((x == 6 && y == 6) ||
				(x == 6 && y == size-7) ||
				(x == size-7 && y == 6)) {
				result = append(result, [2]int{x, y})
			}
		}
	}
	return result
}

// AddTimingPatterns ajoute les motifs de timing à la matrice QR
func AddTimingPatterns(matrix *image.RGBA) {
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

// AddAlignmentPatterns ajoute les motifs d'alignement à la matrice QR
func AddAlignmentPatterns(matrix *image.RGBA, version int) {
	positions := GetAlignmentPatternPositions(version)
	for _, pos := range positions {
		AddAlignmentPattern(matrix, pos.x, pos.y)
	}
}

// Ajoute un motif d'alignement à une position donnée
func AddAlignmentPattern(matrix *image.RGBA, x, y int) {
	// Dessiner le carré extérieur 5x5
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			if i == -2 || i == 2 || j == -2 || j == 2 {
				matrix.Set(x+i, y+j, color.Black)
			} else {
				matrix.Set(x+i, y+j, color.White)
			}
		}
	}

	// Dessiner le carré intérieur 3x3
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == -1 || i == 1 || j == -1 || j == 1 {
				matrix.Set(x+i, y+j, color.White)
			} else {
				matrix.Set(x+i, y+j, color.Black)
			}
		}
	}
}

// Structure pour les tâches de traitement
type DataProcessingTask struct {
	data   string
	result chan string
}

// Pool de workers pour le traitement des données
func ProcessDataWithWorkers(data string, numWorkers int) string {
	tasks := make(chan DataProcessingTask)

	// Démarrage des workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range tasks {
				var result string
				switch DetectDataType(task.data) {
				case "numeric":
					result, _ = EncodeNumeric(task.data)
				case "alphanumeric":
					result, _ = EncodeAlphanumeric(task.data)
				case "byte":
					result, _ = EncodeByte(task.data)
				case "kanji":
					result, _ = EncodeKanji(task.data)
				}
				task.result <- result
			}
		}()
	}

	// Diviser les données en chunks et les envoyer aux workers
	chunkSize := len(data) / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	var finalResult strings.Builder
	var wg sync.WaitGroup

	for i := 0; i < len(data); i += chunkSize {
		wg.Add(1)
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		resultChan := make(chan string)
		tasks <- DataProcessingTask{
			data:   data[i:end],
			result: resultChan,
		}

		go func() {
			defer wg.Done()
			result := <-resultChan
			finalResult.WriteString(result)
		}()
	}

	wg.Wait()
	close(tasks)

	return finalResult.String()
}

// SaveQRImage sauvegarde la matrice QR en image PNG
func SaveQRImage(matrix *image.RGBA, outputFile string, scale int) error {
	bounds := matrix.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Création d'une nouvelle image avec la taille mise à l'échelle
	scaledImage := image.NewRGBA(image.Rect(0, 0, width*scale, height*scale))

	// Mise à l'échelle de l'image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := matrix.At(x, y)
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					scaledImage.Set(x*scale+sx, y*scale+sy, color)
				}
			}
		}
	}

	// Création du fichier de sortie
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier : %v", err)
	}
	defer file.Close()

	// Encodage en PNG
	if err := png.Encode(file, scaledImage); err != nil {
		return fmt.Errorf("erreur lors de l'encodage PNG : %v", err)
	}

	return nil
}

// SaveQRImageWithQuietZone sauvegarde la matrice QR en image PNG avec une zone calme (quiet zone)
func SaveQRImageWithQuietZone(matrix *image.RGBA, outputFile string, scale int, quietZone int) error {
	bounds := matrix.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Calculer les dimensions avec la zone calme
	totalWidth := width + (quietZone * 2)
	totalHeight := height + (quietZone * 2)

	// Création d'une nouvelle image avec la taille mise à l'échelle incluant la zone calme
	scaledImage := image.NewRGBA(image.Rect(0, 0, totalWidth*scale, totalHeight*scale))

	// Remplir l'image entière en blanc pour la zone calme
	for y := 0; y < totalHeight*scale; y++ {
		for x := 0; x < totalWidth*scale; x++ {
			scaledImage.Set(x, y, color.White)
		}
	}

	// Mise à l'échelle de l'image QR au centre
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := matrix.At(x, y)
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					scaledImage.Set((x+quietZone)*scale+sx, (y+quietZone)*scale+sy, c)
				}
			}
		}
	}

	// Création du fichier de sortie
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier : %v", err)
	}
	defer file.Close()

	// Encodage en PNG
	if err := png.Encode(file, scaledImage); err != nil {
		return fmt.Errorf("erreur lors de l'encodage PNG : %v", err)
	}

	return nil
}

// AddFormatInfo ajoute l'information de format au QR code
func AddFormatInfo(matrix *image.RGBA, ecLevel string, maskPattern int) {
	formatInfoBits := FormatInfo[ecLevel][maskPattern]
	size := matrix.Bounds().Max.X

	// Placer l'information de format autour du motif de positionnement en haut à gauche
	for i := 0; i < 15; i++ {
		bit := formatInfoBits[i] == '1'
		if i < 6 {
			// Position horizontale
			if bit {
				matrix.Set(i, 8, color.Black)
			} else {
				matrix.Set(i, 8, color.White)
			}
		} else if i < 8 {
			// Position horizontale (après le timing pattern)
			if bit {
				matrix.Set(i+1, 8, color.Black)
			} else {
				matrix.Set(i+1, 8, color.White)
			}
		} else {
			// Position verticale
			if bit {
				matrix.Set(8, size-1-(14-i), color.Black)
			} else {
				matrix.Set(8, size-1-(14-i), color.White)
			}
		}

		// Copier l'information de format sur le côté droit et en bas
		if i < 7 {
			if bit {
				matrix.Set(size-7+i, 8, color.Black)
			} else {
				matrix.Set(size-7+i, 8, color.White)
			}
		} else {
			if bit {
				matrix.Set(8, 6-(i-7), color.Black)
			} else {
				matrix.Set(8, 6-(i-7), color.White)
			}
		}
	}
}

// Ajout d'une fonction pour calculer la capacité disponible
func calculateAvailableCapacity(size int) int {
	// Version calculée à partir de la taille
	version := (size - 17) / 4

	// Tableau de capacités officielles en bits pour chaque version (Mode byte, niveau de correction M)
	// Ces valeurs sont tirées de la spécification ISO/IEC 18004
	capacityTable := map[int]int{
		1:  128,  // 16 octets
		2:  224,  // 28 octets
		3:  352,  // 44 octets
		4:  512,  // 64 octets
		5:  688,  // 86 octets
		6:  864,  // 108 octets
		7:  992,  // 124 octets
		8:  1232, // 154 octets
		9:  1456, // 182 octets
		10: 1728, // 216 octets
	}

	// Utiliser la valeur tabulée si disponible
	if capacity, ok := capacityTable[version]; ok {
		return capacity
	}

	// Fallback : calculer en comptant les positions valides
	capacity := 0
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			if isValidDataPosition(x, y, size) {
				capacity++
			}
		}
	}
	return capacity
}

// Fonction existante utilisée pour la correction d'erreur
func AddErrorCorrection(data string, level string, version int) string {
	// Utilise la fonction existante dans error_correction.go
	return AddErrorCorrectionEC(data, level, version)
}
