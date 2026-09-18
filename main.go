package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"COMUNITARIO/internal/comentarios" // Cambia si el módulo de tu go.mod es distinto

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	reset := flag.Bool("reset", false, "Borra y recrea la base de datos")
	flag.Parse()

	// 1. Conexión a PostgreSQL (Modifica usuario o contraseña si lo requieres)
	dsn := "host=localhost user=postgres password=2809 dbname=APP-COMUNITARIA port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Fallo al conectar a la base de datos: ", err)
	}

	// 2. Resetear tablas si se pasa la bandera -reset
	if *reset {
		fmt.Println("Borrando tablas antiguas...")
		db.Migrator().DropTable(&comentarios.Comentario{}, &comentarios.Usuario{})
	}

	// 3. Migración automática (Primero el padre, luego la hija)
	fmt.Println("Migrando base de datos comunitaria...")
	db.AutoMigrate(&comentarios.Usuario{}, &comentarios.Comentario{})

	// 4. Semilla de datos iniciales
	comentarios.Sembrar(db)

	// 5. Enrutador Chi
	r := chi.NewRouter()
	manejador := &comentarios.Manejador{DB: db}

	// Rutas del soporte comunitario
	r.Post("/comentarios", manejador.Crear)
	r.Get("/comentarios", manejador.Listar)
	r.Put("/comentarios/{id}", manejador.Actualizar)
	r.Delete("/comentarios/{id}", manejador.Borrar)

	// Ruta optimizada para evitar el problema N+1
	r.Get("/usuarios", manejador.ListarUsuarios)

	// 6. Ejecución del servidor
	fmt.Println("Servidor comunitario corriendo en http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
