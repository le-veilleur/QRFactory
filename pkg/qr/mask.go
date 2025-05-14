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
				r, _, _, _ := matrix.At(x, y).RGBA()
				if r > 0 { // Si c'est blanc (255, 255, 255)
					maskedMatrix.Set(x, y, color.Black)
				} else { // Si c'est noir (0, 0, 0)
					maskedMatrix.Set(x, y, color.White)
				}
			}
		}
	}

	return maskedMatrix
}

// isFunctionPattern vérifie si une position donnée fait partie d'un motif de fonction
func isFunctionPattern(x, y, size int) bool {
	// Motifs de repérage (Finder Patterns)
	if (x < 8 && y < 8) || // Coin supérieur gauche
		(x > size-9 && y < 8) || // Coin supérieur droit
		(x < 8 && y > size-9) { // Coin inférieur gauche
		return true
	}

	// Motifs de synchronisation (Timing Patterns)
	if x == 6 || y == 6 {
		return true
	}

	// Motifs d'alignement (Alignment Patterns)
	if size > 21 { // Version > 1
		alignPos := GetAlignmentPatternPositions(size)
		for _, pos := range alignPos {
			if abs(x-pos.x) <= 2 && abs(y-pos.y) <= 2 {
				return true
			}
		}
	}

	return false
}

// GetAlignmentPatternPositions calcule les positions des motifs d'alignement
func GetAlignmentPatternPositions(size int) []struct{ x, y int } {
	version := (size - 17) / 4
	if version <= 1 {
		return nil
	}

	positions := []struct{ x, y int }{}
	interval := version/7 + 2
	last := size - 7

	for i := 6; i <= last; i += interval {
		positions = append(positions, struct{ x, y int }{i, 6})
		positions = append(positions, struct{ x, y int }{6, i})
		for j := 6; j <= last; j += interval {
			positions = append(positions, struct{ x, y int }{i, j})
		}
	}

	return positions
}
