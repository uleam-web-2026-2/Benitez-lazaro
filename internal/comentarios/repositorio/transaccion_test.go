package repositorio

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"COMUNITARIO/internal/comentarios/servicio"
	"COMUNITARIO/internal/modelos"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var errForzado = errors.New("error forzado a mitad de la transacción")

func abrirBD(t *testing.T) *gorm.DB {
	t.Helper()
	godotenv.Load("../../../.env")
	if os.Getenv("DB_HOST") == "" {
		t.Skip("sin .env: se omite la prueba contra PostgreSQL")
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skip("PostgreSQL no disponible: ", err)
	}
	if err := db.AutoMigrate(modelos.Todas()...); err != nil {
		t.Skip("no se pudo migrar: ", err)
	}
	return db
}

// Cada adaptador debe cumplir: si algo falla a la mitad, NO queda nada guardado.
func TestTransaccionAtomica(t *testing.T) {
	db := abrirBD(t)
	conexion, _ := db.DB()
	gorm1 := Nuevo(db)
	casos := map[string]servicio.Repositorio{
		"gorm": gorm1,
		"sqlc": NuevoSqlc(gorm1, conexion),
	}
	for nombre, repo := range casos {
		t.Run(nombre, func(t *testing.T) {
			neg := modelos.Negocio{Nombre: "T10 " + nombre, RUC: fmt.Sprintf("T%012d", time.Now().UnixNano()%1e12), EstadoPago: modelos.PagoActiva, Creado: time.Now()}
			if err := db.Create(&neg).Error; err != nil {
				t.Fatal(err)
			}
			usr := modelos.Usuario{NegocioID: neg.ID, Nombre: "t10", Correo: fmt.Sprintf("t10-%s-%d@x.com", nombre, time.Now().UnixNano()), Rol: modelos.RolMiembro, Creado: time.Now()}
			db.Create(&usr)
			com := modelos.Comentario{UsuarioID: usr.ID, Texto: "prueba", Estado: modelos.EstadoAprobado, Creado: time.Now(), Actualizado: time.Now()}
			db.Create(&com)
			t.Cleanup(func() {
				db.Unscoped().Where("usuario_id = ?", usr.ID).Delete(&modelos.Punto{})
				db.Unscoped().Delete(&com)
				db.Delete(&usr)
				db.Delete(&neg)
			})

			// 1) Fallo a mitad: guarda el comentario, suma reputación y luego falla.
			err := repo.EnTransaccion(func(tx servicio.Repositorio) error {
				c, err := tx.ObtenerComentario(com.ID)
				if err != nil {
					return err
				}
				c.Estado = modelos.EstadoDestacado
				c.Actualizado = time.Now()
				if err := tx.GuardarComentario(c); err != nil {
					return err
				}
				if _, err := tx.SumarReputacion(usr.ID, 10); err != nil {
					return err
				}
				return errForzado
			})
			if !errors.Is(err, errForzado) {
				t.Fatalf("esperaba el error forzado, llegó: %v", err)
			}
			c, _ := repo.ObtenerComentario(com.ID)
			if c.Estado != modelos.EstadoAprobado {
				t.Errorf("el estado quedó en %q: la transacción NO fue atómica", c.Estado)
			}
			var u modelos.Usuario
			db.First(&u, usr.ID)
			if u.Reputacion != 0 {
				t.Errorf("la reputación quedó en %d: debía volver a 0", u.Reputacion)
			}

			// 2) Camino feliz: sin error, todo se guarda.
			err = repo.EnTransaccion(func(tx servicio.Repositorio) error {
				c, _ := tx.ObtenerComentario(com.ID)
				c.Estado = modelos.EstadoDestacado
				c.Actualizado = time.Now()
				if err := tx.GuardarComentario(c); err != nil {
					return err
				}
				_, err := tx.SumarReputacion(usr.ID, 10)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			c, _ = repo.ObtenerComentario(com.ID)
			db.First(&u, usr.ID)
			if c.Estado != modelos.EstadoDestacado || u.Reputacion != 10 {
				t.Errorf("commit fallido: estado=%q reputación=%d", c.Estado, u.Reputacion)
			}
		})
	}
}

func TestNoEncontradoSqlc(t *testing.T) {
	db := abrirBD(t)
	conexion, _ := db.DB()
	repo := NuevoSqlc(Nuevo(db), conexion)
	if _, err := repo.ObtenerComentario(999999999); !errors.Is(err, modelos.ErrNoEncontrado) {
		t.Fatalf("sql.ErrNoRows debía traducirse a ErrNoEncontrado, llegó: %v", err)
	}
}
