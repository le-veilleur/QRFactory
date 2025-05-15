package qr

import (
	"image"
	"image/color"
)

// ApplyMask applique un masque à la matrice QR et retourne la matrice masquée
func ApplyMask(matrix *image.RGBA, maskPattern int) *image.RGBA {
	size := matrix.Bounds().Max.X
	maskedMatrix := image.NewRGBA(matrix.Bounds())

	// Copie d'abord la matrice originale
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			maskedMatrix.Set(x, y, matrix.At(x, y))
		}
	}

	// Applique le masque uniquement aux données
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Ne pas masquer les motifs de fonction
			if isFunctionPattern(x, y, size) {
				continue
			}

			// Détermine si le pixel doit être inversé selon le motif de masque
			// Définition officielle des masques selon le standard ISO/IEC 18004:2015
			shouldInvert := false
			switch maskPattern {
			case 0: // (x+y) mod 2 == 0
				shouldInvert = (x+y)%2 == 0
			case 1: // y mod 2 == 0
				shouldInvert = y%2 == 0
			case 2: // x mod 3 == 0
				shouldInvert = x%3 == 0
			case 3: // (x+y) mod 3 == 0
				shouldInvert = (x+y)%3 == 0
			case 4: // (floor(y/2) + floor(x/3)) mod 2 == 0
				shouldInvert = ((y/2)+(x/3))%2 == 0
			case 5: // (x*y) mod 2 + (x*y) mod 3 == 0
				shouldInvert = ((x*y)%2 + (x*y)%3) == 0
			case 6: // ((x*y) mod 2 + (x*y) mod 3) mod 2 == 0
				shouldInvert = ((x*y)%2+(x*y)%3)%2 == 0
			case 7: // ((x+y) mod 2 + (x*y) mod 3) mod 2 == 0
				shouldInvert = ((x+y)%2+(x*y)%3)%2 == 0
			}

			if shouldInvert {
				// Inverser la couleur
				// Vérifier la couleur actuelle
				r, _, _, a := maskedMatrix.At(x, y).RGBA()
				if r > 0x7FFF { // Si c'est blanc (valeurs normalisées sont sur 16 bits)
					maskedMatrix.Set(x, y, color.RGBA{0, 0, 0, uint8(a >> 8)})
				} else { // Si c'est noir
					maskedMatrix.Set(x, y, color.RGBA{255, 255, 255, uint8(a >> 8)})
				}
			}
		}
	}

	return maskedMatrix
}

// isFunctionPattern vérifie si une position donnée fait partie d'un motif de fonction
// Corrigé pour mieux identifier tous les motifs de fonction
func isFunctionPattern(x, y, size int) bool {
	// Motifs de repérage (Finder Patterns) + séparateurs
	if (x < 9 && y < 9) || // Coin supérieur gauche + séparateur
		(x >= size-8 && y < 9) || // Coin supérieur droit + séparateur
		(x < 9 && y >= size-8) { // Coin inférieur gauche + séparateur
		return true
	}

	// Zone d'information de format (Format Information)
	if (x < 9 && y == 8) || (x == 8 && y < 9) ||
		(x == 8 && y >= size-8) || (y == 8 && x >= size-8) {
		return true
	}

	// Motifs de synchronisation (Timing Patterns)
	if x == 6 || y == 6 {
		return true
	}

	// Version 7+ a des bits d'information de version
	version := (size - 17) / 4
	if version >= 7 {
		// Zone d'information de version près du finder pattern supérieur droit
		if x >= size-11 && x <= size-9 && y < 6 {
			return true
		}
		// Zone d'information de version près du finder pattern inférieur gauche
		if y >= size-11 && y <= size-9 && x < 6 {
			return true
		}
	}

	// Motifs d'alignement (Alignment Patterns)
	if version >= 2 { // Version >= 2
		alignPositions := getAlignmentPatternCoordinates(version)

		for _, posX := range alignPositions {
			for _, posY := range alignPositions {
				// Éviter les collisions avec les motifs de repérage
				if !((posX <= 8 && posY <= 8) ||
					(posX >= size-8 && posY <= 8) ||
					(posX <= 8 && posY >= size-8)) {

					// Vérifier si le point actuel est dans un motif d'alignement
					if abs(x-posX) <= 2 && abs(y-posY) <= 2 {
						return true
					}
				}
			}
		}
	}

	// Motif sombre (Dark Module) - toujours présent à cette position spécifique
	if x == 8 && y == size-8 {
		return true
	}

	return false
}

// getAlignmentPatternCoordinates retourne les coordonnées des motifs d'alignement
// selon le tableau de référence de la spécification ISO/IEC 18004
func getAlignmentPatternCoordinates(version int) []int {
	switch version {
	case 1:
		return []int{}
	case 2:
		return []int{6, 18}
	case 3:
		return []int{6, 22}
	case 4:
		return []int{6, 26}
	case 5:
		return []int{6, 30}
	case 6:
		return []int{6, 34}
	case 7:
		return []int{6, 22, 38}
	case 8:
		return []int{6, 24, 42}
	case 9:
		return []int{6, 26, 46}
	case 10:
		return []int{6, 28, 50}
	default:
		// Pour les versions supérieures, calculer approximativement
		step := 4 + (version / 7)
		result := []int{6}

		for pos := 6 + step; pos < version*4+10; pos += step {
			result = append(result, pos)
		}

		return result
	}
}

// GetAlignmentPatternPositions calcule les positions des motifs d'alignement
// Cette fonction est utilisée ailleurs, nous la gardons pour la compatibilité
func GetAlignmentPatternPositions(size int) []struct{ x, y int } {
	version := (size - 17) / 4
	if version <= 1 {
		return nil
	}

	coords := getAlignmentPatternCoordinates(version)
	positions := []struct{ x, y int }{}

	for _, x := range coords {
		for _, y := range coords {
			// Éviter les collisions avec les motifs de repérage
			if !((x <= 8 && y <= 8) ||
				(x >= size-8 && y <= 8) ||
				(x <= 8 && y >= size-8)) {
				positions = append(positions, struct{ x, y int }{x, y})
			}
		}
	}

	return positions
}
