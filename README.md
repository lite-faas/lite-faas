# LiteFaaS - Plateforme Function-as-a-Service Légère

LiteFaaS est une plateforme Function-as-a-Service légère et efficace en ressources, développée en Go. Elle permet de déployer et exécuter des fonctions serverless dans des conteneurs isolés.

## 🚀 Fonctionnalités

-   **Déploiement rapide** : Démarrage des fonctions en moins de 2 secondes
-   **Isolation des conteneurs** : Utilisation de containerd pour l'isolation
-   **Support multi-langages** : Python et Node.js
-   **Interface web moderne** : Dashboard intuitif pour la gestion des fonctions
-   **API RESTful** : Interface programmatique complète
-   **Base de données SQLite** : Stockage léger et efficace
-   **Déploiement en conteneur** : Architecture container-native

## 📋 Prérequis

-   Go 1.21+
-   containerd
-   Linux/macOS
-   Accès root pour l'installation

## 🛠️ Installation

### Installation automatique

```bash
# Cloner le repository
git clone https://github.com/litefaas/litefaas.git
cd litefaas

# Exécuter le script d'installation
chmod +x install.sh
sudo ./install.sh
```

### Installation manuelle

```bash
# Compiler l'application
go build -o litefaas cmd/litefaas/main.go

# Démarrer LiteFaaS
./litefaas

# Démarrer en mode développement (sans containerd)
./litefaas --dev --db-path ./dev.db --log-format text
```

### Déploiement en conteneur

```bash
# Construire l'image
docker build -t litefaas:latest .

# Exécuter avec accès à containerd
docker run --privileged \
  -v /run/containerd/containerd.sock:/run/containerd/containerd.sock \
  -v /var/lib/containerd:/var/lib/containerd \
  -v /var/lib/litefaas:/var/lib/litefaas \
  -p 8080:8080 \
  litefaas:latest
```

## 🎯 Utilisation

### Mode Développement

Pour tester LiteFaaS sans containerd (utile sur macOS ou pour le développement) :

```bash
# Démarrer en mode développement
./litefaas --dev --db-path ./dev.db --port 8080 --log-format text --log-level info

# L'API sera disponible mais les fonctions ne pourront pas être démarrées
curl http://localhost:8080/api/health
```

**Note** : En mode développement, vous pouvez créer et gérer des fonctions via l'API et l'interface web, mais vous ne pouvez pas les démarrer car containerd n'est pas disponible.

### Interface Web

Accédez à l'interface web sur `http://localhost:8080` pour :

-   Créer et gérer des fonctions
-   Démarrer/arrêter des fonctions
-   Visualiser les logs et métriques
-   Tester les fonctions

### API REST

#### Créer une fonction

```bash
curl -X POST http://localhost:8080/api/functions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello-world",
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
curl -X POST http://localhost:8080/api/functions/hello-world/start
```

#### Appeler une fonction

```bash
curl http://localhost:8080/hello-world
```

### Exemples de fonctions

#### Python

```python
import json
from http.server import HTTPServer, BaseHTTPRequestHandler

class FunctionHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()

        response = {
            'message': 'Hello from Python function!',
            'method': 'GET',
            'path': self.path
        }

        self.wfile.write(json.dumps(response).encode())

if __name__ == '__main__':
    server = HTTPServer(('127.0.0.1', 8080), FunctionHandler)
    server.serve_forever()
```

#### Node.js

```javascript
const http = require("http");

const server = http.createServer((req, res) => {
    const response = {
        message: "Hello from Node.js function!",
        method: req.method,
        path: req.url,
    };

    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify(response));
});

server.listen(8080, "127.0.0.1", () => {
    console.log("Node.js function server running on port 8080");
});
```

## ⚙️ Configuration

### Variables d'environnement

| Variable                          | Défaut                            | Description                      |
| --------------------------------- | --------------------------------- | -------------------------------- |
| `LITEFAAS_PORT`                   | 8080                              | Port d'écoute                    |
| `LITEFAAS_DB_PATH`                | `/var/lib/litefaas/functions.db`  | Chemin de la base de données     |
| `LITEFAAS_CONTAINERD_SOCKET`      | `/run/containerd/containerd.sock` | Socket containerd                |
| `LITEFAAS_FUNCTION_PORT_START`    | 9000                              | Port de début pour les fonctions |
| `LITEFAAS_FUNCTION_PORT_END`      | 9999                              | Port de fin pour les fonctions   |
| `LITEFAAS_CONTAINER_IMAGE_PREFIX` | `mirror.gcr.io/library/`          | Préfixe des images conteneur     |
| `LITEFAAS_LOG_LEVEL`              | info                              | Niveau de log                    |
| `LITEFAAS_LOG_FORMAT`             | json                              | Format des logs                  |

### Flags de ligne de commande

| Flag           | Défaut                           | Description                               |
| -------------- | -------------------------------- | ----------------------------------------- |
| `--dev`        | false                            | Mode développement (désactive containerd) |
| `--port`       | 8080                             | Port d'écoute                             |
| `--db-path`    | `/var/lib/litefaas/functions.db` | Chemin de la base de données              |
| `--log-level`  | info                             | Niveau de log                             |
| `--log-format` | json                             | Format des logs (json/text)               |

### Fichier de configuration

