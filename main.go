package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/godror/godror"
	"github.com/joho/godotenv"
)

// StoredProcedureCall estructura para almacenar la llamada al procedimiento almacenado
type StoredProcedureCall struct {
	ProcedureName string        `json:"ProcedureName"`
	Inputs        []interface{} `json:"Inputs"`
	Outputs       int           `json:"Outputs"`
}

// DBConnection estructura para manejar la conexión a la base de datos
type DBConnection struct {
	db *sql.DB
}

func main() {
	// Cargar variables de entorno desde el archivo .env usando la librería godotenv
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error cargando el archivo .env: %v", err)
	}

	// Crear una conexión a la base de datos
	conn := DBConnection{}
	conn.Connect()

	// Endpoint para manejar la ejecución de procedimientos almacenados
	http.HandleFunc("/execute-procedure", conn.handleExecuteProcedure)

	// Configurar y lanzar el servidor HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto si no se especifica en .env
	}
	log.Printf("Servidor escuchando en el puerto %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Connect establece la conexión a la base de datos Oracle
func (conn *DBConnection) Connect() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbService := os.Getenv("DB_SERVICE")

	connStr := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%s/%s"`, dbUser, dbPassword, dbHost, dbPort, dbService)

	db, err := sql.Open("godror", connStr)
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}
	conn.db = db
}

// handleExecuteProcedure maneja las solicitudes HTTP para ejecutar procedimientos almacenados
func (conn *DBConnection) handleExecuteProcedure(w http.ResponseWriter, r *http.Request) {
	// Decodificar el cuerpo JSON de la solicitud
	var call StoredProcedureCall
	err := json.NewDecoder(r.Body).Decode(&call)
	if err != nil {
		http.Error(w, "Error al decodificar la solicitud JSON", http.StatusBadRequest)
		return
	}

	// Verificar que se proporcionó el nombre del procedimiento
	if call.ProcedureName == "" {
		http.Error(w, "Debe proporcionar el nombre del procedimiento", http.StatusBadRequest)
		return
	}

	// Ejecutar el procedimiento almacenado
	result, err := conn.executeStoredProcedure(call)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Responder con los resultados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// executeStoredProcedure ejecuta un procedimiento almacenado con los parámetros proporcionados
func (conn *DBConnection) executeStoredProcedure(call StoredProcedureCall) (map[string]interface{}, error) {
	// Construir la llamada al procedimiento usando un bloque PL/SQL anónimo
	var placeholders []string
	for i := range call.Inputs {
		placeholders = append(placeholders, fmt.Sprintf(":p%d", i+1))
	}

	for i := 0; i < call.Outputs; i++ {
		placeholders = append(placeholders, fmt.Sprintf(":p%d", len(call.Inputs)+i+1))
	}
	query := fmt.Sprintf("BEGIN %s(%s); END;", call.ProcedureName, strings.Join(placeholders, ", "))

	// Preparar los argumentos para ejecutar el procedimiento almacenado
	args := make([]interface{}, len(call.Inputs)+call.Outputs)
	for i, v := range call.Inputs {
		args[i] = v
	}
	for i := 0; i < call.Outputs; i++ {
		var dest string // Usamos un string como destino para el resultado
		args[len(call.Inputs)+i] = sql.Out{Dest: &dest}
	}
	// Ejecutar el procedimiento almacenado
	_, err := conn.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	// Recopilar resultados de salida en un mapa
	result := make(map[string]interface{})
	for i := 0; i < call.Outputs; i++ {
		outputKey := fmt.Sprintf("output%d", i+1)
		result[outputKey] = args[len(call.Inputs)+i].(sql.Out).Dest
	}
	return result, nil
}
