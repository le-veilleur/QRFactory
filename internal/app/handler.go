package app

import (
    "encoding/json"
    "net/http"
    "QRFactory/internal/qrcode"  // Assurez-vous que le chemin est correct
)

type Server struct {
    // Ajoutez ici des champs si nécessaire (ex : configuration, logger, etc.)
}

type QRCodeRequest struct {
    Content string `json:"content"`
}

func (s *Server) handleGenerateQRCode() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req QRCodeRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        png, err := qrcode.Generate(req.Content)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "image/png")
        w.Write(png)
    }
}
