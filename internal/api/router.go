// Package api arma las rutas: aquí se ve qué ruta pide token y qué rol.
package api

import (
	"net/http"

	"COMUNITARIO/internal/auth"
	"COMUNITARIO/internal/comentarios/handler"
	"COMUNITARIO/internal/modelos"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Router(h *handler.Manejador, a *auth.Manejador, secreto []byte) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/login", a.Login) // pública

	r.Group(func(r chi.Router) {
		r.Use(auth.Autenticar(secreto)) // desde aquí: pide token (401)

		r.Get("/comentarios", h.Listar)
		r.Post("/comentarios", h.Crear)
		r.Get("/comentarios/{id}", h.Obtener)
		r.Get("/usuarios", h.ListarUsuarios)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequiereRol(modelos.RolModerador)) // desde aquí: solo moderador (403)
			r.Patch("/comentarios/{id}/estado", h.CambiarEstado)
			r.Delete("/comentarios/{id}", h.Borrar)
		})
	})
	return r
}
