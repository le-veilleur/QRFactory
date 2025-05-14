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
func CalculateMinVersion(data string) int {
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
			return version
		}
	}

	return -1 // Impossible de stocker les données même avec la version maximale
}

// GenerateQRMatrix génère la matrice QR pour les données fournies
func GenerateQRMatrix(version int, data string) *image.RGBA {
	// Calculer la version minimale nécessaire
	minVersion := CalculateMinVersion(data)
	if minVersion == -1 {
		fmt.Printf("ERREUR: Les données sont trop longues même pour la version maximale\n")
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

	// Calculer la taille totale nécessaire en bits
	totalBits := 4 + 8 + (len(data) * 8) // Mode (4) + Length (8) + Data (8 par caractère)

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
	if capacity, ok := capacityTable[version]; ok {
		if capacity < totalBits {
			fmt.Printf("ERREUR: Les données (%d bits) dépassent la capacité de la version %d (%d bits)\n",
				totalBits, version, capacity)
			return nil
		}
	}

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

	// Préparer les données avec l'indicateur de mode et la longueur
	var encodedData strings.Builder

	// Ajouter l'indicateur de mode (0100 pour mode byte)
	encodedData.WriteString("0100")

	// Ajouter la longueur des données (8 bits pour version 1-9 en mode byte)
	lengthBits := fmt.Sprintf("%08b", len(data))
	encodedData.WriteString(lengthBits)

	// Encoder les données en mode byte
	byteData, err := EncodeByte(data)
	if err != nil {
		fmt.Printf("Erreur d'encodage: %v\n", err)
		return nil
	}
	encodedData.WriteString(byteData)

	// Calculer la capacité disponible
	capacity := calculateAvailableCapacity(size)

	// Vérifier si les données encodées dépassent la capacité
	if encodedData.Len() > capacity {
		fmt.Printf("ERREUR: Les données encodées (%d bits) dépassent la capacité (%d bits)\n", encodedData.Len(), capacity)
		return nil
	}

	// Ajouter le terminateur
	for encodedData.Len() < capacity && encodedData.Len() < capacity-4 {
		encodedData.WriteString("0")
	}

	// Ajouter la correction d'erreur
	finalData := AddErrorCorrection(encodedData.String(), "M", version)

	fmt.Printf("Capacité disponible: %d bits\n", capacity)
	fmt.Printf("Longueur des données encodées: %d bits\n", len(finalData))
	fmt.Printf("Données encodées: %s\n", finalData)

	// Placer les données
	PlaceData(matrix, finalData)

	// Appliquer le meilleur masque
	bestScore := math.MaxInt32
	var bestMatrix *image.RGBA
	bestMask := 0

	for mask := 0; mask < 8; mask++ {
		maskedMatrix := ApplyMask(matrix, mask)
		score := EvaluateMask(maskedMatrix)

		if score < bestScore {
			bestScore = score
			bestMatrix = maskedMatrix
			bestMask = mask
		}
	}

	if bestMatrix != nil {
		matrix = bestMatrix
		AddFormatInfo(matrix, "M", bestMask)
	}

	return matrix
}

// EvaluateMask évalue la qualité d'un masque selon les règles de pénalité du QR code
func EvaluateMask(matrix *image.RGBA) int {
	score := 0
	size := matrix.Bounds().Max.X

	// Règle 1: Groupes de modules de même couleur
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if x+4 < size {
				// Vérifier horizontalement
				count := 1
				color := matrix.At(x, y)
				for i := 1; i < 5; i++ {
					if matrix.At(x+i, y) == color {
						count++
					}
				}
				if count >= 5 {
					score += count - 2
				}
			}
			if y+4 < size {
				// Vérifier verticalement
				count := 1
				color := matrix.At(x, y)
				for i := 1; i < 5; i++ {
					if matrix.At(x, y+i) == color {
						count++
					}
				}
				if count >= 5 {
					score += count - 2
				}
			}
		}
	}

	return score
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

	// Commencer par le coin en bas à droite
	for x := size - 1; x >= 0 && dataIndex < len(data); x -= 2 {
		// Traiter deux colonnes à la fois
		for dx := 0; dx <= 1 && x-dx >= 0; dx++ {
			currentX := x - dx

			// Sauter la colonne de timing
			if currentX == 6 {
				continue
			}

			if upward {
				// Monter
				for y := size - 1; y >= 0 && dataIndex < len(data); y-- {
					if isValidDataPosition(currentX, y, size) {
						if dataIndex < len(data) {
							fmt.Printf("Module placé à (%d,%d): %c [Total: %d]\n",
								currentX, y, data[dataIndex], dataIndex+1)
							if data[dataIndex] == '1' {
								matrix.Set(currentX, y, color.Black)
							} else {
								matrix.Set(currentX, y, color.White)
							}
							dataIndex++
						}
					}
				}
			} else {
				// Descendre
				for y := 0; y < size && dataIndex < len(data); y++ {
					if isValidDataPosition(currentX, y, size) {
						if dataIndex < len(data) {
							fmt.Printf("Module placé à (%d,%d): %c [Total: %d]\n",
								currentX, y, data[dataIndex], dataIndex+1)
							if data[dataIndex] == '1' {
								matrix.Set(currentX, y, color.Black)
							} else {
								matrix.Set(currentX, y, color.White)
							}
							dataIndex++
						}
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
	for _, ax := range alignmentPositions {
		for _, ay := range alignmentPositions {
			if x >= ax-2 && x <= ax+2 && y >= ay-2 && y <= ay+2 {
				return false
			}
		}
	}

	return true
}

// getAlignmentPositions calcule les positions des motifs d'alignement
func getAlignmentPositions(size int) []int {
	version := (size - 17) / 4
	if version < 2 {
		return []int{}
	}

	// Table des positions des motifs d'alignement pour les versions 2-7
	positions := map[int][]int{
		2: {6, 18},
		3: {6, 22},
		4: {6, 26},
		5: {6, 30},
		6: {6, 34},
		7: {6, 22, 38},
	}

	if pos, ok := positions[version]; ok {
		return pos
	}
	return []int{}
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
