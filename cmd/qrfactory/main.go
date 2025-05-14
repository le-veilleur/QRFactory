package main

import (
	"qrfactory/pkg/qr"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	data            string
	version         int
	errorLevel      string
	outputFile      string
	backgroundColor string
	foregroundColor string
	scale           int
)

var rootCmd = &cobra.Command{
	Use:   "qrfactory",
	Short: "QRFactory - Générateur de codes QR",
	Long: `QRFactory est un outil en ligne de commande pour générer des codes QR.
Il supporte différents types de données et niveaux de correction d'erreur.`,
	Run: func(cmd *cobra.Command, args []string) {
		if data == "" {
			fmt.Println("Erreur : Veuillez spécifier les données à encoder (-d ou --data)")
			os.Exit(1)
		}

		// Initialisation du champ de Galois
		qr.InitGaloisField()

		// Génération du QR code
		matrix := qr.GenerateQRMatrix(version, data)
		if matrix == nil {
			fmt.Println("Erreur lors de la génération du QR code")
			os.Exit(1)
		}

		// Sauvegarde de l'image
		if err := qr.SaveQRImage(matrix, outputFile, scale); err != nil {
			fmt.Printf("Erreur lors de la sauvegarde de l'image : %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("QR code généré avec succès dans %s\n", outputFile)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&data, "data", "d", "", "Données à encoder dans le QR code")
	rootCmd.Flags().IntVarP(&version, "version", "v", 1, "Version du QR code (1-40)")
	rootCmd.Flags().StringVarP(&errorLevel, "error-level", "e", "L", "Niveau de correction d'erreur (L, M, Q, H)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "qrcode.png", "Nom du fichier de sortie")
	rootCmd.Flags().StringVar(&backgroundColor, "bg-color", "white", "Couleur de fond")
	rootCmd.Flags().StringVar(&foregroundColor, "fg-color", "black", "Couleur des modules")
	rootCmd.Flags().IntVarP(&scale, "scale", "s", 10, "Facteur d'échelle pour l'image de sortie")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
