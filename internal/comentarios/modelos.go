package comentarios

import "time"

// Estados de un comentario (máquina de estados)
const (
	EstadoPendiente = "pendiente"
	EstadoAprobado  = "aprobado"
	EstadoDestacado = "destacado"
)

// Roles de la comunidad
const (
	RolMiembro   = "miembro"
	RolModerador = "moderador"
)

// Estados de pago de la suscripción del negocio (pago por transferencia)
const (
	PagoPorVerificar = "pago_por_verificar"
	PagoActiva       = "activa"
	PagoVencida      = "vencida"
)

// PuntosDestacado es la reputación que gana el autor cuando su comentario pasa a destacado
const PuntosDestacado = 10

// Negocio es el cliente que paga la suscripción de su comunidad
type Negocio struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Nombre     string    `gorm:"not null" json:"nombre"`
	RUC        string    `gorm:"not null;size:13" json:"ruc"`
	EstadoPago string    `gorm:"not null;default:pago_por_verificar" json:"estado_pago"`
	Creado     time.Time `json:"creado"`
}

// Usuario es la entidad del lado del "uno" (un usuario escribe muchos comentarios)
type Usuario struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	NegocioID   uint         `gorm:"not null;index" json:"negocio_id"`
	Nombre      string       `gorm:"not null" json:"nombre"`
	Correo      string       `gorm:"not null" json:"correo"`
	Rol         string       `gorm:"not null;default:miembro" json:"rol"`
	Reputacion  int          `gorm:"not null;default:0" json:"reputacion"`
	Comentarios []Comentario `json:"comentarios,omitempty"`
}

// Comentario es la entidad principal: una ayuda que un usuario publica y que maneja estados
type Comentario struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UsuarioID uint      `gorm:"not null;index" json:"usuario_id"`
	Texto     string    `gorm:"not null" json:"texto"`
	Medalla   string    `json:"medalla"`
	Estado    string    `gorm:"not null;default:pendiente" json:"estado"`
	Creado    time.Time `json:"creado"`
}

// Estados válidos para el control de la comunidad
var estadosValidos = map[string]bool{
	EstadoPendiente: true,
	EstadoAprobado:  true,
	EstadoDestacado: true,
}
