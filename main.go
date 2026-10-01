package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"COMUNITARIO/internal/comentarios" // Cambia si el módulo de tu go.mod es distinto

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	reset := flag.Bool("reset", false, "Borra y recrea la base de datos")
	flag.Parse()

	// 1. Conexión a PostgreSQL: los datos salen de las variables de entorno (archivo .env)
	godotenv.Load() // si no hay .env, usa las variables del sistema
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Fallo al conectar a la base de datos: ", err)
	}

	// 2. Resetear tablas si se pasa la bandera -reset
	if *reset {
		fmt.Println("Borrando tablas antiguas...")
		db.Migrator().DropTable(&comentarios.Comentario{}, &comentarios.Usuario{}, &comentarios.Negocio{})
	}

	// 3. Migración automática (Primero los padres, luego las hijas)
	fmt.Println("Migrando base de datos comunitaria...")
	db.AutoMigrate(&comentarios.Negocio{}, &comentarios.Usuario{}, &comentarios.Comentario{})

	// 4. Semilla de datos iniciales
	comentarios.Sembrar(db)

	// 5. Enrutador Chi
	r := chi.NewRouter()
	manejador := &comentarios.Manejador{DB: db}

	// Rutas del soporte comunitario
	r.Post("/comentarios", manejador.Crear)
	r.Get("/comentarios", manejador.Listar)
	r.Get("/comentarios/{id}", manejador.Obtener)
	r.Patch("/comentarios/{id}/estado", manejador.CambiarEstado)
	r.Delete("/comentarios/{id}", manejador.Borrar)

	// Ruta optimizada para evitar el problema N+1
	r.Get("/usuarios", manejador.ListarUsuarios)

	// 6. Ejecución del servidor
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	fmt.Println("Servidor comunitario corriendo en http://localhost:" + puerto)
	http.ListenAndServe(":"+puerto, r)
}
