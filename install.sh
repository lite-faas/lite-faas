#!/bin/bash

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LITEFAAS_VERSION="latest"
LITEFAAS_BINARY="litefaas"
LITEFAAS_SERVICE="litefaas"
LITEFAAS_USER="litefaas"
LITEFAAS_GROUP="litefaas"
LITEFAAS_DIR="/opt/litefaas"
LITEFAAS_DATA_DIR="/var/lib/litefaas"
LITEFAAS_CONFIG_DIR="/etc/litefaas"

# Fonctions utilitaires
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Détection du système d'exploitation
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [[ -f /etc/os-release ]]; then
            . /etc/os-release
            OS=$NAME
            VER=$VERSION_ID
        else
            OS=$(uname -s)
            VER=$(uname -r)
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macOS"
        VER=$(sw_vers -productVersion)
    else
        OS="Unknown"
        VER="Unknown"
    fi
}

# Vérification des prérequis
check_prerequisites() {
    log_info "Vérification des prérequis..."

    # Vérifier si Go est installé
    if ! command -v go &> /dev/null; then
        log_warning "Go n'est pas installé. Installation..."
        install_go
    else
        log_success "Go est déjà installé"
    fi

    # Vérifier si containerd est installé
    if ! command -v containerd &> /dev/null; then
        log_warning "Containerd n'est pas installé. Installation..."
        install_containerd
    else
        log_success "Containerd est déjà installé"
    fi
}

# Installation de Go
install_go() {
    log_info "Installation de Go..."

    if [[ "$OS" == "Ubuntu" ]] || [[ "$OS" == "Debian GNU/Linux" ]]; then
        sudo apt-get update
        sudo apt-get install -y golang-go
    elif [[ "$OS" == "CentOS Linux" ]] || [[ "$OS" == "Red Hat Enterprise Linux" ]]; then
        sudo yum install -y golang
    elif [[ "$OS" == "macOS" ]]; then
        if command -v brew &> /dev/null; then
            brew install go
        else
            log_error "Homebrew n'est pas installé. Veuillez installer Homebrew ou Go manuellement."
            exit 1
        fi
    else
        log_error "Système d'exploitation non supporté: $OS"
        exit 1
    fi
}

# Installation de containerd
install_containerd() {
    log_info "Installation de containerd..."

    if [[ "$OS" == "Ubuntu" ]] || [[ "$OS" == "Debian GNU/Linux" ]]; then
        sudo apt-get update
        sudo apt-get install -y containerd
    elif [[ "$OS" == "CentOS Linux" ]] || [[ "$OS" == "Red Hat Enterprise Linux" ]]; then
        sudo yum install -y containerd
    elif [[ "$OS" == "macOS" ]]; then
        if command -v brew &> /dev/null; then
            brew install containerd
        else
            log_error "Homebrew n'est pas installé. Veuillez installer Homebrew ou containerd manuellement."
            exit 1
        fi
    else
        log_error "Système d'exploitation non supporté: $OS"
        exit 1
    fi

    # Démarrer containerd
    sudo systemctl enable containerd
    sudo systemctl start containerd
}

# Création de l'utilisateur et des répertoires
setup_directories() {
    log_info "Configuration des répertoires et utilisateur..."

    # Créer l'utilisateur si il n'existe pas
    if ! id "$LITEFAAS_USER" &>/dev/null; then
        sudo useradd -r -s /bin/false -d "$LITEFAAS_DIR" "$LITEFAAS_USER"
        log_success "Utilisateur $LITEFAAS_USER créé"
    fi

    # Créer les répertoires
    sudo mkdir -p "$LITEFAAS_DIR"
    sudo mkdir -p "$LITEFAAS_DATA_DIR"
    sudo mkdir -p "$LITEFAAS_CONFIG_DIR"

    # Définir les permissions
    sudo chown -R "$LITEFAAS_USER:$LITEFAAS_GROUP" "$LITEFAAS_DIR"
    sudo chown -R "$LITEFAAS_USER:$LITEFAAS_GROUP" "$LITEFAAS_DATA_DIR"
    sudo chown -R "$LITEFAAS_USER:$LITEFAAS_GROUP" "$LITEFAAS_CONFIG_DIR"

    log_success "Répertoires configurés"
}

# Téléchargement et installation du binaire
install_binary() {
    log_info "Téléchargement de LiteFaaS..."

    # Pour l'instant, on compile depuis les sources
    # TODO: Implémenter le téléchargement depuis GitHub releases

    if [[ -f "go.mod" ]]; then
        log_info "Compilation depuis les sources..."
        go build -o "$LITEFAAS_BINARY" cmd/litefaas/main.go
        sudo cp "$LITEFAAS_BINARY" "$LITEFAAS_DIR/"
        sudo chown "$LITEFAAS_USER:$LITEFAAS_GROUP" "$LITEFAAS_DIR/$LITEFAAS_BINARY"
        sudo chmod +x "$LITEFAAS_DIR/$LITEFAAS_BINARY"
    else
        log_error "Sources non trouvées. Veuillez exécuter ce script depuis le répertoire du projet."
        exit 1
    fi

    log_success "Binaire installé dans $LITEFAAS_DIR/$LITEFAAS_BINARY"
}

