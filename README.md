# LiteFaaS

Une plateforme légère de fonctions serverless écrite en Go, conçue pour une utilisation locale avec une empreinte mémoire minimale.

## Fonctionnalités

-   **Léger** : Empreinte mémoire < 50MB
-   **Rapide** : Démarrage des fonctions < 2 secondes
-   **Simple** : Interface web intuitive
-   **Efficace** : Utilise containerd pour la gestion des conteneurs
-   **Local** : Fonctionne entièrement en local
-   **Multi-langage** : Support Python et Node.js

## Architecture

```
LiteFaaS/
├── cmd/litefaas/          # Point d'entrée principal
├── internal/
│   ├── api/              # Gestionnaires API HTTP
│   ├── container/        # Gestion des conteneurs
│   ├── database/         # Opérations SQLite
│   ├── proxy/            # Routage des requêtes
│   └── web/              # Interface web
├── templates/            # Templates de conteneurs
└── scripts/              # Scripts utilitaires
```

## Installation

### Prérequis

-   containerd
-   SQLite

### Distributions supportées

-   Ubuntu/Debian
-   Alpine Linux
-   RHEL/CentOS/Fedora
-   macOS (via Homebrew)

### Installation rapide

#### Option 1: Installation depuis le code source (recommandé pour le développement)

```bash
# Cloner le repository
git clone https://github.com/your-org/litefaas.git
cd litefaas

# Installer (télécharge le binaire depuis GitHub)
chmod +x install.sh
./install.sh

# Démarrer
./bin/litefaas
```

**Note:** Ce script télécharge le binaire depuis GitHub. Pour compiler localement, utilisez :

```bash
./scripts/build.sh
```

#### Option 2: Installation standalone (recommandé pour la production)

```bash
# Télécharger et exécuter le script d'installation
curl -fsSL https://raw.githubusercontent.com/litefaas/litefaas/main/install-standalone.sh | bash

# Ou télécharger le script et l'exécuter
wget https://raw.githubusercontent.com/litefaas/litefaas/main/install-standalone.sh
chmod +x install-standalone.sh
./install-standalone.sh

# Démarrer
./bin/litefaas
```

### Installation manuelle

```bash
# Installer containerd
# Sur Ubuntu/Debian:
sudo apt-get update && sudo apt-get install -y containerd.io

# Sur RHEL/CentOS/Fedora:
sudo dnf install -y containerd  # ou sudo yum install -y containerd

# Sur macOS:
brew install containerd

# Télécharger le binaire depuis GitHub
# Remplacer VERSION par la version souhaitée (ex: v1.0.0)
curl -L -o litefaas https://github.com/litefaas/litefaas/releases/download/VERSION/litefaas-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
chmod +x litefaas

# Créer le répertoire de données
sudo mkdir -p /var/lib/litefaas
sudo chown $USER:$USER /var/lib/litefaas

# Démarrer
./litefaas
```

## Utilisation

### Interface Web

Accédez à l'interface web à l'adresse : http://localhost:8080/web/

### API REST

#### Créer une fonction

```bash
curl -X POST http://localhost:8080/api/functions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello",
    "language": "python",
    "code": "print(\"Hello, World!\")"
  }'
```

#### Lister les fonctions

```bash
curl http://localhost:8080/api/functions
```

#### Démarrer une fonction

```bash
curl -X POST http://localhost:8080/api/functions/hello/start
```

#### Appeler une fonction

```bash
curl http://localhost:8080/hello
```

## Configuration

### Variables d'environnement

-   `LITEFAAS_PORT` : Port du serveur (défaut: 8080)
-   `LITEFAAS_DB_PATH` : Chemin de la base de données (défaut: /var/lib/litefaas/functions.db)
-   `LITEFAAS_CONTAINERD_SOCKET` : Socket containerd (défaut: /run/containerd/containerd.sock)
-   `LITEFAAS_FUNCTION_PORT_START` : Port de début pour les fonctions (défaut: 9000)
-   `LITEFAAS_FUNCTION_PORT_END` : Port de fin pour les fonctions (défaut: 9999)

### Arguments de ligne de commande

```bash
./litefaas \
  --port 8080 \
  --db-path /var/lib/litefaas/functions.db \
  --containerd-socket /run/containerd/containerd.sock \
  --function-port-start 9000 \
  --function-port-end 9999 \
  --log-level info
```

## Développement

### Structure du projet

-   `cmd/litefaas/` : Point d'entrée de l'application
-   `internal/api/` : Gestionnaires HTTP pour l'API REST
-   `internal/container/` : Gestion des conteneurs avec containerd
-   `internal/database/` : Opérations de base de données SQLite
-   `internal/proxy/` : Routage des requêtes vers les fonctions
-   `internal/web/` : Interface web intégrée
-   `templates/` : Templates de conteneurs pour Python et Node.js

### Tests

```bash
go test ./...
```

### Compilation

```bash
# Pour le développement local
go build -o bin/litefaas cmd/litefaas/main.go

# Pour la production (optimisé)
go build -ldflags="-s -w" -o bin/litefaas cmd/litefaas/main.go

# Mode développement (avec script automatisé)
./scripts/dev.sh
```

## Sécurité

-   Isolation des conteneurs avec containerd
-   Validation des entrées utilisateur
-   Limites de ressources par fonction
-   Liaison localhost uniquement

## Performance

-   Démarrage rapide des fonctions (< 2s)
-   Gestion efficace des connexions
-   Pool de connexions pour la base de données
-   Optimisation des requêtes SQL

## Releases

Les binaires précompilés sont automatiquement générés pour chaque release via GitHub Actions.

### Plateformes supportées

-   Linux (amd64)
-   macOS (amd64)

### Téléchargement

```bash
# Dernière version stable
curl -L -o litefaas https://github.com/litefaas/litefaas/releases/latest/download/litefaas-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

# Version spécifique
curl -L -o litefaas https://github.com/litefaas/litefaas/releases/download/v1.0.0/litefaas-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

# Version de développement (dernière build de la branche dev)
# Note: Les builds de développement sont disponibles dans les artifacts GitHub Actions
```

## Support

Pour les questions et problèmes, veuillez ouvrir une issue sur GitHub.

## Licence

MIT License
