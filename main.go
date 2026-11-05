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
	"COMUNITARIO/internal/modelos"
	"COMUNITARIO/internal/semilla"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	reset := flag.Bool("reset", false, "Borra y recrea las tablas")
	flag.Parse()

	// 1. Configuración fuera del código (.env). Nada de claves en el repositorio.
	godotenv.Load()
	secreto := os.Getenv("JWT_SECRET")
	if secreto == "" {
		log.Fatal("Falta JWT_SECRET en el archivo .env")
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		log.Fatal("Fallo al conectar a la base de datos: ", err)
	}

	// 2. Tablas: con -reset se borran todas y se recrean.
	if *reset {
		fmt.Println("Borrando tablas antiguas...")
		if err := db.Migrator().DropTable(modelos.TodasAlReves()...); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Migrando las 8 tablas...")
	if err := db.AutoMigrate(modelos.Todas()...); err != nil {
		log.Fatal("Fallo la migración: ", err)
	}
	if err := semilla.Sembrar(db); err != nil {
		log.Fatal("Fallo la semilla: ", err)
	}

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
