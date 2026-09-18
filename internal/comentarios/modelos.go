package comentarios

// Usuario es la entidad del lado del "uno"
type Usuario struct {
	ID          uint
	Nombre      string
	Reputacion  int
	Comentarios []Comentario
}

// Comentario es la entidad del lado de los "muchos" y maneja estados
type Comentario struct {
	ID        uint
	UsuarioID uint
	Texto     string
	Medalla   string
	Estado    string // Ejemplo: pendiente, aprobado, destacado
}

// Estados válidos para el control de la comunidad (Fase 2b)
var estadosValidos = map[string]bool{
	"pendiente": true,
	"aprobado":  true,
	"destacado": true,
}
