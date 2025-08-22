# Guide de développement LiteFaaS

## Problème résolu : Erreur 500 lors du démarrage de fonction

### Symptômes

-   `POST http://localhost:8080/api/functions/test/start` retourne une erreur 500
-   Message d'erreur : "container management is disabled in development mode"

### Cause

L'application était configurée pour exiger containerd même en mode développement, ce qui causait une erreur quand containerd n'était pas disponible.

### Solution implémentée

1. **Mode développement amélioré** : L'application peut maintenant fonctionner sans containerd en mode développement
2. **Simulation de conteneurs** : En mode dev, les fonctions sont marquées comme "running" sans créer de vrais conteneurs
3. **Gestion d'erreurs améliorée** : Messages d'erreur plus informatifs avec détails JSON

### Utilisation

#### Mode développement (recommandé pour les tests)

```bash
# Démarrage rapide
./dev.sh

# Ou avec make
make dev

# Ou manuellement
go build -o litefaas cmd/litefaas/main.go
./litefaas --dev --db-path ./dev.db --port 8080 --log-format text --log-level debug
```

#### Mode production (avec containerd)

```bash
# Construire et exécuter
make build
./litefaas --containerd-socket /run/containerd/containerd.sock
```

### Test de l'API

```bash
# Test complet
make test-api

# Ou manuellement
./test_diagnostic.sh
```

### Endpoints API

-   `GET /api/health` - Santé du service
-   `GET /api/functions` - Liste des fonctions
-   `POST /api/functions` - Créer une fonction
-   `POST /api/functions/{name}/start` - Démarrer une fonction
-   `POST /api/functions/{name}/stop` - Arrêter une fonction
-   `GET /api/functions/{name}` - Détails d'une fonction
-   `PUT /api/functions/{name}` - Modifier une fonction
-   `DELETE /api/functions/{name}` - Supprimer une fonction

### Structure du projet

```
LiteFaaS/
├── cmd/litefaas/main.go      # Point d'entrée principal
├── internal/
│   ├── api/handler.go        # Handlers API REST
│   ├── container/manager.go  # Gestion des conteneurs
│   ├── database/database.go  # Couche base de données
│   ├── proxy/proxy.go        # Proxy pour les fonctions
│   └── web/handler.go        # Interface web
├── dev.sh                    # Script de développement
├── test_diagnostic.sh        # Tests de diagnostic
└── Makefile                  # Commandes de build
```

### Débogage

1. **Vérifier les logs** : L'application affiche des logs détaillés en mode debug
2. **Tester l'API** : Utiliser `test_diagnostic.sh` pour tester tous les endpoints
3. **Vérifier la base de données** : Le fichier `dev.db` contient les métadonnées des fonctions

### Prochaines étapes

-   [ ] Implémenter un vrai runtime de fonctions en mode développement
-   [ ] Ajouter des tests unitaires complets
-   [ ] Améliorer l'interface web
-   [ ] Ajouter la gestion des versions de fonctions
