package comentarios

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestTransicionPendienteAAprobado(t *testing.T) {
	if !TransicionPermitida(EstadoPendiente, EstadoAprobado) {
		t.Error("pendiente -> aprobado debe estar permitido")
	}
}

func TestTransicionAprobadoADestacado(t *testing.T) {
	if !TransicionPermitida(EstadoAprobado, EstadoDestacado) {
		t.Error("aprobado -> destacado debe estar permitido")
	}
}

func TestTransicionProhibidaDestacadoNoRetrocede(t *testing.T) {
	for _, a := range []string{EstadoPendiente, EstadoAprobado} {
		if TransicionPermitida(EstadoDestacado, a) {
			t.Errorf("destacado -> %s debe estar prohibido (se podría sumar reputación varias veces)", a)
		}
	}
}

func TestTransicionProhibidaSaltoDePendienteADestacado(t *testing.T) {
	if TransicionPermitida(EstadoPendiente, EstadoDestacado) {
		t.Error("pendiente -> destacado debe estar prohibido: primero hay que aprobar")
	}
}

func TestMedallaPorReputacion(t *testing.T) {
	casos := map[int]string{0: "Novato", 24: "Novato", 25: "Ayudante", 49: "Ayudante", 50: "Expert"}
	for rep, esperada := range casos {
		if got := MedallaPara(rep); got != esperada {
			t.Errorf("reputación %d: medalla %q, se esperaba %q", rep, got, esperada)
		}
	}
}

func TestCrearJSONInvalidoDevuelve400(t *testing.T) {
	m := &Manejador{} // sin base de datos: la validación ocurre antes de consultarla
	req := httptest.NewRequest(http.MethodPost, "/comentarios", strings.NewReader("{no es json"))
	rec := httptest.NewRecorder()
	m.Crear(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("código %d, se esperaba 400", rec.Code)
	}
}

func TestCrearTextoVacioDevuelve422(t *testing.T) {
	m := &Manejador{}
	req := httptest.NewRequest(http.MethodPost, "/comentarios", strings.NewReader(`{"usuario_id":1,"texto":""}`))
	rec := httptest.NewRecorder()
	m.Crear(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("código %d, se esperaba 422", rec.Code)
	}
}

func peticionEstado(cuerpo, usuarioID string) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, "/comentarios/1/estado", strings.NewReader(cuerpo))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	if usuarioID != "" {
		req.Header.Set("X-Usuario-ID", usuarioID)
	}
	return req
}

func TestCambiarEstadoDesconocidoDevuelve422(t *testing.T) {
	m := &Manejador{}
	rec := httptest.NewRecorder()
	m.CambiarEstado(rec, peticionEstado(`{"estado":"inventado"}`, "2"))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("código %d, se esperaba 422", rec.Code)
	}
}

func TestCambiarEstadoSinUsuarioDevuelve401(t *testing.T) {
	m := &Manejador{}
	rec := httptest.NewRecorder()
	m.CambiarEstado(rec, peticionEstado(`{"estado":"aprobado"}`, ""))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("código %d, se esperaba 401", rec.Code)
	}
}
