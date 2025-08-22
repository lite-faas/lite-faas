class LiteFaaSApp {
    constructor() {
        this.functions = [];
        this.currentFunction = null;
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadFunctions();
    }

    bindEvents() {
        document.getElementById("newFunctionBtn").addEventListener("click", () => {
            this.showModal();
        });

        document.getElementById("functionForm").addEventListener("submit", e => {
            e.preventDefault();
            this.saveFunction();
        });

        document.querySelector(".close").addEventListener("click", () => {
            this.closeModal();
        });

        window.addEventListener("click", e => {
            if (e.target.classList.contains("modal")) {
                this.closeModal();
            }
        });
    }

    async loadFunctions() {
        try {
            const response = await fetch("/api/functions");
            if (!response.ok) {
                throw new Error("Failed to load functions");
            }

            this.functions = await response.json();
            this.renderFunctions();
        } catch (error) {
            this.showError("Erreur lors du chargement des fonctions: " + error.message);
        }
    }

    renderFunctions() {
        const container = document.getElementById("functionsContainer");

        if (this.functions.length === 0) {
            container.innerHTML = '<div class="loading">Aucune fonction trouvée</div>';
            return;
        }

        container.innerHTML = this.functions.map(func => this.renderFunctionCard(func)).join("");
    }

    renderFunctionCard(func) {
        const statusClass = func.status === "running" ? "status-running" : "status-stopped";
        const statusText = func.status === "running" ? "En cours" : "Arrêtée";

        return `
            <div class="function-card" data-function-name="${func.name}">
                <div class="function-header">
                    <div class="function-name">${func.name}</div>
                    <div class="function-status ${statusClass}">${statusText}</div>
                </div>
                <div class="function-details">
                    <div class="function-detail">
                        <div class="detail-label">Langage</div>
                        <div class="detail-value">${func.language}</div>
                    </div>
                    <div class="function-detail">
                        <div class="detail-label">Port</div>
                        <div class="detail-value">${func.port || "Non assigné"}</div>
                    </div>
                    <div class="function-detail">
                        <div class="detail-label">ID Conteneur</div>
                        <div class="detail-value">${func.container_id || "Aucun"}</div>
                    </div>
                </div>
                <div class="function-actions">
                    ${
                        func.status === "running"
                            ? `<button class="btn btn-danger" onclick="app.stopFunction('${func.name}')">Arrêter</button>`
                            : `<button class="btn btn-success" onclick="app.startFunction('${func.name}')">Démarrer</button>`
                    }
                    <button class="btn btn-secondary" onclick="app.editFunction('${
                        func.name
                    }')">Modifier</button>
                    <button class="btn btn-danger" onclick="app.deleteFunction('${
                        func.name
                    }')">Supprimer</button>
                </div>
            </div>
        `;
    }

    showModal(functionName = null) {
        const modal = document.getElementById("functionModal");
        const title = document.getElementById("modalTitle");
        const form = document.getElementById("functionForm");
        const nameInput = document.getElementById("functionName");
        const languageSelect = document.getElementById("functionLanguage");
        const codeTextarea = document.getElementById("functionCode");

        if (functionName) {
            this.currentFunction = this.functions.find(f => f.name === functionName);
            title.textContent = "Modifier la Fonction";
            nameInput.value = this.currentFunction.name;
            nameInput.disabled = true;
            languageSelect.value = this.currentFunction.language;
            codeTextarea.value = this.currentFunction.code;
        } else {
            this.currentFunction = null;
            title.textContent = "Nouvelle Fonction";
            form.reset();
            nameInput.disabled = false;
        }

        modal.style.display = "block";
    }

    closeModal() {
        const modal = document.getElementById("functionModal");
        modal.style.display = "none";
        this.currentFunction = null;
    }

    async saveFunction() {
        const formData = new FormData(document.getElementById("functionForm"));
        const functionData = {
            name: formData.get("name"),
            language: formData.get("language"),
            code: formData.get("code"),
        };

        try {
            let response;
            if (this.currentFunction) {
                response = await fetch(`/api/functions/${this.currentFunction.name}`, {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({
                        language: functionData.language,
                        code: functionData.code,
                    }),
                });
            } else {
                response = await fetch("/api/functions", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(functionData),
                });
            }

            if (!response.ok) {
                throw new Error("Failed to save function");
            }

            this.closeModal();
            this.loadFunctions();
        } catch (error) {
            this.showError("Erreur lors de la sauvegarde: " + error.message);
        }
    }

    async startFunction(functionName) {
        try {
            const response = await fetch(`/api/functions/${functionName}/start`, {
                method: "POST",
            });

            if (!response.ok) {
                throw new Error("Failed to start function");
            }

            this.loadFunctions();
        } catch (error) {
            this.showError("Erreur lors du démarrage: " + error.message);
        }
    }

    async stopFunction(functionName) {
        try {
            const response = await fetch(`/api/functions/${functionName}/stop`, {
                method: "POST",
            });

            if (!response.ok) {
                throw new Error("Failed to stop function");
            }

            this.loadFunctions();
        } catch (error) {
            this.showError("Erreur lors de l'arrêt: " + error.message);
        }
    }

    editFunction(functionName) {
        this.showModal(functionName);
    }

    async deleteFunction(functionName) {
        if (!confirm(`Êtes-vous sûr de vouloir supprimer la fonction "${functionName}" ?`)) {
            return;
        }

        try {
            const response = await fetch(`/api/functions/${functionName}`, {
                method: "DELETE",
            });

            if (!response.ok) {
                throw new Error("Failed to delete function");
            }

            this.loadFunctions();
        } catch (error) {
            this.showError("Erreur lors de la suppression: " + error.message);
        }
    }

    showError(message) {
        const container = document.getElementById("functionsContainer");
        const errorDiv = document.createElement("div");
        errorDiv.className = "error";
        errorDiv.textContent = message;

        container.insertBefore(errorDiv, container.firstChild);

        setTimeout(() => {
            errorDiv.remove();
        }, 5000);
    }
}

// Initialiser l'application quand le DOM est chargé
document.addEventListener("DOMContentLoaded", () => {
    window.app = new LiteFaaSApp();
});

// Fonction globale pour fermer le modal
function closeModal() {
    if (window.app) {
        window.app.closeModal();
    }
}
