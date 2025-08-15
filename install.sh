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
            if [ -f /etc/os-release ]; then
                . /etc/os-release
                DISTRO=$ID
                VERSION=$VERSION_ID
            elif [ -f /etc/debian_version ]; then
                DISTRO="debian"
            elif [ -f /etc/redhat-release ]; then
                DISTRO="rhel"
            else
                DISTRO="unknown"
            fi

            case $DISTRO in
                ubuntu|debian)
                    echo "Distribution détectée: $DISTRO"
                    # Ajouter le repository Docker officiel
                    sudo apt-get update
                    sudo apt-get install -y ca-certificates curl gnupg lsb-release
                    sudo mkdir -p /etc/apt/keyrings
                    curl -fsSL https://download.docker.com/linux/$DISTRO/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
                    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$DISTRO $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
                    sudo apt-get update
                    sudo apt-get install -y containerd.io
                    ;;
                alpine)
                    echo "Distribution détectée: Alpine Linux"
                    # Alpine utilise apk
                    sudo apk update
                    sudo apk add containerd
                    # Créer le répertoire de configuration
                    sudo mkdir -p /etc/containerd
                    ;;
                rhel|centos|fedora|rocky|almalinux)
                    echo "Distribution détectée: $DISTRO"
                    if command -v dnf &> /dev/null; then
                        sudo dnf install -y containerd
                    else
                        sudo yum install -y containerd
                    fi
                    ;;
                *)
                    echo "Distribution non reconnue: $DISTRO, installation manuelle..."
                    # Installation manuelle
                    CONTAINERD_VERSION="1.7.0"
                    wget https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
                    sudo tar -C /usr/local -xzf containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
                    sudo systemctl enable containerd
                    sudo systemctl start containerd
                    rm containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
                    ;;
            esac
        else
            echo "containerd est déjà installé"
            # Définir DISTRO même si containerd est déjà installé
            if [ -f /etc/os-release ]; then
                . /etc/os-release
                DISTRO=$ID
            elif [ -f /etc/debian_version ]; then
                DISTRO="debian"
            elif [ -f /etc/redhat-release ]; then
                DISTRO="rhel"
            else
                DISTRO="unknown"
            fi
        fi
        ;;
    Darwin)
        if ! command -v containerd &> /dev/null; then
            echo "Installation de containerd via Homebrew..."
            brew install containerd
        else
            echo "containerd est déjà installé"
        fi
        DISTRO="darwin"
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
    echo "Recherche de la dernière version stable..."
    LATEST_VERSION=$(curl -s https://api.github.com/repos/litefaas/litefaas/releases/latest 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ] || [ "$LATEST_VERSION" = "null" ]; then
        echo "⚠️  Impossible de déterminer la dernière version depuis GitHub."
        echo "   Cela peut être dû à:"
        echo "   - Aucune release publiée"
        echo "   - Problème de connectivité réseau"
        echo "   - Repository privé ou inexistant"
        echo ""
        echo "   Utilisation de la version de développement..."
        LATEST_VERSION="dev"
    else
        echo "✅ Version trouvée: $LATEST_VERSION"
    fi
fi

# Déterminer l'architecture
case $ARCH in
    x86_64)
        BINARY_ARCH="amd64"
        ;;
    aarch64|arm64)
        echo "Architecture ARM64 non supportée. Seule l'architecture AMD64 est supportée."
        exit 1
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
    # Vérifier si nous sommes dans le bon répertoire
    if [ ! -f "go.mod" ]; then
        echo "❌ Fichier go.mod non trouvé. Assurez-vous d'être dans le répertoire du projet LiteFaaS."
        echo "   Cloner le repository d'abord:"
        echo "   git clone https://github.com/litefaas/litefaas.git"
        echo "   cd litefaas"
        exit 1
    fi

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
        # Vérifier si nous sommes dans le bon répertoire
        if [ ! -f "go.mod" ]; then
            echo "❌ Fichier go.mod non trouvé. Assurez-vous d'être dans le répertoire du projet LiteFaaS."
            echo "   Cloner le repository d'abord:"
            echo "   git clone https://github.com/litefaas/litefaas.git"
            echo "   cd litefaas"
            echo ""
            echo "   Ou télécharger directement le binaire depuis GitHub:"
            echo "   curl -L -o litefaas https://github.com/litefaas/litefaas/releases/latest/download/litefaas-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')"
            exit 1
        fi

        go mod download
        go build -o bin/litefaas cmd/litefaas/main.go
    fi
fi

# Création des répertoires nécessaires
echo "Création des répertoires nécessaires..."

case $DISTRO in
    alpine)
        # Alpine peut avoir des contraintes sur /var/lib
        sudo mkdir -p /var/lib/litefaas
        # Vérifier si l'utilisateur existe
        if id "$USER" >/dev/null 2>&1; then
            sudo chown $USER:$USER /var/lib/litefaas
        else
            # Sur Alpine, l'utilisateur peut être différent
            sudo chown root:root /var/lib/litefaas
            sudo chmod 755 /var/lib/litefaas
        fi
        ;;
    *)
        # Ubuntu et autres distributions
        sudo mkdir -p /var/lib/litefaas
        sudo chown $USER:$USER /var/lib/litefaas
        ;;
esac

# Configuration de containerd
echo "Configuration de containerd..."
sudo mkdir -p /etc/containerd
if [ ! -f /etc/containerd/config.toml ]; then
    containerd config default | sudo tee /etc/containerd/config.toml > /dev/null
fi

# Démarrer/redémarrer containerd selon la distribution
case $DISTRO in
    alpine)
        echo "Configuration pour Alpine Linux..."
        # Alpine utilise openrc ou systemd selon la version
        if command -v systemctl &> /dev/null; then
            sudo systemctl enable containerd
            sudo systemctl restart containerd
        elif command -v rc-update &> /dev/null; then
            sudo rc-update add containerd default
            sudo rc-service containerd start
        else
            echo "Démarrage manuel de containerd..."
            sudo containerd &
        fi
        ;;
    *)
        # Ubuntu et autres distributions avec systemd
        if command -v systemctl &> /dev/null; then
            sudo systemctl enable containerd
            sudo systemctl restart containerd
        else
            echo "Démarrage manuel de containerd..."
            sudo containerd &
        fi
        ;;
esac

# Vérification finale
echo "Vérification de l'installation..."

# Vérifier que containerd fonctionne
if command -v containerd &> /dev/null; then
    echo "✅ containerd installé"
else
    echo "❌ containerd non trouvé"
    exit 1
fi

# Vérifier que le binaire existe
if [ -f "bin/litefaas" ]; then
    echo "✅ Binaire LiteFaaS trouvé"
else
    echo "❌ Binaire LiteFaaS non trouvé"
    exit 1
fi

# Vérifier les permissions
if [ -d "/var/lib/litefaas" ]; then
    echo "✅ Répertoire de données créé"
else
    echo "❌ Répertoire de données non créé"
    exit 1
fi

echo ""
echo "🎉 Installation terminée avec succès!"
echo "Distribution détectée: $DISTRO"
echo "Pour démarrer LiteFaaS: ./bin/litefaas"
echo ""
echo "Note: Si vous utilisez Alpine Linux, assurez-vous que containerd est démarré:"
echo "  - Avec systemd: sudo systemctl start containerd"
echo "  - Avec openrc: sudo rc-service containerd start"
echo "  - Manuellement: sudo containerd &"
