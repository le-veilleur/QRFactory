package qrcode

import (
    "github.com/skip2/go-qrcode"
)

func Generate(content string) ([]byte, error) {
    return qrcode.Encode(content, qrcode.Medium, 256)
}
