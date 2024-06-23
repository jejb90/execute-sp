package handlers

import (
	"encoding/json"
	"main/models"
	"main/services"
	"net/http"
)

// HTTPHandler estructura para manejar solicitudes HTTP
type HTTPHandler struct {
	db services.Database
}

// NewHTTPHandler constructor para HTTPHandler
func NewHTTPHandler(db services.Database) *HTTPHandler {
	return &HTTPHandler{db: db}
}

// HandleExecuteProcedure maneja las solicitudes HTTP para ejecutar procedimientos almacenados
func (h *HTTPHandler) HandleExecuteProcedure(w http.ResponseWriter, r *http.Request) {
	var call models.StoredProcedureCall
	err := json.NewDecoder(r.Body).Decode(&call)
	if err != nil {
		http.Error(w, "Error al decodificar la solicitud JSON", http.StatusBadRequest)
		return
	}

	if call.ProcedureName == "" {
		http.Error(w, "Debe proporcionar el nombre del procedimiento", http.StatusBadRequest)
		return
	}

	result, err := h.db.ExecuteStoredProcedure(call)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
