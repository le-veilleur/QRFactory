package model


type QRCode struct {
    Size  int    // Taille du QR code
    Data  string // Données à encoder
    Image []byte // Image générée du QR code
}

