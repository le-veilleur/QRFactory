package main

import (
	"QRFactory/pkg/qr" // Assurez-vous que c'est le bon chemin
	"fmt"
)

func main() {
	qr.InitGaloisField()

	data := []byte{ /* tes données */ }
	level := "M"
	ecBytes := qr.GenerateErrorCorrection(data, level)

	fmt.Printf("Syndromes Reed-Solomon : %v\n", ecBytes)
}
