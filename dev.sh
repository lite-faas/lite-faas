#!/bin/bash

echo "=== Démarrage LiteFaaS en mode développement ==="

# Vérifier si Go est installé
if ! command -v go &> /dev/null; then
    echo "Erreur: Go n'est pas installé"
    exit 1
fi

# Installer les dépendances
echo "Installation des dépendances..."
go mod download
go mod tidy

# Construire l'application
echo "Construction de l'application..."
go build -o litefaas cmd/litefaas/main.go

# Démarrer en mode développement
echo "Démarrage du serveur..."
./litefaas --dev --db-path ./dev.db --port 8080 --log-format text --log-level debug
