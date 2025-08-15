#!/bin/bash

set -e

echo "Installation de LiteFaaS..."

OS=$(uname -s)
ARCH=$(uname -m)

case $OS in
    Linux)
        echo "Détection: Linux"
        ;;
    Darwin)
        echo "Détection: macOS"
        ;;
    *)
        echo "Système d'exploitation non supporté: $OS"
        exit 1
        ;;
esac

# Installation de containerd
echo "Installation de containerd..."

case $OS in
    Linux)
        if ! command -v containerd &> /dev/null; then
            echo "Téléchargement et installation de containerd..."

            # Détection de la distribution
            if [ -f /etc/debian_version ]; then
                # Debian/Ubuntu
                sudo apt-get update
                sudo apt-get install -y containerd.io
            elif [ -f /etc/redhat-release ]; then
                # RHEL/CentOS/Fedora
                if command -v dnf &> /dev/null; then
                    sudo dnf install -y containerd
                else
                    sudo yum install -y containerd
                fi
            else
                # Installation manuelle
                CONTAINERD_VERSION="1.7.0"
                wget https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
                sudo tar -C /usr/local -xzf containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
                sudo systemctl enable containerd
                sudo systemctl start containerd
                rm containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
            fi
        else
            echo "containerd est déjà installé"
        fi
        ;;
    Darwin)
        if ! command -v containerd &> /dev/null; then
            echo "Installation de containerd via Homebrew..."
            brew install containerd
        else
            echo "containerd est déjà installé"
        fi
        ;;
esac

# Vérification de containerd
if ! command -v containerd &> /dev/null; then
    echo "Erreur: containerd n'a pas pu être installé"
    exit 1
fi

echo "containerd installé avec succès"

# Téléchargement du binaire LiteFaaS
echo "Téléchargement du binaire LiteFaaS..."

# Déterminer la version à télécharger
if [ "$1" = "dev" ]; then
    echo "Téléchargement de la version de développement..."
    LATEST_VERSION="dev"
else
    LATEST_VERSION=$(curl -s https://api.github.com/repos/litefaas/litefaas/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        echo "Impossible de déterminer la dernière version, utilisation de la version de développement..."
        LATEST_VERSION="dev"
    fi
fi

# Déterminer l'architecture
case $ARCH in
    x86_64)
        BINARY_ARCH="amd64"
        ;;
    aarch64|arm64)
        BINARY_ARCH="arm64"
        ;;
    *)
        echo "Architecture non supportée: $ARCH"
        exit 1
        ;;
esac

# Déterminer le système d'exploitation pour le nom du binaire
case $OS in
    Linux)
        BINARY_OS="linux"
        ;;
    Darwin)
        BINARY_OS="darwin"
        ;;
esac

# Créer le répertoire bin
mkdir -p bin

# Télécharger le binaire
BINARY_NAME="litefaas-${BINARY_OS}-${BINARY_ARCH}"
if [ "$LATEST_VERSION" = "dev" ]; then
    echo "Téléchargement de la version de développement..."
    # Pour la version de développement, on compile localement
    if ! command -v go &> /dev/null; then
        echo "Installation de Go..."
        case $OS in
            Linux)
                wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
                sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
                export PATH=$PATH:/usr/local/go/bin
                rm go1.21.0.linux-amd64.tar.gz
                ;;
            Darwin)
                brew install go
                ;;
        esac
    fi

    echo "Compilation de LiteFaaS..."
    go mod download
    go build -o bin/litefaas cmd/litefaas/main.go
else
    echo "Téléchargement de la version $LATEST_VERSION..."
    DOWNLOAD_URL="https://github.com/litefaas/litefaas/releases/download/${LATEST_VERSION}/${BINARY_NAME}"

    if curl -L -o bin/litefaas "$DOWNLOAD_URL"; then
        chmod +x bin/litefaas
        echo "Binaire téléchargé avec succès"
    else
        echo "Échec du téléchargement, compilation locale..."
        if ! command -v go &> /dev/null; then
            echo "Installation de Go..."
            case $OS in
                Linux)
                    wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
                    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
                    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
                    export PATH=$PATH:/usr/local/go/bin
                    rm go1.21.0.linux-amd64.tar.gz
                    ;;
                Darwin)
                    brew install go
                    ;;
            esac
        fi

        echo "Compilation de LiteFaaS..."
        go mod download
        go build -o bin/litefaas cmd/litefaas/main.go
    fi
fi

# Création des répertoires nécessaires
echo "Création des répertoires nécessaires..."
sudo mkdir -p /var/lib/litefaas
sudo chown $USER:$USER /var/lib/litefaas

# Configuration de containerd
echo "Configuration de containerd..."
sudo mkdir -p /etc/containerd
if [ ! -f /etc/containerd/config.toml ]; then
    containerd config default | sudo tee /etc/containerd/config.toml > /dev/null
fi

# Redémarrer containerd pour appliquer la configuration
if command -v systemctl &> /dev/null; then
    sudo systemctl restart containerd
fi

echo "Installation terminée!"
echo "Pour démarrer LiteFaaS: ./bin/litefaas"
