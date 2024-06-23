package main

import (
	"github.com/joho/godotenv"
	"log"
	"main/config"
	"main/handlers"
	"main/services"
	"net/http"
)

func main() {
	// Cargar variables de entorno desde el archivo .env usando la librería godotenv
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error cargando el archivo .env: %v", err)
	}

	// Crear una conexión a la base de datos
	db := services.NewOracleDB()
	err = db.Connect()
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}

	// Crear el manejador HTTP
	handler := handlers.NewHTTPHandler(db)

	// Endpoint para manejar la ejecución de procedimientos almacenados
	http.HandleFunc("/execute-procedure", handler.HandleExecuteProcedure)

	// Configurar y lanzar el servidor HTTP
	port := config.GetPort()
	log.Printf("Servidor escuchando en el puerto %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
