package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"COMUNITARIO/internal/modelos"
	"COMUNITARIO/internal/semilla"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ident(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }
func lit(s string) string   { return `'` + strings.ReplaceAll(s, `'`, `''`) + `'` }

// abrir conecta a la base. La API normal usa el usuario de pocos permisos
// (APP_DB_*). Solo con -migrar se usa el dueño de las tablas (DB_*): crea las
// tablas, la semilla y los permisos, y el programa termina.
func abrir(migrar, reset bool) *gorm.DB {
	usuario, clave := os.Getenv("APP_DB_USER"), os.Getenv("APP_DB_PASSWORD")
	if migrar {
		usuario, clave = os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")
	} else if usuario == "" || clave == "" {
		log.Fatal("Faltan APP_DB_USER y APP_DB_PASSWORD en el archivo .env")
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), usuario, clave, os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		log.Fatal("Fallo al conectar a la base de datos: ", err)
	}
	if !migrar {
		return db
	}
	if reset {
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
	if err := darPermisos(db, os.Getenv("APP_DB_USER"), os.Getenv("APP_DB_PASSWORD")); err != nil {
		log.Fatal("Fallo al dar permisos: ", err)
	}
	fmt.Println("Migración lista. Ahora se arranca la API sin -migrar.")
	os.Exit(0)
	return db
}

// darPermisos crea el usuario de la API (sin superusuario, sin crear bases ni
// roles, y que no es dueño de nada) y le da solo lo que el código usa.
func darPermisos(db *gorm.DB, usuario, clave string) error {
	if usuario == "" || clave == "" {
		return fmt.Errorf("faltan APP_DB_USER y APP_DB_PASSWORD")
	}
	u := ident(usuario)
	var existe int64
	db.Raw("SELECT count(*) FROM pg_roles WHERE rolname = ?", usuario).Scan(&existe)
	if existe == 0 {
		if err := db.Exec("CREATE ROLE " + u + " LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE").Error; err != nil {
			return err
		}
	}
	sentencias := []string{
		"ALTER ROLE " + u + " PASSWORD " + lit(clave),
		"REVOKE ALL ON ALL TABLES IN SCHEMA public FROM " + u,
		"REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM " + u,
		// Solo se leen:
		"GRANT SELECT ON negocios, cuenta, categoria, medallas TO " + u,
		// Se lee y se cambia la reputación:
		"GRANT SELECT, UPDATE ON usuarios TO " + u,
		// Se crean, se leen, se cambian y se borran:
		"GRANT SELECT, INSERT, UPDATE, DELETE ON comentarios TO " + u,
		// Solo se agregan filas:
		"GRANT SELECT, INSERT ON puntos, usuario_medallas TO " + u,
		// Para que INSERT pueda usar el contador de id:
		"GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO " + u,
	}
	for _, s := range sentencias {
		if err := db.Exec(s).Error; err != nil {
			return fmt.Errorf("%s: %w", s, err)
		}
	}
	return nil
}
