// Package handler atiende las peticiones HTTP de comentarios.
// NO consulta la base de datos: todo pasa por el servicio.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"COMUNITARIO/internal/auth"
	"COMUNITARIO/internal/comentarios/servicio"
	"COMUNITARIO/internal/modelos"

	"github.com/go-chi/chi/v5"
)

type Manejador struct {
	Servicio *servicio.Servicio
}

func responderJSON(w http.ResponseWriter, codigo int, valor any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(valor)
}

// responderError traduce el error del negocio a un código HTTP.
func responderError(w http.ResponseWriter, err error) {
	codigo := http.StatusInternalServerError
	switch {
	case errors.Is(err, modelos.ErrNoEncontrado):
		codigo = http.StatusNotFound
	case errors.Is(err, servicio.ErrParametros), errors.Is(err, servicio.ErrEstadoFiltro):
		codigo = http.StatusBadRequest
	case errors.Is(err, servicio.ErrSuscripcion):
		codigo = http.StatusForbidden
	case errors.Is(err, servicio.ErrTextoVacio), errors.Is(err, servicio.ErrTextoLargo),
		errors.Is(err, servicio.ErrUsuarioInvalido), errors.Is(err, servicio.ErrCategoria),
		errors.Is(err, servicio.ErrEstadoInvalido), errors.Is(err, servicio.ErrTransicion):
		codigo = http.StatusUnprocessableEntity
	}
	if codigo == http.StatusInternalServerError {
		http.Error(w, "Error interno", codigo)
		return
	}
	http.Error(w, err.Error(), codigo)
}

func leerID(r *http.Request) uint {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	return uint(id)
}

// POST /comentarios
func (m *Manejador) Crear(w http.ResponseWriter, r *http.Request) {
	var in ComentarioPeticion
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	c, err := m.Servicio.Crear(auth.ClaimsDe(r).UsuarioID, in.Texto, in.CategoriaID)
	if err != nil {
		responderError(w, err)
		return
	}
	responderJSON(w, http.StatusCreated, nuevaRespuesta(*c))
}

// GET /comentarios?estado=&page=&limit=
func (m *Manejador) Listar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err1 := strconv.Atoi(porDefecto(q.Get("page"), "1"))
	limit, err2 := strconv.Atoi(porDefecto(q.Get("limit"), "20"))
	if err1 != nil || err2 != nil {
		responderError(w, servicio.ErrParametros)
		return
	}
	lista, err := m.Servicio.Listar(q.Get("estado"), page, limit)
	if err != nil {
		responderError(w, err)
		return
	}
	responderJSON(w, http.StatusOK, nuevasRespuestas(lista))
}

func porDefecto(valor, otro string) string {
	if valor == "" {
		return otro
	}
	return valor
}

// GET /comentarios/{id}
func (m *Manejador) Obtener(w http.ResponseWriter, r *http.Request) {
	c, err := m.Servicio.Obtener(leerID(r))
	if err != nil {
		responderError(w, err)
		return
	}
	responderJSON(w, http.StatusOK, nuevaRespuesta(*c))
}

// PATCH /comentarios/{id}/estado  (solo moderador: lo exige el router)
func (m *Manejador) CambiarEstado(w http.ResponseWriter, r *http.Request) {
	var in CambioEstadoPeticion
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	c, err := m.Servicio.CambiarEstado(auth.ClaimsDe(r).UsuarioID, leerID(r), in.Estado)
	if err != nil {
		responderError(w, err)
		return
	}
	responderJSON(w, http.StatusOK, nuevaRespuesta(*c))
}

// DELETE /comentarios/{id}  (solo moderador)
func (m *Manejador) Borrar(w http.ResponseWriter, r *http.Request) {
	if err := m.Servicio.Borrar(leerID(r)); err != nil {
		responderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /usuarios
func (m *Manejador) ListarUsuarios(w http.ResponseWriter, r *http.Request) {
	u, err := m.Servicio.ListarUsuarios()
	if err != nil {
		responderError(w, err)
		return
	}
	responderJSON(w, http.StatusOK, nuevasRespuestasUsuario(u))
}
