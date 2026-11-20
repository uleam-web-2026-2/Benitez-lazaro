package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"COMUNITARIO/internal/api"
	"COMUNITARIO/internal/auth"
	"COMUNITARIO/internal/comentarios/handler"
	"COMUNITARIO/internal/comentarios/repositorio"
	"COMUNITARIO/internal/comentarios/servicio"

	"github.com/joho/godotenv"
)

func main() {
	reset := flag.Bool("reset", false, "Borra y recrea las tablas")
	migrar := flag.Bool("migrar", false, "Crea tablas, semilla y permisos con el usuario dueño, y termina")
	flag.Parse()

	// 1. Configuración fuera del código (.env). Nada de claves en el repositorio.
	godotenv.Load()
	secreto := os.Getenv("JWT_SECRET")
	if secreto == "" {
		log.Fatal("Falta JWT_SECRET en el archivo .env")
	}
	db := abrir(*migrar, *reset)

	// 3. Se arman las capas: repositorio → servicio → handler → router.
	repoGorm := repositorio.Nuevo(db)
	conexion, err := db.DB() // la conexión de database/sql que usa sqlc
	if err != nil {
		log.Fatal("no se pudo obtener la conexión: ", err)
	}
	// Taller 10: el servicio recibe el adaptador sqlc (hereda de GORM lo demás).
	// Para volver a GORM puro basta cambiar a: servicio.Nuevo(repoGorm)
	repo := repositorio.NuevoSqlc(repoGorm, conexion)
	h := &handler.Manejador{Servicio: servicio.Nuevo(repo)}
	a := &auth.Manejador{Cuentas: repoGorm, Secreto: []byte(secreto), Vida: time.Hour}

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	fmt.Println("Servidor comunitario corriendo en http://localhost:" + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, api.Router(h, a, []byte(secreto))))
}
