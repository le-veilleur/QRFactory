package main

import (
	"fmt"
	"os"
	"qrfactory/internal/model"
	"qrfactory/pkg/config"
	"qrfactory/pkg/qr"
	"time"

	"github.com/spf13/cobra"
)

var cfg *config.QRConfig
var scale int

var rootCmd = &cobra.Command{
	Use:   "qrfactory",
	Short: "QRFactory - Générateur de codes QR",
	Long: `QRFactory est un outil en ligne de commande pour générer des codes QR.
Il supporte différents types de données et niveaux de correction d'erreur.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Démarrage du programme...")
		start := time.Now()

		// Validation de la configuration
		fmt.Println("Validation de la configuration...")
		if err := config.ValidateConfig(cfg); err != nil {
			fmt.Printf("Erreur de configuration : %v\n", err)
			os.Exit(1)
		}

		// Initialisation du champ de Galois
		fmt.Println("Initialisation du champ de Galois...")
		qr.InitGaloisField()
		fmt.Printf("Initialisation terminée en %v\n", time.Since(start))

		// Création d'un modèle QRCode
		fmt.Println("Création du modèle QRCode...")
		qrCode := model.NewQRCode(cfg.Data, cfg.Version, cfg.ErrorCorrectionLevel)

		// Détection du type de données
		fmt.Println("Détection du type de données...")
		dataType := qr.DetectDataType(cfg.Data)
		fmt.Printf("Type de données détecté: %s\n", dataType)

		// Calcul de la version minimale nécessaire
		fmt.Println("Calcul de la version minimale nécessaire...")
		minVersion, err := qr.CalculateMinVersionForDataType(cfg.Data, dataType)
		if err != nil {
			fmt.Printf("Erreur lors du calcul de la version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Version minimale calculée: %d\n", minVersion)

		// Utilisation de la version minimale si nécessaire
		if cfg.Version < minVersion {
			fmt.Printf("La version %d est trop petite pour les données. Utilisation de la version %d.\n", cfg.Version, minVersion)
			cfg.Version = minVersion
			qrCode.Version = minVersion
			qrCode.Size = minVersion*4 + 17
		}

		// Génération du QR code
		fmt.Println("Génération de la matrice QR...")
		genStart := time.Now()
		matrix := qr.GenerateQRMatrix(cfg.Version, cfg.Data, cfg.ErrorCorrectionLevel)
		fmt.Printf("Génération de la matrice terminée en %v\n", time.Since(genStart))

		if matrix == nil {
			fmt.Println("Erreur lors de la génération du QR code")
			os.Exit(1)
		}

		// Association de la matrice au modèle
		fmt.Println("Association de la matrice au modèle...")
		qrCode.SetMatrix(matrix)

		// Sauvegarde de l'image
		fmt.Println("Sauvegarde de l'image...")
		saveStart := time.Now()
		if err := qr.SaveQRImage(matrix, cfg.OutputFile, scale); err != nil {
			fmt.Printf("Erreur lors de la sauvegarde de l'image : %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Sauvegarde terminée en %v\n", time.Since(saveStart))

		fmt.Printf("QR code généré avec succès dans %s\n", cfg.OutputFile)
		fmt.Printf("Temps total d'exécution: %v\n", time.Since(start))
	},
}

func init() {
	// Création d'une configuration par défaut
	cfg = config.NewDefaultConfig()

	// Configuration des flags de la commande
	rootCmd.Flags().StringVarP(&cfg.Data, "data", "d", "", "Données à encoder dans le QR code")
	rootCmd.Flags().IntVarP(&cfg.Version, "version", "v", cfg.Version, "Version du QR code (1-40)")
	rootCmd.Flags().StringVarP(&cfg.ErrorCorrectionLevel, "error-correction", "e", "M", "Niveau de correction d'erreur (L, M, Q, H)")
	rootCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "qrcode.png", "Chemin du fichier de sortie")
	rootCmd.Flags().StringVar(&cfg.BackgroundColor, "bg-color", cfg.BackgroundColor, "Couleur de fond")
	rootCmd.Flags().StringVar(&cfg.ForegroundColor, "fg-color", cfg.ForegroundColor, "Couleur des modules")
	rootCmd.Flags().IntVarP(&scale, "scale", "s", 20, "Échelle de l'image (défaut: 20)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
