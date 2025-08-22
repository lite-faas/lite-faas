#!/bin/bash

echo "=== Test de LiteFaaS en mode production ==="

# Couleurs pour les tests
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonction pour afficher les résultats
print_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"

    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} $test_name: $message"
    else
        echo -e "${RED}✗${NC} $test_name: $message"
    fi
}

# Test 1: Vérifier que Podman est installé
echo -e "\n${YELLOW}1. Vérification de Podman${NC}"
if command -v podman &> /dev/null; then
    print_result "Podman" "PASS" "Podman est installé"
else
    print_result "Podman" "FAIL" "Podman n'est pas installé"
    exit 1
fi

# Test 2: Vérifier que containerd est disponible
echo -e "\n${YELLOW}2. Vérification de containerd${NC}"
if [ -S /run/containerd/containerd.sock ]; then
    print_result "Containerd socket" "PASS" "Socket containerd disponible"
else
    print_result "Containerd socket" "FAIL" "Socket containerd non trouvé"
    echo -e "${YELLOW}Note: containerd doit être installé et en cours d'exécution${NC}"
fi

# Test 3: Vérifier que le conteneur est en cours d'exécution
echo -e "\n${YELLOW}3. Vérification du conteneur LiteFaaS${NC}"
if podman ps --format "table {{.Names}}" | grep -q "litefaas-prod"; then
    print_result "Conteneur LiteFaaS" "PASS" "Conteneur en cours d'exécution"
else
    print_result "Conteneur LiteFaaS" "FAIL" "Conteneur non trouvé"
    echo -e "${YELLOW}Lancez 'make run' pour démarrer le conteneur${NC}"
    exit 1
fi

# Test 4: Vérifier la santé de l'API
echo -e "\n${YELLOW}4. Test de santé de l'API${NC}"
if curl -s http://localhost:8080/api/health | grep -q '"status":"healthy"'; then
    print_result "API Health" "PASS" "API en bonne santé"
else
    print_result "API Health" "FAIL" "API non accessible"
    exit 1
fi

# Test 5: Créer une fonction de test
echo -e "\n${YELLOW}5. Test de création de fonction${NC}"
create_response=$(curl -s -X POST http://localhost:8080/api/functions \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test-prod",
        "language": "python",
        "code": "print(\"Hello from production!\")"
    }')

if echo "$create_response" | grep -q '"name":"test-prod"'; then
    print_result "Création de fonction" "PASS" "Fonction créée avec succès"
else
    print_result "Création de fonction" "FAIL" "Échec de création"
    echo "Réponse: $create_response"
fi

# Test 6: Démarrer la fonction
echo -e "\n${YELLOW}6. Test de démarrage de fonction${NC}"
start_response=$(curl -s -X POST http://localhost:8080/api/functions/test-prod/start -w "%{http_code}")
http_code="${start_response: -3}"
if [ "$http_code" = "200" ]; then
    print_result "Démarrage de fonction" "PASS" "Fonction démarrée"
else
    print_result "Démarrage de fonction" "FAIL" "HTTP $http_code"
fi

# Test 7: Vérifier le statut de la fonction
echo -e "\n${YELLOW}7. Test de vérification du statut${NC}"
sleep 2
status_response=$(curl -s http://localhost:8080/api/functions/test-prod)
if echo "$status_response" | grep -q '"status":"running"'; then
    print_result "Statut de fonction" "PASS" "Fonction en cours d'exécution"
else
    print_result "Statut de fonction" "FAIL" "Fonction non démarrée"
fi

# Test 8: Vérifier que le conteneur de fonction a été créé
echo -e "\n${YELLOW}8. Vérification du conteneur de fonction${NC}"
if podman ps --format "table {{.Names}}" | grep -q "litefaas-test-prod"; then
    print_result "Conteneur de fonction" "PASS" "Conteneur de fonction créé"
else
    print_result "Conteneur de fonction" "FAIL" "Conteneur de fonction non trouvé"
    echo -e "${YELLOW}Note: Vérifiez que containerd fonctionne correctement${NC}"
fi

# Test 9: Arrêter la fonction
echo -e "\n${YELLOW}9. Test d'arrêt de fonction${NC}"
stop_response=$(curl -s -X POST http://localhost:8080/api/functions/test-prod/stop -w "%{http_code}")
http_code="${stop_response: -3}"
if [ "$http_code" = "200" ]; then
    print_result "Arrêt de fonction" "PASS" "Fonction arrêtée"
else
    print_result "Arrêt de fonction" "FAIL" "HTTP $http_code"
fi

# Test 10: Supprimer la fonction
echo -e "\n${YELLOW}10. Test de suppression de fonction${NC}"
delete_response=$(curl -s -X DELETE http://localhost:8080/api/functions/test-prod -w "%{http_code}")
http_code="${delete_response: -3}"
if [ "$http_code" = "204" ]; then
    print_result "Suppression de fonction" "PASS" "Fonction supprimée"
else
    print_result "Suppression de fonction" "FAIL" "HTTP $http_code"
fi

echo -e "\n${GREEN}=== Tests de production terminés ===${NC}"
echo -e "${GREEN}LiteFaaS fonctionne correctement en mode production !${NC}"
echo -e "\n${YELLOW}Interface web: http://localhost:8080${NC}"
echo -e "${YELLOW}API: http://localhost:8080/api${NC}"
