package handler

import (
	"time"

	"COMUNITARIO/internal/modelos"
)

// ComentarioPeticion es lo único que el cliente puede mandar al crear.
// No lleva id, estado, autor ni moderador: los pone el servidor.
type ComentarioPeticion struct {
	Texto       string `json:"texto"`
	CategoriaID *uint  `json:"categoria_id"`
}

// CambioEstadoPeticion es el cuerpo del PATCH de estado.
type CambioEstadoPeticion struct {
	Estado string `json:"estado"`
}

// ComentarioRespuesta es lo que el servidor devuelve de un comentario.
type ComentarioRespuesta struct {
	ID          uint      `json:"id"`
	UsuarioID   uint      `json:"usuario_id"`
	CategoriaID *uint     `json:"categoria_id"`
	Texto       string    `json:"texto"`
	Estado      string    `json:"estado"`
	ModeradorID *uint     `json:"moderador_id"`
	Creado      time.Time `json:"creado"`
	Actualizado time.Time `json:"actualizado"`
}

func nuevaRespuesta(c modelos.Comentario) ComentarioRespuesta {
	return ComentarioRespuesta{
		ID: c.ID, UsuarioID: c.UsuarioID, CategoriaID: c.CategoriaID, Texto: c.Texto,
		Estado: c.Estado, ModeradorID: c.ModeradorID, Creado: c.Creado, Actualizado: c.Actualizado,
	}
}

func nuevasRespuestas(lista []modelos.Comentario) []ComentarioRespuesta {
	salida := make([]ComentarioRespuesta, 0, len(lista))
	for _, c := range lista {
		salida = append(salida, nuevaRespuesta(c))
	}
	return salida
}

// UsuarioRespuesta: el perfil con sus comentarios y medallas.
type UsuarioRespuesta struct {
	ID          uint                     `json:"id"`
	NegocioID   uint                     `json:"negocio_id"`
	Nombre      string                   `json:"nombre"`
	Correo      string                   `json:"correo"`
	Rol         string                   `json:"rol"`
	Reputacion  int                      `json:"reputacion"`
	Creado      time.Time                `json:"creado"`
	Comentarios []ComentarioRespuesta    `json:"comentarios,omitempty"`
	Medallas    []modelos.UsuarioMedalla `json:"medallas,omitempty"`
}

func nuevasRespuestasUsuario(lista []modelos.Usuario) []UsuarioRespuesta {
	salida := make([]UsuarioRespuesta, 0, len(lista))
	for _, u := range lista {
		salida = append(salida, UsuarioRespuesta{
			ID: u.ID, NegocioID: u.NegocioID, Nombre: u.Nombre, Correo: u.Correo, Rol: u.Rol,
			Reputacion: u.Reputacion, Creado: u.Creado,
			Comentarios: nuevasRespuestas(u.Comentarios), Medallas: u.Medallas,
		})
	}
	return salida
}
