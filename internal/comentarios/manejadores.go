package comentarios

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Manejador struct {
	DB *gorm.DB
}

const limiteMaximo = 20 // datos móviles limitados: nunca más de 20 por página

func leerID(r *http.Request) uint {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	return uint(id)
}

func responderJSON(w http.ResponseWriter, codigo int, valor any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(valor)
}

// POST /comentarios - Crear un comentario de ayuda (rol: miembro o moderador)
func (m *Manejador) Crear(w http.ResponseWriter, r *http.Request) {
	var c Comentario
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if len(c.Texto) == 0 {
		http.Error(w, "El texto es obligatorio", http.StatusUnprocessableEntity)
		return
	}

	// Validación: el usuario debe existir en PostgreSQL antes de guardar
	var usr Usuario
	if err := m.DB.First(&usr, c.UsuarioID).Error; err != nil {
		http.Error(w, "El usuario no existe", http.StatusUnprocessableEntity)
		return
	}

	// Regla de negocio: la comunidad solo acepta comentarios si el negocio pagó (estado activa)
	var neg Negocio
	if err := m.DB.First(&neg, usr.NegocioID).Error; err != nil || neg.EstadoPago != PagoActiva {
		http.Error(w, "La suscripción del negocio no está activa", http.StatusForbidden)
		return
	}

	// Regla de negocio: todo comentario nace pendiente y sin medalla
	c.ID = 0
	c.Estado = EstadoPendiente
	c.Medalla = ""
	c.Creado = time.Now()

	if err := m.DB.Create(&c).Error; err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	responderJSON(w, http.StatusCreated, c)
}

// GET /comentarios?estado=&page=&limit= - Listar con filtro seguro y paginado
func (m *Manejador) Listar(w http.ResponseWriter, r *http.Request) {
	estado := r.URL.Query().Get("estado")
	if estado != "" && !estadosValidos[estado] {
		http.Error(w, "Estado no permitido", http.StatusBadRequest)
		return
	}

	page, err1 := strconv.Atoi(valorONada(r.URL.Query().Get("page"), "1"))
	limit, err2 := strconv.Atoi(valorONada(r.URL.Query().Get("limit"), "20"))
	if err1 != nil || err2 != nil || page < 1 || limit < 1 || limit > limiteMaximo {
		http.Error(w, "page y limit deben ser números válidos (limit máximo 20)", http.StatusBadRequest)
		return
	}

	consulta := m.DB.Order("id desc").Limit(limit).Offset((page - 1) * limit)
	if estado != "" {
		// Parámetro seguro '?' (evita inyección SQL)
		consulta = consulta.Where("estado = ?", estado)
	}

	var lista []Comentario
	consulta.Find(&lista)
	responderJSON(w, http.StatusOK, lista)
}

func valorONada(valor, porDefecto string) string {
	if valor == "" {
		return porDefecto
	}
	return valor
}

// GET /comentarios/{id} - Detalle de un comentario
func (m *Manejador) Obtener(w http.ResponseWriter, r *http.Request) {
	var c Comentario
	if err := m.DB.First(&c, leerID(r)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Comentario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	responderJSON(w, http.StatusOK, c)
}

// GET /usuarios - Solución al Problema N+1 usando Preload (perfil con reputación)
func (m *Manejador) ListarUsuarios(w http.ResponseWriter, r *http.Request) {
	var usuarios []Usuario
	// Ejecuta estrictamente 2 consultas en la base de datos
	m.DB.Preload("Comentarios").Find(&usuarios)
	responderJSON(w, http.StatusOK, usuarios)
}

// actor identifica quién hace la petición con la cabecera X-Usuario-ID.
// (En el Hito 3 se reemplaza por autenticación real.)
func (m *Manejador) actor(r *http.Request) (*Usuario, int) {
	idStr := r.Header.Get("X-Usuario-ID")
	if idStr == "" {
		return nil, http.StatusUnauthorized
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, http.StatusUnauthorized
	}
	var u Usuario
	if err := m.DB.First(&u, id).Error; err != nil {
		return nil, http.StatusUnauthorized
	}
	return &u, 0
}

// PATCH /comentarios/{id}/estado - Cambiar el estado (solo moderador, solo transiciones permitidas)
func (m *Manejador) CambiarEstado(w http.ResponseWriter, r *http.Request) {
	var datos struct {
		Estado string `json:"estado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if !estadosValidos[datos.Estado] {
		http.Error(w, "Estado no permitido", http.StatusUnprocessableEntity)
		return
	}
	if r.Header.Get("X-Usuario-ID") == "" {
		http.Error(w, "Falta la cabecera X-Usuario-ID", http.StatusUnauthorized)
		return
	}

	quien, codigo := m.actor(r)
	if quien == nil {
		http.Error(w, "Usuario no identificado", codigo)
		return
	}
	if quien.Rol != RolModerador {
		http.Error(w, "Solo un moderador puede cambiar el estado", http.StatusForbidden)
		return
	}

	var c Comentario
	if err := m.DB.First(&c, leerID(r)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Comentario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	if !TransicionPermitida(c.Estado, datos.Estado) {
		http.Error(w, "Transición de estado no permitida", http.StatusConflict)
		return
	}

	err := m.DB.Transaction(func(tx *gorm.DB) error {
		c.Estado = datos.Estado
		if datos.Estado == EstadoDestacado {
			// Regla de negocio: la reputación y la medalla se dan solo al destacar
			var autor Usuario
			if err := tx.First(&autor, c.UsuarioID).Error; err != nil {
				return err
			}
			autor.Reputacion += PuntosDestacado
			c.Medalla = MedallaPara(autor.Reputacion)
			if err := tx.Save(&autor).Error; err != nil {
				return err
			}
		}
		return tx.Save(&c).Error
	})
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	responderJSON(w, http.StatusOK, c)
}

// DELETE /comentarios/{id} - Borrar comentario (solo moderador)
func (m *Manejador) Borrar(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Usuario-ID") == "" {
		http.Error(w, "Falta la cabecera X-Usuario-ID", http.StatusUnauthorized)
		return
	}
	quien, codigo := m.actor(r)
	if quien == nil {
		http.Error(w, "Usuario no identificado", codigo)
		return
	}
	if quien.Rol != RolModerador {
		http.Error(w, "Solo un moderador puede borrar", http.StatusForbidden)
		return
	}

	res := m.DB.Delete(&Comentario{}, leerID(r))
	if res.Error == nil && res.RowsAffected == 0 {
		http.Error(w, "Comentario no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
