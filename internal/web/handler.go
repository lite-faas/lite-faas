package web

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	logger *logrus.Logger
}

func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/web")

	if path == "" || path == "/" {
		path = "/index.html"
	}

	content, contentType := h.getStaticContent(path)
	if content == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}

func (h *Handler) getStaticContent(path string) ([]byte, string) {
	switch path {
	case "/index.html":
		return []byte(htmlContent), "text/html"
	case "/style.css":
		return []byte(cssContent), "text/css"
	case "/script.js":
		return []byte(jsContent), "application/javascript"
	default:
		return nil, ""
	}
}

const htmlContent = `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LiteFaaS - Gestionnaire de Fonctions</title>
    <link rel="stylesheet" href="/web/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LiteFaaS</h1>
            <p>Plateforme légère de fonctions serverless</p>
        </header>

        <main>
            <section class="functions-list">
                <h2>Fonctions</h2>
                <button id="newFunctionBtn" class="btn btn-primary">Nouvelle Fonction</button>
                <div id="functionsList"></div>
            </section>

            <section class="function-editor" id="functionEditor" style="display: none;">
                <h2>Éditer la Fonction</h2>
                <form id="functionForm">
                    <div class="form-group">
                        <label for="functionName">Nom:</label>
                        <input type="text" id="functionName" required>
                    </div>
                    <div class="form-group">
                        <label for="functionLanguage">Langage:</label>
                        <select id="functionLanguage" required>
                            <option value="python">Python</option>
                            <option value="nodejs">Node.js</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="functionCode">Code:</label>
                        <textarea id="functionCode" rows="10" required></textarea>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Sauvegarder</button>
                        <button type="button" class="btn btn-secondary" onclick="cancelEdit()">Annuler</button>
                    </div>
                </form>
            </section>
        </main>
    </div>

    <script src="/web/script.js"></script>
</body>
</html>`

const cssContent = `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
    background-color: #f5f5f5;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    text-align: center;
    margin-bottom: 40px;
    padding: 20px;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

header h1 {
    color: #2563eb;
    margin-bottom: 10px;
}

.functions-list {
    background: white;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    margin-bottom: 20px;
}

.btn {
    padding: 10px 20px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    margin: 5px;
}

.btn-primary {
    background-color: #2563eb;
    color: white;
}

.btn-secondary {
    background-color: #6b7280;
    color: white;
}

.btn-danger {
    background-color: #dc2626;
    color: white;
}

.btn-success {
    background-color: #059669;
    color: white;
}

.function-item {
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    padding: 15px;
    margin: 10px 0;
    background: #f9fafb;
}

.function-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
}

.function-name {
    font-weight: bold;
    color: #2563eb;
}

.function-status {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: bold;
}

.status-running {
    background-color: #d1fae5;
    color: #065f46;
}

.status-stopped {
    background-color: #fee2e2;
    color: #991b1b;
}

.function-actions {
    display: flex;
    gap: 10px;
}

.function-editor {
    background: white;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.form-group {
    margin-bottom: 15px;
}

.form-group label {
    display: block;
    margin-bottom: 5px;
    font-weight: bold;
}

.form-group input,
.form-group select,
.form-group textarea {
    width: 100%;
    padding: 8px;
    border: 1px solid #d1d5db;
    border-radius: 4px;
    font-size: 14px;
}

.form-group textarea {
    resize: vertical;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.form-actions {
    display: flex;
    gap: 10px;
    margin-top: 20px;
}`

const jsContent = `let functions = [];
let editingFunction = null;

document.addEventListener('DOMContentLoaded', function() {
    loadFunctions();
    document.getElementById('newFunctionBtn').addEventListener('click', showNewFunctionForm);
    document.getElementById('functionForm').addEventListener('submit', saveFunction);
});

async function loadFunctions() {
    try {
        const response = await fetch('/api/functions');
        functions = await response.json();
        renderFunctions();
    } catch (error) {
        console.error('Erreur lors du chargement des fonctions:', error);
    }
}

function renderFunctions() {
    const container = document.getElementById('functionsList');
    container.innerHTML = '';

    functions.forEach(function(fn) {
        const div = document.createElement('div');
        div.className = 'function-item';
        div.innerHTML = '<div class="function-header"><span class="function-name">' + fn.name + '</span><span class="function-status status-' + fn.status + '">' + fn.status + '</span></div><div>Langage: ' + fn.language + '</div><div>Port: ' + fn.port + '</div><div class="function-actions"><button class="btn btn-primary" onclick="editFunction(\'' + fn.name + '\')">Éditer</button><button class="btn btn-success" onclick="startFunction(\'' + fn.name + '\')">Démarrer</button><button class="btn btn-secondary" onclick="stopFunction(\'' + fn.name + '\')">Arrêter</button><button class="btn btn-danger" onclick="deleteFunction(\'' + fn.name + '\')">Supprimer</button></div>';
        container.appendChild(div);
    });
}

function showNewFunctionForm() {
    editingFunction = null;
    document.getElementById('functionName').value = '';
    document.getElementById('functionLanguage').value = 'python';
    document.getElementById('functionCode').value = '';
    document.getElementById('functionEditor').style.display = 'block';
}

function editFunction(name) {
    const fn = functions.find(f => f.name === name);
    if (!fn) return;

    editingFunction = fn;
    document.getElementById('functionName').value = fn.name;
    document.getElementById('functionLanguage').value = fn.language;
    document.getElementById('functionCode').value = fn.code;
    document.getElementById('functionEditor').style.display = 'block';
}

function cancelEdit() {
    document.getElementById('functionEditor').style.display = 'none';
    editingFunction = null;
}

async function saveFunction(e) {
    e.preventDefault();

    const functionData = {
        name: document.getElementById('functionName').value,
        language: document.getElementById('functionLanguage').value,
        code: document.getElementById('functionCode').value
    };

    try {
        const url = editingFunction ? '/api/functions/' + editingFunction.name : '/api/functions';
        const method = editingFunction ? 'PUT' : 'POST';

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(functionData)
        });

        if (response.ok) {
            cancelEdit();
            loadFunctions();
        } else {
            alert('Erreur lors de la sauvegarde');
        }
    } catch (error) {
        console.error('Erreur:', error);
        alert('Erreur lors de la sauvegarde');
    }
}

async function startFunction(name) {
    try {
        const response = await fetch('/api/functions/' + name + '/start', {
            method: 'POST'
        });

        if (response.ok) {
            loadFunctions();
        }
    } catch (error) {
        console.error('Erreur lors du démarrage:', error);
    }
}

async function stopFunction(name) {
    try {
        const response = await fetch('/api/functions/' + name + '/stop', {
            method: 'POST'
        });

        if (response.ok) {
            loadFunctions();
        }
    } catch (error) {
        console.error('Erreur lors de l\'arrêt:', error);
    }
}

async function deleteFunction(name) {
    if (!confirm('Êtes-vous sûr de vouloir supprimer cette fonction ?')) {
        return;
    }

    try {
        const response = await fetch('/api/functions/' + name, {
            method: 'DELETE'
        });

        if (response.ok) {
            loadFunctions();
        }
    } catch (error) {
        console.error('Erreur lors de la suppression:', error);
    }
}`
