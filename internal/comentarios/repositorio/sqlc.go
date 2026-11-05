package repositorio

import (
	"context"
	"database/sql"
	"errors"

	"COMUNITARIO/internal/comentarios/repositorio/filas"
	"COMUNITARIO/internal/comentarios/servicio"
	"COMUNITARIO/internal/modelos"
)

// Sqlc es el SEGUNDO adaptador del repositorio. Incrusta el de GORM y solo
// reemplaza los métodos que ahora van con SQL escrito a mano (sqlc).
// Todo lo demás (listar, crear, borrar...) lo sigue haciendo GORM.
type Sqlc struct {
	*Gorm
	Conexion *sql.DB
	q        *filas.Queries
	enTx     bool // true cuando este adaptador vive dentro de una transacción
}

var _ servicio.Repositorio = (*Sqlc)(nil)

func NuevoSqlc(g *Gorm, c *sql.DB) *Sqlc {
	return &Sqlc{Gorm: g, Conexion: c, q: filas.New(c)}
}

func aNullInt(p *uint) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

func aPuntero(n sql.NullInt64) *uint {
	if !n.Valid {
		return nil
	}
	v := uint(n.Int64)
	return &v
}

func (a *Sqlc) ObtenerComentario(id uint) (*modelos.Comentario, error) {
	f, err := a.q.BuscarComentario(context.Background(), int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, modelos.ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &modelos.Comentario{
		ID: uint(f.ID), UsuarioID: uint(f.UsuarioID), CategoriaID: aPuntero(f.CategoriaID),
		Texto: f.Texto, Estado: f.Estado, ModeradorID: aPuntero(f.ModeradorID),
		Creado: f.Creado, Actualizado: f.Actualizado,
	}, nil
}

func (a *Sqlc) GuardarComentario(c *modelos.Comentario) error {
	n, err := a.q.GuardarComentario(context.Background(), filas.GuardarComentarioParams{
		ID: int64(c.ID), CategoriaID: aNullInt(c.CategoriaID), Texto: c.Texto,
		Estado: c.Estado, ModeradorID: aNullInt(c.ModeradorID), Actualizado: c.Actualizado,
	})
	if err == nil && n == 0 {
		return modelos.ErrNoEncontrado
	}
	return err
}

func (a *Sqlc) RegistrarPunto(p *modelos.Punto) error {
	id, err := a.q.RegistrarPunto(context.Background(), filas.RegistrarPuntoParams{
		UsuarioID: int64(p.UsuarioID), ComentarioID: aNullInt(p.ComentarioID),
		Cantidad: int64(p.Cantidad), Motivo: p.Motivo, Creado: p.Creado,
	})
	if err == nil {
		p.ID = uint(id)
	}
	return err
}

func (a *Sqlc) SumarReputacion(usuarioID uint, puntos int) (int, error) {
	r, err := a.q.SumarReputacion(context.Background(), filas.SumarReputacionParams{
		ID: int64(usuarioID), Puntos: int64(puntos),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, modelos.ErrNoEncontrado
	}
	return int(r), err
}

func (a *Sqlc) CatalogoMedallas() ([]modelos.Medalla, error) {
	filasDB, err := a.q.ListarMedallas(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]modelos.Medalla, 0, len(filasDB))
	for _, f := range filasDB {
		out = append(out, modelos.Medalla{ID: uint(f.ID), Nombre: f.Nombre,
			PuntosMinimos: int(f.PuntosMinimos), Descripcion: f.Descripcion})
	}
	return out, nil
}

func (a *Sqlc) OtorgarMedalla(um *modelos.UsuarioMedalla) error {
	return a.q.OtorgarMedalla(context.Background(), filas.OtorgarMedallaParams{
		UsuarioID: int64(um.UsuarioID), MedallaID: int64(um.MedallaID),
		ComentarioID: aNullInt(um.ComentarioID), Otorgada: um.Otorgada,
	})
}

// EnTransaccion (Fase 6): con sqlc la transacción es de database/sql.
// Se abre con Begin y la consulta se "ata" a ella con q.WithTx(tx).
// Si fn devuelve error → Rollback; si no → Commit.
func (a *Sqlc) EnTransaccion(fn func(servicio.Repositorio) error) error {
	if a.enTx {
		return fn(a)
	}
	tx, err := a.Conexion.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	dentro := &Sqlc{Gorm: a.Gorm, Conexion: a.Conexion, q: a.q.WithTx(tx), enTx: true}
	if err := fn(dentro); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
