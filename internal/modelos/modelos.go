// Package modelos contiene los structs de TODAS las tablas del proyecto.
// No sabe nada de HTTP ni de reglas: solo describe los datos.
package modelos

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrNoEncontrado lo devuelve el repositorio cuando no existe el registro.
var ErrNoEncontrado = errors.New("registro no encontrado")

// Estados del comentario (máquina de estados).
const (
	EstadoPendiente = "pendiente"
	EstadoAprobado  = "aprobado"
	EstadoDestacado = "destacado"
)

// Roles de la comunidad.
const (
	RolMiembro   = "miembro"
	RolModerador = "moderador"
)

// Estados de pago del negocio (pago por transferencia).
const (
	PagoPorVerificar = "pago_por_verificar"
	PagoActiva       = "activa"
	PagoVencida      = "vencida"
)

// 1. Negocio: el cliente que paga la suscripción mensual.
type Negocio struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Nombre     string    `gorm:"not null" json:"nombre"`
	RUC        string    `gorm:"not null;size:13;uniqueIndex" json:"ruc"`
	EstadoPago string    `gorm:"not null;default:pago_por_verificar" json:"estado_pago"`
	Creado     time.Time `json:"creado"`
	Usuarios   []Usuario `json:"usuarios,omitempty"`
}

// 2. Usuario: persona de la comunidad de un negocio (miembro o moderador).
type Usuario struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	NegocioID  uint      `gorm:"not null;index" json:"negocio_id"`
	Nombre     string    `gorm:"not null" json:"nombre"`
	Correo     string    `gorm:"not null;uniqueIndex" json:"correo"`
	Rol        string    `gorm:"not null;default:miembro" json:"rol"`
	Reputacion int       `gorm:"not null;default:0" json:"reputacion"`
	Creado     time.Time `json:"creado"`

	Comentarios []Comentario     `json:"comentarios,omitempty"`
	Puntos      []Punto          `json:"puntos,omitempty"`
	Medallas    []UsuarioMedalla `json:"medallas,omitempty"`
}

// 3. Cuenta: SOLO para iniciar sesión. Una cuenta por usuario.
type Cuenta struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	Login     string  `gorm:"not null;uniqueIndex" json:"login"`
	ClaveHash string  `gorm:"not null" json:"-"` // nunca la clave en claro
	UsuarioID uint    `gorm:"not null;uniqueIndex" json:"usuario_id"`
	Persona   Usuario `gorm:"foreignKey:UsuarioID;references:ID" json:"-"`
}

// 4. Categoria: tema de la ayuda (técnica, pagos, uso de la app...).
type Categoria struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Nombre      string `gorm:"not null;uniqueIndex" json:"nombre"`
	Descripcion string `json:"descripcion"`

	Comentarios []Comentario `json:"comentarios,omitempty"`
}

// 5. Comentario (corregida): ya no guarda la medalla como texto;
// tiene categoría, quién lo moderó, fecha de cambio y borrado lógico.
// Taller 10: sin etiquetas json. Lo que entra y sale por HTTP son los DTO
// de internal/comentarios/handler/dtos.go.
type Comentario struct {
	ID          uint   `gorm:"primaryKey"`
	UsuarioID   uint   `gorm:"not null;index"`
	CategoriaID *uint  `gorm:"index"`
	Texto       string `gorm:"not null;size:1000"`
	Estado      string `gorm:"not null;default:pendiente;index"`
	ModeradorID *uint
	Creado      time.Time
	Actualizado time.Time
	Eliminado   gorm.DeletedAt `gorm:"index"`
}

// 6. Punto: libro de movimientos. Cada fila es una suma de puntos.
type Punto struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UsuarioID    uint      `gorm:"not null;index" json:"usuario_id"`
	ComentarioID *uint     `gorm:"index" json:"comentario_id"`
	Cantidad     int       `gorm:"not null" json:"cantidad"`
	Motivo       string    `gorm:"not null" json:"motivo"`
	Creado       time.Time `json:"creado"`
}

// 7. Medalla: catálogo de medallas simbólicas.
type Medalla struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Nombre        string `gorm:"not null;uniqueIndex" json:"nombre"`
	PuntosMinimos int    `gorm:"not null" json:"puntos_minimos"`
	Descripcion   string `json:"descripcion"`
}

// 8. UsuarioMedalla: qué medalla ganó cada usuario (una sola vez).
type UsuarioMedalla struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UsuarioID    uint      `gorm:"not null;uniqueIndex:idx_usuario_medalla" json:"usuario_id"`
	MedallaID    uint      `gorm:"not null;uniqueIndex:idx_usuario_medalla" json:"medalla_id"`
	ComentarioID *uint     `json:"comentario_id"`
	Otorgada     time.Time `json:"otorgada"`
	Medalla      Medalla   `gorm:"foreignKey:MedallaID" json:"medalla"`
}

// Todas devuelve las tablas en orden de migración (padres primero).
func Todas() []any {
	return []any{&Negocio{}, &Usuario{}, &Cuenta{}, &Categoria{}, &Comentario{}, &Punto{}, &Medalla{}, &UsuarioMedalla{}}
}

// TodasAlReves sirve para borrar tablas (hijas primero).
func TodasAlReves() []any {
	t := Todas()
	for i, j := 0, len(t)-1; i < j; i, j = i+1, j-1 {
		t[i], t[j] = t[j], t[i]
	}
	return t
}