# Configuration du service systemd
setup_systemd_service() {
    if [[ "$OS" == "macOS" ]]; then
        log_warning "Systemd non disponible sur macOS. Service non configuré."
        return
    fi

    log_info "Configuration du service systemd..."

    cat << EOF | sudo tee /etc/systemd/system/$LITEFAAS_SERVICE.service
[Unit]
Description=LiteFaaS Function Server
After=network.target containerd.service
Requires=containerd.service

[Service]
Type=simple
User=$LITEFAAS_USER
Group=$LITEFAAS_GROUP
ExecStart=$LITEFAAS_DIR/$LITEFAAS_BINARY
Restart=always
RestartSec=5
Environment=LITEFAAS_DB_PATH=$LITEFAAS_DATA_DIR/functions.db
Environment=LITEFAAS_CONTAINERD_SOCKET=/run/containerd/containerd.sock

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable $LITEFAAS_SERVICE

    log_success "Service systemd configuré"
}

# Configuration du firewall
setup_firewall() {
    if [[ "$OS" == "macOS" ]]; then
        log_warning "Configuration du firewall macOS non implémentée"
        return
    fi

    log_info "Configuration du firewall..."

    if command -v ufw &> /dev/null; then
        sudo ufw allow 8080/tcp
        log_success "Port 8080 ouvert avec ufw"
    elif command -v firewall-cmd &> /dev/null; then
        sudo firewall-cmd --permanent --add-port=8080/tcp
        sudo firewall-cmd --reload
        log_success "Port 8080 ouvert avec firewalld"
    else
        log_warning "Aucun gestionnaire de firewall détecté"
    fi
}

# Démarrage du service
start_service() {
    if [[ "$OS" == "macOS" ]]; then
        log_info "Démarrage manuel de LiteFaaS..."
        log_info "Pour démarrer LiteFaaS, exécutez: $LITEFAAS_DIR/$LITEFAAS_BINARY"
        return
    fi

    log_info "Démarrage du service..."
    sudo systemctl start $LITEFAAS_SERVICE

    if sudo systemctl is-active --quiet $LITEFAAS_SERVICE; then
        log_success "Service démarré avec succès"
    else
        log_error "Échec du démarrage du service"
        sudo systemctl status $LITEFAAS_SERVICE
        exit 1
    fi
}

# Vérification de l'installation
verify_installation() {
    log_info "Vérification de l'installation..."

    # Vérifier que le binaire existe
    if [[ ! -f "$LITEFAAS_DIR/$LITEFAAS_BINARY" ]]; then
        log_error "Binaire non trouvé"
        exit 1
    fi

    # Vérifier que le service répond
    sleep 3
    if curl -s http://localhost:8080/api/health > /dev/null; then
        log_success "LiteFaaS répond correctement sur http://localhost:8080"
    else
        log_warning "LiteFaaS ne répond pas encore. Vérifiez les logs avec: sudo journalctl -u $LITEFAAS_SERVICE"
    fi
}

# Affichage des informations finales
show_final_info() {
    log_success "Installation de LiteFaaS terminée !"
    echo
    echo "Informations importantes:"
    echo "- Interface web: http://localhost:8080"
    echo "- API: http://localhost:8080/api"
    echo "- Base de données: $LITEFAAS_DATA_DIR/functions.db"
    echo "- Configuration: $LITEFAAS_CONFIG_DIR"
    echo
    if [[ "$OS" != "macOS" ]]; then
        echo "Commandes utiles:"
        echo "- Démarrer: sudo systemctl start $LITEFAAS_SERVICE"
        echo "- Arrêter: sudo systemctl stop $LITEFAAS_SERVICE"
        echo "- Status: sudo systemctl status $LITEFAAS_SERVICE"
        echo "- Logs: sudo journalctl -u $LITEFAAS_SERVICE -f"
    fi
    echo
    echo "Documentation: https://github.com/litefaas/litefaas"
}

# Fonction principale
main() {
    log_info "Début de l'installation de LiteFaaS..."

    detect_os
    log_info "Système détecté: $OS $VER"

    check_prerequisites
    setup_directories
    install_binary
    setup_systemd_service
    setup_firewall
    start_service
    verify_installation
    show_final_info
}

# Exécution du script
main "$@"
