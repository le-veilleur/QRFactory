package qr

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// TestEncodeAlphanumeric teste la fonction EncodeAlphanumeric pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne alphanumérique donnée.
func TestEncodeAlphanumeric(t *testing.T) {
	// Données d'entrée pour le test
	data := "TEST MAXIME"
	// Résultat attendu pour les données d'entrée
	expected := "1010010011110100001001110011010100011110001101101000000001110"

	// Appel de la fonction à tester
	fmt.Println("Testing EncodeAlphanumeric with input:", data)
	result, err := EncodeAlphanumeric(data)

	// Vérification des erreurs potentielles
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Comparaison du résultat obtenu avec le résultat attendu (en supprimant les espaces)
	if strings.ReplaceAll(result, " ", "") != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	} else {
		fmt.Println("Test passed! Result:", result)
	}
}

// TestEncodeByte teste la fonction EncodeByte pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne de caractères donnée.
func TestEncodeByte(t *testing.T) {
	// Données d'entrée pour le test
	data := "HELLO"
	// Résultat attendu pour les données d'entrée
	expected := "0100100001000101010011000100110001001111"

	// Appel de la fonction à tester
	result, err := EncodeByte(data)

	// Vérification des erreurs potentielles
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Comparaison du résultat obtenu avec le résultat attendu
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

// TestEncodeNumeric teste la fonction EncodeNumeric pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne numérique donnée.
func TestEncodeNumeric(t *testing.T) {
	// Données d'entrée pour le test
	data := "123456890"
	// Résultat attendu pour les données d'entrée
	expected := "000111101101110010001101111010"

	// Appel de la fonction à tester
	result, err := EncodeNumeric(data)

	// Vérification des erreurs potentielles
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Comparaison du résultat obtenu avec le résultat attendu
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

// TestEncodeKanji teste la fonction EncodeKanji pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne Kanji donnée.
func TestEncodeKanji(t *testing.T) {
	// Données d'entrée pour le test (chaîne Kanji)
	data := "世界"
	// Résultat attendu pour les données d'entrée
	expected := "00000000000101100010011101001100" // Remplace ceci par le binaire correct après avoir vérifié l'encodage exact

	// Appel de la fonction à tester
	result, err := EncodeKanji(data)

	// Vérification des erreurs potentielles
	if err != nil {
		t.Fatalf("Erreur inattendue: %v", err)
	}

	// Comparaison du résultat obtenu avec le résultat attendu
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

// TestToInt teste la fonction toInt pour s'assurer qu'elle convertit correctement une chaîne en entier.
func TestToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		name     string
		isError  bool // Indique si on s'attend à une erreur
	}{
		{"123", 123, "Nombre positif", false},
		{"-456", -456, "Nombre négatif", false},
		{"0", 0, "Zéro", false},
		{"999999999", 999999999, "Nombre très grand", false},
		{"-999999999", -999999999, "Nombre très petit", false},
		{"abc", 0, "Chaîne non numérique (doit retourner 0)", true}, // Attend une erreur
		{"2147483647", math.MaxInt32, "Plus grand nombre entier 32 bits", false},
		{"-2147483648", math.MinInt32, "Plus petit nombre entier 32 bits", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toInt(tt.input)

			// Vérification des erreurs
			if tt.isError {
				if err == nil {
					t.Error("Expected error for non-numeric input, but got nil")
				}
				return // Sortir du test si on attend une erreur
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("For input %s, expected %d but got %d", tt.input, tt.expected, result)
			}
		})
	}
}
