package comentarios

import (
	"time"

	"gorm.io/gorm"
)

// Sembrar carga datos iniciales de soporte comunitario
func Sembrar(db *gorm.DB) {
	var total int64
	db.Model(&Negocio{}).Count(&total)
	if total > 0 {
		return
	}

	negocio := Negocio{Nombre: "Tienda Manta Tech", RUC: "1391234567001", EstadoPago: PagoActiva, Creado: time.Now()}
	db.Create(&negocio)

	datos := []Usuario{
		{
			NegocioID:  negocio.ID,
			Nombre:     "Jose Fuentes",
			Correo:     "jose@ejemplo.ec",
			Rol:        RolMiembro,
			Reputacion: 15,
			Comentarios: []Comentario{
				{Texto: "¿Cómo configuro las rutas en Chi?", Medalla: "Novato", Estado: EstadoAprobado, Creado: time.Now()},
			},
		},
		{
			NegocioID:  negocio.ID,
			Nombre:     "Luis Anchundia",
			Correo:     "luis@ejemplo.ec",
			Rol:        RolModerador,
			Reputacion: 50,
			Comentarios: []Comentario{
				{Texto: "Usa chi.NewRouter() y define tus métodos con r.Get()", Medalla: "Expert", Estado: EstadoDestacado, Creado: time.Now()},
				{Texto: "Recuerda validar los errores de GORM", Medalla: "Ayudante", Estado: EstadoPendiente, Creado: time.Now()},
			},
		},
	}

	db.Create(&datos)
}
