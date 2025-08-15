#!/bin/bash

set -e

echo "Mode développement LiteFaaS"

# Vérifier que Go est installé
if ! command -v go &> /dev/null; then
    echo "Erreur: Go n'est pas installé"
    exit 1
fi

# Vérifier que containerd est installé
if ! command -v containerd &> /dev/null; then
    echo "Attention: containerd n'est pas installé. Installation..."
    case $(uname -s) in
        Linux)
            if [ -f /etc/debian_version ]; then
                sudo apt-get update && sudo apt-get install -y containerd.io
            elif [ -f /etc/redhat-release ]; then
                if command -v dnf &> /dev/null; then
                    sudo dnf install -y containerd
                else
                    sudo yum install -y containerd
                fi
            fi
            ;;
        Darwin)
            brew install containerd
            ;;
    esac
fi

# Créer le répertoire de données pour le développement
mkdir -p /tmp/litefaas-dev
export LITEFAAS_DB_PATH="/tmp/litefaas-dev/functions.db"

# Télécharger les dépendances
echo "Téléchargement des dépendances..."
go mod download

# Compiler en mode développement
echo "Compilation en mode développement..."
go build -o bin/litefaas-dev cmd/litefaas/main.go

# Démarrer containerd si nécessaire
if command -v systemctl &> /dev/null; then
    if ! systemctl is-active --quiet containerd; then
        echo "Démarrage de containerd..."
        sudo systemctl start containerd
    fi
fi

echo "Démarrage de LiteFaaS en mode développement..."
echo "Base de données: $LITEFAAS_DB_PATH"
echo "Interface web: http://localhost:8080/web/"
echo "API: http://localhost:8080/api/"

./bin/litefaas-dev --db-path "$LITEFAAS_DB_PATH" --log-level debug
