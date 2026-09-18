package comentarios

import "gorm.io/gorm"

// Sembrar carga datos iniciales de soporte comunitario
func Sembrar(db *gorm.DB) {
	var total int64
	db.Model(&Usuario{}).Count(&total)
	if total > 0 {
		return
	}

	datos := []Usuario{
		{
			Nombre:     "Jose Fuentes",
			Reputacion: 15,
			Comentarios: []Comentario{
				{Texto: "¿Cómo configuro las rutas en Chi?", Medalla: "Novato", Estado: "aprobado"},
			},
		},
		{
			Nombre:     "Luis Anchundia",
			Reputacion: 50,
			Comentarios: []Comentario{
				{Texto: "Usa chi.NewRouter() y define tus métodos con r.Get()", Medalla: "Expert", Estado: "destacado"},
				{Texto: "Recuerda validar los errores de GORM", Medalla: "Ayudante", Estado: "pendiente"},
			},
		},
	}

	db.Create(&datos)
}
