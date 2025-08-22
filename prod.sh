#!/bin/bash

echo "=== Démarrage LiteFaaS en mode production ==="

# Vérifier que Buildah est installé
if ! command -v buildah &> /dev/null; then
    echo "Erreur: Buildah n'est pas installé"
    echo "Installez Buildah:"
    echo "  - macOS: brew install buildah"
    echo "  - Ubuntu/Debian: sudo apt-get install buildah"
    echo "  - CentOS/RHEL: sudo yum install buildah"
    exit 1
fi

# Vérifier que Podman est installé
if ! command -v podman &> /dev/null; then
    echo "Erreur: Podman n'est pas installé"
    echo "Installez Podman:"
    echo "  - macOS: brew install podman"
    echo "  - Ubuntu/Debian: sudo apt-get install podman"
    echo "  - CentOS/RHEL: sudo yum install podman"
    exit 1
fi

# Vérifier que containerd est disponible
if [ ! -S /run/containerd/containerd.sock ]; then
    echo "Attention: Socket containerd non trouvé"
    echo "Assurez-vous que containerd est installé et en cours d'exécution"
    echo "Sur macOS, vous pouvez utiliser Docker Desktop"
    echo "Sur Linux, installez containerd: sudo apt-get install containerd"
fi

# Arrêter le conteneur existant s'il y en a un
echo "Arrêt du conteneur existant..."
docker stop litefaas-prod 2>/dev/null || echo "Aucun conteneur à arrêter"

# Construire et démarrer le conteneur
echo "Construction et démarrage du conteneur..."
make run

# Attendre que le conteneur démarre
echo "Attente du démarrage du conteneur..."
sleep 5

# Vérifier que le conteneur fonctionne
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo "✅ LiteFaaS est démarré en mode production"
    echo "🌐 Interface web: http://localhost:8080"
    echo "🔧 API: http://localhost:8080/api"
    echo ""
    echo "Commandes utiles:"
    echo "  make logs     - Voir les logs"
    echo "  make stop     - Arrêter le conteneur"
    echo "  make test-production - Tester le mode production"
else
    echo "❌ Erreur: Le conteneur ne répond pas"
    echo "Vérifiez les logs avec: make logs"
    exit 1
fi
