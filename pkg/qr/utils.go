package qr

import (
	"strconv"
)

// Min renvoie le minimum de deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ToInt convertit une chaîne en entier
func ToInt(s string) (int, error) {
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Abs renvoie la valeur absolue d'un entier
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
