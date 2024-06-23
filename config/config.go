package config

import (
	"os"
)

// GetPort obtiene el puerto del entorno o devuelve el valor por defecto
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto si no se especifica en .env
	}
	return port
}
