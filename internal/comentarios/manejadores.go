package comentarios

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Manejador struct {
	DB *gorm.DB
}

func leerID(r *http.Request) uint {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	return uint(id)
}

// POST /comentarios - Crear un comentario de ayuda
func (m *Manejador) Crear(w http.ResponseWriter, r *http.Request) {
	var c Comentario
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Regla de negocio: Por seguridad, todo comentario nace pendiente de moderación
	c.Estado = "pendiente"

	// Validación: Verificar que el usuario exista en PostgreSQL antes de guardar
	var usr Usuario
	if err := m.DB.First(&usr, c.UsuarioID).Error; err != nil {
		http.Error(w, "El usuario no existe", http.StatusUnprocessableEntity)
		return
	}

	m.DB.Create(&c)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GET /comentarios - Listar con filtro seguro (Evita Inyección SQL)
func (m *Manejador) Listar(w http.ResponseWriter, r *http.Request) {
	var lista []Comentario
	estado := r.URL.Query().Get("estado")

	consulta := m.DB
	if estado != "" {
		// Parámetro seguro '?' exigido en la fase 2d
		consulta = consulta.Where("estado = ?", estado)
	}

	consulta.Find(&lista)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

// GET /usuarios - Solución al Problema N+1 usando Preload
func (m *Manejador) ListarUsuarios(w http.ResponseWriter, r *http.Request) {
	var usuarios []Usuario
	// Ejecuta estrictamente 2 consultas en la base de datos
	m.DB.Debug().Preload("Comentarios").Find(&usuarios)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usuarios)
}

// PUT /comentarios/{id} - Actualizar estado con validación de map
func (m *Manejador) Actualizar(w http.ResponseWriter, r *http.Request) {
	var datos Comentario
	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validación de estados permitidos (Fase 2b)
	if !estadosValidos[datos.Estado] {
		http.Error(w, "Estado no permitido", http.StatusUnprocessableEntity)
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

	c.Estado = datos.Estado
	c.Medalla = datos.Medalla
	m.DB.Save(&c)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// DELETE /comentarios/{id} - Borrar comentario
func (m *Manejador) Borrar(w http.ResponseWriter, r *http.Request) {
	res := m.DB.Delete(&Comentario{}, leerID(r))
	if res.Error == nil && res.RowsAffected == 0 {
		http.Error(w, "Comentario no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
