# Guide de Développement LiteFaaS

Ce document décrit comment contribuer au développement de LiteFaaS.

## 🛠️ Environnement de Développement

### Prérequis

-   Go 1.21+
-   containerd
-   Git
-   Make (optionnel)

### Installation de l'environnement

```bash
# Cloner le repository
git clone https://github.com/litefaas/litefaas.git
cd litefaas

# Installer les dépendances
go mod download
go mod tidy

# Vérifier l'installation
go build -o litefaas cmd/litefaas/main.go
```

## 🏗️ Architecture du Code

### Structure des Packages

```
internal/
├── api/          # Handlers HTTP pour l'API REST
├── container/    # Gestion des conteneurs via containerd
├── database/     # Couche d'accès aux données SQLite
├── proxy/        # Proxy pour router les requêtes
└── web/          # Interface web embarquée
```

### Conventions de Code

-   **Nommage** : Utiliser des noms explicites en anglais
-   **Commentaires** : Commenter les fonctions publiques
-   **Erreurs** : Toujours retourner des erreurs avec contexte
-   **Tests** : Écrire des tests pour les nouvelles fonctionnalités

### Formatage et Linting

```bash
# Formater le code
go fmt ./...

# Linter le code
golangci-lint run

# Vérifier les imports
goimports -w .
```

## 🧪 Tests

### Exécuter les Tests

```bash
# Tous les tests
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests d'un package spécifique
go test ./internal/api/...

# Tests avec verbosité
go test -v ./...
```

### Écrire des Tests

-   Utiliser `testing` package standard
-   Créer des tests unitaires pour chaque fonction
-   Utiliser des mocks pour les dépendances externes
-   Tester les cas d'erreur

Exemple de test :

```go
func TestCreateFunction(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    handler := api.NewHandler(db, nil, nil)

    // Act
    req := createTestRequest(t, "POST", "/api/functions", functionData)
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    // Assert
    assert.Equal(t, http.StatusCreated, rr.Code)
}
```

## 🚀 Développement Local

### Démarrer en Mode Développement

```bash
# Démarrer avec des paramètres de développement
./litefaas \
  --port 8080 \
  --db-path ./dev.db \
  --log-level debug \
  --log-format text
```

### Variables d'Environnement de Développement

```bash
export LITEFAAS_PORT=8080
export LITEFAAS_DB_PATH=./dev.db
export LITEFAAS_LOG_LEVEL=debug
export LITEFAAS_LOG_FORMAT=text
```

### Debugging

Pour déboguer avec Delve :

```bash
# Installer Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Démarrer en mode debug
dlv debug cmd/litefaas/main.go
```

## 📦 Build et Déploiement

### Build Local

```bash
# Build de développement
go build -o litefaas cmd/litefaas/main.go

# Build de production
CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o litefaas cmd/litefaas/main.go
```

### Build Docker

```bash
# Build de l'image
docker build -t litefaas:dev .

# Run en mode développement
docker run --privileged \
  -v /run/containerd/containerd.sock:/run/containerd/containerd.sock \
  -p 8080:8080 \
  litefaas:dev
```

## 🔧 Outils de Développement

### Makefile

Le projet inclut un Makefile avec des commandes utiles :

```bash
# Voir toutes les commandes
make help

# Build
make build

# Tests
make test

# Lint
make lint

# Format
make fmt

# Docker
make docker-build
make docker-run
```

### Scripts Utiles

```bash
# Nettoyer les artefacts de build
make clean

# Installer les dépendances
make deps

# Installer l'application
make install
```

## 📝 Contribution

### Workflow Git

1. **Fork** le repository
2. **Clone** votre fork
3. **Créer** une branche feature
4. **Développer** votre fonctionnalité
5. **Tester** votre code
6. **Commit** avec des messages clairs
7. **Push** vers votre fork
8. **Créer** une Pull Request

### Messages de Commit

Utiliser le format conventionnel :

```
type(scope): description

[optional body]

[optional footer]
```

Types :

-   `feat` : Nouvelle fonctionnalité
-   `fix` : Correction de bug
-   `docs` : Documentation
-   `style` : Formatage
-   `refactor` : Refactoring
-   `test` : Tests
-   `chore` : Maintenance

Exemples :

```
feat(api): add function creation endpoint
fix(container): resolve container startup issue
docs(readme): update installation instructions
```

### Pull Request

Avant de soumettre une PR :

-   [ ] Code formaté (`make fmt`)
-   [ ] Tests passent (`make test`)
-   [ ] Lint passe (`make lint`)
-   [ ] Documentation mise à jour
-   [ ] Description claire de la PR

## 🐛 Debugging

### Logs

Les logs sont structurés en JSON par défaut :

```bash
# Logs en mode texte pour le développement
./litefaas --log-format text --log-level debug
```

### Métriques

L'application expose des métriques sur `/api/health` :

```bash
curl http://localhost:8080/api/health
```

### Profiling

Pour activer le profiling CPU :

```bash
# Démarrer avec profiling
./litefaas --cpu-profile cpu.prof

# Analyser le profil
go tool pprof cpu.prof
```

## 🔒 Sécurité

### Bonnes Pratiques

-   Valider toutes les entrées utilisateur
-   Utiliser des limites de ressources
-   Isoler les conteneurs de fonctions
-   Éviter l'exécution de code arbitraire
-   Logs sécurisés (pas de données sensibles)

### Tests de Sécurité

```bash
# Tests de sécurité
go test -tags=security ./...

# Audit des dépendances
go list -json -deps ./... | jq 'select(.Vulnerabilities)'
```

## 📚 Ressources

-   [Documentation Go](https://golang.org/doc/)
-   [Containerd Documentation](https://github.com/containerd/containerd)
-   [SQLite Documentation](https://www.sqlite.org/docs.html)
-   [Go Testing](https://golang.org/pkg/testing/)

## 🤝 Support

Pour des questions de développement :

-   **Issues** : [GitHub Issues](https://github.com/litefaas/litefaas/issues)
-   **Discussions** : [GitHub Discussions](https://github.com/litefaas/litefaas/discussions)
-   **Documentation** : [GitHub Wiki](https://github.com/litefaas/litefaas/wiki)
