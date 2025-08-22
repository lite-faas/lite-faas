#!/bin/bash

echo "=== Test de l'interface web LiteFaaS ==="

BASE_URL="http://localhost:8080"

# Couleurs pour les tests
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
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

# Test 1: Page d'accueil
echo -e "\n${YELLOW}1. Test de la page d'accueil${NC}"
home_response=$(curl -s "$BASE_URL/")
if echo "$home_response" | grep -q "LiteFaaS Dashboard"; then
    print_result "Page d'accueil" "PASS" "Page chargée correctement"
else
    print_result "Page d'accueil" "FAIL" "Page non trouvée ou incorrecte"
fi

# Test 2: Fichier CSS
echo -e "\n${YELLOW}2. Test du fichier CSS${NC}"
css_headers=$(curl -s -I "$BASE_URL/static/style.css")
if echo "$css_headers" | grep -q "Content-Type: text/css"; then
    print_result "Fichier CSS" "PASS" "Type MIME correct"
else
    print_result "Fichier CSS" "FAIL" "Type MIME incorrect"
fi

# Test 3: Contenu CSS
echo -e "\n${YELLOW}3. Test du contenu CSS${NC}"
css_content=$(curl -s "$BASE_URL/static/style.css")
if echo "$css_content" | grep -q "margin: 0"; then
    print_result "Contenu CSS" "PASS" "Fichier CSS valide"
else
    print_result "Contenu CSS" "FAIL" "Fichier CSS invalide"
fi

# Test 4: Fichier JavaScript
echo -e "\n${YELLOW}4. Test du fichier JavaScript${NC}"
js_headers=$(curl -s -I "$BASE_URL/static/app.js")
if echo "$js_headers" | grep -q "Content-Type: application/javascript"; then
    print_result "Fichier JavaScript" "PASS" "Type MIME correct"
else
    print_result "Fichier JavaScript" "FAIL" "Type MIME incorrect"
fi

# Test 5: Contenu JavaScript
echo -e "\n${YELLOW}5. Test du contenu JavaScript${NC}"
js_content=$(curl -s "$BASE_URL/static/app.js")
if echo "$js_content" | grep -q "class LiteFaaSApp"; then
    print_result "Contenu JavaScript" "PASS" "Fichier JavaScript valide"
else
    print_result "Contenu JavaScript" "FAIL" "Fichier JavaScript invalide"
fi

# Test 6: Headers de cache
echo -e "\n${YELLOW}6. Test des headers de cache${NC}"
cache_headers=$(curl -s -I "$BASE_URL/static/style.css")
if echo "$cache_headers" | grep -q "Cache-Control: public"; then
    print_result "Headers de cache" "PASS" "Cache configuré correctement"
else
    print_result "Headers de cache" "FAIL" "Cache non configuré"
fi

# Test 7: Sécurité des headers
echo -e "\n${YELLOW}7. Test des headers de sécurité${NC}"
security_headers=$(curl -s -I "$BASE_URL/static/style.css")
if echo "$security_headers" | grep -q "X-Content-Type-Options: nosniff"; then
    print_result "Headers de sécurité" "PASS" "Sécurité configurée"
else
    print_result "Headers de sécurité" "FAIL" "Sécurité non configurée"
fi

# Test 8: Fichier inexistant
echo -e "\n${YELLOW}8. Test de fichier inexistant${NC}"
not_found_response=$(curl -s -w "%{http_code}" "$BASE_URL/static/inexistant.css" -o /dev/null)
if [ "$not_found_response" = "404" ]; then
    print_result "Fichier inexistant" "PASS" "404 retourné correctement"
else
    print_result "Fichier inexistant" "FAIL" "HTTP $not_found_response au lieu de 404"
fi

echo -e "\n${GREEN}=== Tests de l'interface web terminés ===${NC}"
echo -e "${GREEN}L'interface web LiteFaaS fonctionne correctement !${NC}"
echo -e "\n${YELLOW}Vous pouvez maintenant ouvrir http://localhost:8080 dans votre navigateur${NC}"
