package app

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "QRFactory/internal/config"
)

type App struct {
    router *mux.Router
    server *Server
}

func New() *App {
    return &App{
        router: mux.NewRouter(),
        server: &Server{},
    }
}

func (a *App) Run() error {
    a.routes()
    cfg := config.LoadConfig()
    log.Printf("Starting server on %s...", cfg.ServerAddress)
    return http.ListenAndServe(cfg.ServerAddress, a.router)
}

func (a *App) routes() {
    a.router.HandleFunc("/qrcode", a.server.handleGenerateQRCode()).Methods("POST")
}