```yaml
# config.yaml
server:
    port: 8080
    host: "0.0.0.0"

database:
    path: "/var/lib/litefaas/functions.db"
    max_connections: 10

containerd:
    socket: "/run/containerd/containerd.sock"
    namespace: "litefaas"

container:
    image_prefix: "mirror.gcr.io/library/"
    privileged: true
    network_mode: "host"

functions:
    port_range:
        start: 9000
        end: 9999
    resource_limits:
        memory: "512MB"
        cpu: "0.5"

logging:
    level: "info"
    format: "json"
```

## 🏗️ Architecture

```
LiteFaaS/
├── cmd/litefaas/          # Point d'entrée principal
├── internal/
│   ├── api/              # Handlers API HTTP
│   ├── container/        # Gestion des conteneurs
│   ├── database/         # Opérations SQLite
│   ├── proxy/            # Logique de proxy
│   └── web/              # Interface web embarquée
├── pkg/                   # Packages publics
├── scripts/              # Scripts d'installation
├── templates/            # Templates de conteneurs
├── Dockerfile            # Définition du conteneur
├── install.sh            # Script d'installation
└── README.md             # Documentation
```

### Composants principaux

1. **Main Application** : Point d'entrée avec gestion de la configuration
2. **Container Manager** : Gestion des conteneurs via containerd
3. **Database Layer** : Stockage des métadonnées avec SQLite
4. **API Handlers** : Interface REST pour la gestion des fonctions
5. **Proxy Layer** : Routage des requêtes vers les conteneurs
6. **Web Interface** : Interface utilisateur web

## 🔧 Développement

### Prérequis de développement

```bash
# Installer les dépendances
go mod download

# Exécuter les tests
go test ./...

# Linter le code
golangci-lint run

# Construire l'application
go build -o bin/litefaas cmd/litefaas/main.go
```

### Structure du projet

-   **Go 1.21+** : Langage principal
-   **containerd** : Runtime de conteneurs
-   **SQLite** : Base de données
-   **Alpine Linux** : Images de base légères
-   **mirror.gcr.io** : Registry d'images

### Tests

```bash
# Tests unitaires
go test ./...

# Tests d'intégration
go test -tags=integration ./...

# Benchmarks
go test -bench=. ./...
```

## 🚀 Déploiement

### Docker Compose

```yaml
version: "3.8"
services:
    litefaas:
        image: litefaas:latest
        privileged: true
        volumes:
            - /run/containerd/containerd.sock:/run/containerd/containerd.sock
            - /var/lib/containerd:/var/lib/containerd
            - /var/lib/litefaas:/var/lib/litefaas
        ports:
            - "8080:8080"
        environment:
            - LITEFAAS_CONTAINERD_SOCKET=/run/containerd/containerd.sock
        restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: litefaas
spec:
    replicas: 1
    selector:
        matchLabels:
            app: litefaas
    template:
        metadata:
            labels:
                app: litefaas
        spec:
            containers:
                - name: litefaas
                  image: litefaas:latest
                  ports:
                      - containerPort: 8080
                  securityContext:
                      privileged: true
                  volumeMounts:
                      - name: containerd-sock
                        mountPath: /run/containerd/containerd.sock
                      - name: litefaas-data
                        mountPath: /var/lib/litefaas
            volumes:
                - name: containerd-sock
                  hostPath:
                      path: /run/containerd/containerd.sock
                - name: litefaas-data
                  persistentVolumeClaim:
                      claimName: litefaas-pvc
```

## 📊 Monitoring

### Métriques disponibles

-   Nombre de fonctions actives
-   Temps de démarrage des conteneurs
-   Utilisation des ressources
-   Latence des appels de fonctions

### Logs

Les logs sont structurés en JSON et incluent :

-   Création/suppression de fonctions
-   Démarrage/arrêt de conteneurs
-   Erreurs et exceptions
-   Métriques de performance

## 🔒 Sécurité

-   **Isolation des conteneurs** : Chaque fonction s'exécute dans son propre conteneur
-   **Limites de ressources** : CPU et mémoire limités par fonction
-   **Binding localhost** : Les fonctions ne sont accessibles que localement
-   **Validation des entrées** : Toutes les entrées utilisateur sont validées
-   **Pas d'authentification** : Conçu pour un usage local/privé

## 🤝 Contribution

1. Fork le projet
2. Créer une branche feature (`git checkout -b feature/AmazingFeature`)
3. Commit les changements (`git commit -m 'Add some AmazingFeature'`)
4. Push vers la branche (`git push origin feature/AmazingFeature`)
5. Ouvrir une Pull Request

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails.

## 🆘 Support

-   **Documentation** : [GitHub Wiki](https://github.com/litefaas/litefaas/wiki)
-   **Issues** : [GitHub Issues](https://github.com/litefaas/litefaas/issues)
-   **Discussions** : [GitHub Discussions](https://github.com/litefaas/litefaas/discussions)

## 🗺️ Roadmap

-   [ ] Support de Rust et Go pour les fonctions
-   [ ] Versioning des fonctions
-   [ ] Monitoring avancé et métriques
-   [ ] Scaling automatique basé sur la demande
-   [ ] Intégration avec des services externes
-   [ ] Système de plugins pour les runtimes personnalisés
-   [ ] Support multi-tenant
-   [ ] API GraphQL
-   [ ] Interface CLI
-   [ ] Templates de fonctions prédéfinis

---

**LiteFaaS** - Une plateforme FaaS légère et efficace pour vos besoins serverless.
