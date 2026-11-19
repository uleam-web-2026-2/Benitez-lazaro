// Package api arma las rutas: aquí se ve qué ruta pide token y qué rol.
package api

import (
	"net/http"

	"COMUNITARIO/internal/auth"
	"COMUNITARIO/internal/comentarios/handler"
	"COMUNITARIO/internal/modelos"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Router(h *handler.Manejador, a *auth.Manejador, secreto []byte) http.Handler {
	r := chi.NewRouter()
	// CORS arriba de todo: responde las peticiones previas (OPTIONS), que no tienen
	// manejador. Minimo privilegio: se nombra el front, no "*".
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300,
	}))
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
