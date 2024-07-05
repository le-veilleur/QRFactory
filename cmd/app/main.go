package main

import (
    "log"
    "QRFactory/internal/app"
)

func main() {
    application := app.New()
    if err := application.Run(); err != nil {
        log.Fatalf("could not run application: %v", err)
    }
}
