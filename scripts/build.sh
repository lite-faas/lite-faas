#!/bin/bash

set -e

echo "Compilation locale de LiteFaaS..."

# Vérifier que Go est installé
if ! command -v go &> /dev/null; then
    echo "❌ Go n'est pas installé."
    echo "   Installez Go depuis: https://golang.org/dl/"
    exit 1
fi

# Vérifier que nous sommes dans le bon répertoire
if [ ! -f "go.mod" ]; then
    echo "❌ Fichier go.mod non trouvé."
    echo "   Assurez-vous d'être dans le répertoire du projet LiteFaaS."
    exit 1
fi

# Créer le répertoire bin
mkdir -p bin

# Télécharger les dépendances
echo "Téléchargement des dépendances..."
go mod download

# Compiler
echo "Compilation de LiteFaaS..."
go build -ldflags="-s -w" -o bin/litefaas cmd/litefaas/main.go

# Vérifier que le binaire a été créé
if [ -f "bin/litefaas" ]; then
    echo "✅ Compilation réussie!"
    echo "   Binaire créé: bin/litefaas"
    echo "   Taille: $(du -h bin/litefaas | cut -f1)"
else
    echo "❌ Échec de la compilation"
    exit 1
fi
