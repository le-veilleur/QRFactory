# QRFactory

QRFactory est une application Go en cours de développement, conçue pour générer des codes QR en respectant la norme ISO/IEC 18004. Ce projet utilise une architecture modulaire et une approche de développement pilotée par les tests (TDD) pour garantir un code robuste et maintenable.

## Table des matières

- [Installation](#installation)
- [Utilisation](#utilisation)
- [Architecture](#architecture)
- [Tests](#tests)
- [Contribuer](#contribuer)
- [État du développement](#état-du-développement)
- [Licence](#licence)

## Installation

1. **Cloner le dépôt :**

   ```sh
   git clone https://github.com/votre-utilisateur/QRFactory.git
   cd QRFactory
   ```

2. **Initialiser les modules Go :**

   ```sh
   go mod tidy
   ```

## Utilisation

Pour générer un code QR, exécutez la commande suivante :

```sh
go run cmd/qrfactory/main.go
```

### Exemple

Pour générer un code QR avec le texte "HELLO WORLD" :

1. Ouvrez `cmd/qrfactory/main.go` et modifiez le contenu comme suit :

    ```go
    package main

    import (
        "QRFactory/pkg/qr"
        "log"
    )

    func main() {
        err := qr.GenerateQRCode("HELLO WORLD", 1, "L", "qrcode.png")
        if err != nil {
            log.Fatalf("Failed to generate QR code: %v", err)
        }
    }
    ```

2. Exécutez le programme :

    ```sh
    go run cmd/qrfactory/main.go
    ```

    Cela générera un fichier `qrcode.png` dans le répertoire courant.

## Architecture

Le projet est structuré comme suit :

```
/QRFactory
│
├── cmd/
│   └── qrfactory/
│       └── main.go
│
├── internal/
│   │
│   ├── model/
│       ├── qr_code.go
│       └── qr_code_test.go
│
├── pkg/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   │
│   └── qr/
│       ├── generator.go
│       └── generator_test.go
│
├── go.mod
└── go.sum
```

- **cmd/** : Contient l'application principale.
- **internal/** : Contient la logique métier et les handlers API.
- **pkg/** : Contient les packages réutilisables, y compris la logique de génération des QR codes.

## Tests

Les tests sont écrits en utilisant le package de test standard de Go. Pour exécuter les tests, utilisez la commande suivante :

```sh
go test ./...
```

### Exemple de test

Un test d'encodage numérique :

```go
package qr

import "testing"

func TestEncodeNumeric(t *testing.T) {
    data := "1234567890"
    expected := "00010000001100010000110100110000001100011000110100"
    result := EncodeNumeric(data)
    if result != expected {
        t.Errorf("Expected %s but got %s", expected, result)
    }
}
```

## Contribuer

Les contributions sont les bienvenues ! Veuillez suivre les étapes suivantes pour contribuer :

1. Forkez le dépôt.
2. Créez une branche pour votre fonctionnalité (`git checkout -b feature/ma-nouvelle-fonctionnalité`).
3. Commitez vos modifications (`git commit -am 'Ajoute une nouvelle fonctionnalité'`).
4. Poussez votre branche (`git push origin feature/ma-nouvelle-fonctionnalité`).
5. Créez une Pull Request.

## État du développement

Ce projet est en cours de développement. Voici les fonctionnalités actuellement implémentées :

- [x] Encodage numérique
- [x] Encodage alphanumérique
- [x] Encodage byte
- [x] Encodage Kanji
- [ ] Génération d'image QR code
- [ ] Interface utilisateur (API ou CLI)

Nous travaillons activement sur l'ajout de nouvelles fonctionnalités et l'amélioration des fonctionnalités existantes.

## Licence

Ce projet est sous licence MIT. Voir le fichier [LICENSE](LICENSE) pour plus de détails.

Ce `README.md` reflète maintenant l'état de développement en cours du projet QRFactory et indique les fonctionnalités déjà implémentées et celles qui restent à développer.
Cela permet aux contributeurs et aux utilisateurs de mieux comprendre où en est le projet et ce qu'il reste à faire.
