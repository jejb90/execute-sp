package services

import (
	"database/sql"
	"fmt"
	"main/models"
	"os"
	"strings"

	_ "github.com/godror/godror"
)

// Database interface para manejar la conexión a la base de datos
type Database interface {
	Connect() error
	ExecuteStoredProcedure(call models.StoredProcedureCall) (map[string]interface{}, error)
}

// OracleDB estructura para manejar la conexión a la base de datos Oracle
type OracleDB struct {
	db *sql.DB
}

// NewOracleDB constructor para OracleDB
func NewOracleDB() *OracleDB {
	return &OracleDB{}
}

// Connect establece la conexión a la base de datos Oracle
func (db *OracleDB) Connect() error {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbService := os.Getenv("DB_SERVICE")

	connStr := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%s/%s"`, dbUser, dbPassword, dbHost, dbPort, dbService)
	database, err := sql.Open("godror", connStr)
	if err != nil {
		return fmt.Errorf("error al conectar a la base de datos: %v", err)
	}
	db.db = database
	return nil
}

// ExecuteStoredProcedure ejecuta un procedimiento almacenado con los parámetros proporcionados
func (db *OracleDB) ExecuteStoredProcedure(call models.StoredProcedureCall) (map[string]interface{}, error) {
	var placeholders []string
	for i := range call.Inputs {
		placeholders = append(placeholders, fmt.Sprintf(":p%d", i+1))
	}

	for i := 0; i < call.Outputs; i++ {
		placeholders = append(placeholders, fmt.Sprintf(":p%d", len(call.Inputs)+i+1))
	}
	query := fmt.Sprintf("BEGIN %s(%s); END;", call.ProcedureName, strings.Join(placeholders, ", "))

	fmt.Println(query)
	args := make([]interface{}, len(call.Inputs)+call.Outputs)
	for i, v := range call.Inputs {
		args[i] = v
	}
	for i := 0; i < call.Outputs; i++ {
		var dest string
		args[len(call.Inputs)+i] = sql.Out{Dest: &dest}
	}
	_, err := db.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	fmt.Println(args)
	result := make(map[string]interface{})
	for i := 0; i < call.Outputs; i++ {
		outputKey := fmt.Sprintf("output%d", i+1)
		result[outputKey] = *(args[len(call.Inputs)+i].(sql.Out).Dest.(*string))
	}
	return result, nil
}
