#!/bin/bash

echo "=== Tests complets LiteFaaS ==="

# Couleurs pour les tests
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variables
API_TESTS_PASSED=0
WEB_TESTS_PASSED=0
TOTAL_TESTS=0

# Fonction pour exécuter un test et compter les résultats
run_test() {
    local test_name="$1"
    local test_script="$2"

    echo -e "\n${BLUE}=== $test_name ===${NC}"
    if ./"$test_script" > /tmp/test_output.log 2>&1; then
        echo -e "${GREEN}✓ $test_name: SUCCESS${NC}"
        if [ "$test_name" = "Tests API" ]; then
            API_TESTS_PASSED=1
        elif [ "$test_name" = "Tests Interface Web" ]; then
            WEB_TESTS_PASSED=1
        fi
    else
        echo -e "${RED}✗ $test_name: FAILED${NC}"
        echo -e "${YELLOW}Détails:${NC}"
        cat /tmp/test_output.log
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Vérifier que le serveur est en cours d'exécution
echo -e "${YELLOW}Vérification du serveur...${NC}"
if ! curl -s http://localhost:8080/api/health > /dev/null; then
    echo -e "${RED}Erreur: Le serveur LiteFaaS n'est pas en cours d'exécution${NC}"
    echo -e "${YELLOW}Démarrez le serveur avec: ./dev.sh${NC}"
    exit 1
fi

echo -e "${GREEN}Serveur détecté, lancement des tests...${NC}"

# Exécuter les tests API
run_test "Tests API" "test_complete.sh"

# Exécuter les tests web
run_test "Tests Interface Web" "test_web.sh"

# Résumé
echo -e "\n${BLUE}=== RÉSUMÉ DES TESTS ===${NC}"
echo -e "Tests API: $([ $API_TESTS_PASSED -eq 1 ] && echo -e "${GREEN}✓ PASS${NC}" || echo -e "${RED}✗ FAIL${NC}")"
echo -e "Tests Web: $([ $WEB_TESTS_PASSED -eq 1 ] && echo -e "${GREEN}✓ PASS${NC}" || echo -e "${RED}✗ FAIL${NC}")"

if [ $API_TESTS_PASSED -eq 1 ] && [ $WEB_TESTS_PASSED -eq 1 ]; then
    echo -e "\n${GREEN}🎉 Tous les tests sont passés avec succès !${NC}"
    echo -e "${GREEN}LiteFaaS est prêt à être utilisé.${NC}"
    echo -e "\n${YELLOW}Interface web: http://localhost:8080${NC}"
    echo -e "${YELLOW}API: http://localhost:8080/api${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Certains tests ont échoué.${NC}"
    echo -e "${YELLOW}Consultez les logs ci-dessus pour plus de détails.${NC}"
    exit 1
fi
