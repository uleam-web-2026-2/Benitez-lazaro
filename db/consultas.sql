-- name: BuscarComentario :one
SELECT id, usuario_id, categoria_id, texto, estado, moderador_id, creado, actualizado
FROM comentarios WHERE id = sqlc.arg(id) AND eliminado IS NULL;

-- name: GuardarComentario :execrows
UPDATE comentarios SET categoria_id = sqlc.narg(categoria_id), texto = sqlc.arg(texto),
    estado = sqlc.arg(estado), moderador_id = sqlc.narg(moderador_id),
    actualizado = sqlc.arg(actualizado)
WHERE id = sqlc.arg(id) AND eliminado IS NULL;

-- name: RegistrarPunto :one
INSERT INTO puntos (usuario_id, comentario_id, cantidad, motivo, creado)
VALUES (sqlc.arg(usuario_id), sqlc.narg(comentario_id), sqlc.arg(cantidad),
        sqlc.arg(motivo), sqlc.arg(creado))
RETURNING id;

-- name: SumarReputacion :one
UPDATE usuarios SET reputacion = reputacion + sqlc.arg(puntos)
WHERE id = sqlc.arg(id) RETURNING reputacion;

-- name: ListarMedallas :many
SELECT id, nombre, puntos_minimos, descripcion FROM medallas ORDER BY puntos_minimos;

-- name: OtorgarMedalla :exec
INSERT INTO usuario_medallas (usuario_id, medalla_id, comentario_id, otorgada)
VALUES (sqlc.arg(usuario_id), sqlc.arg(medalla_id), sqlc.narg(comentario_id), sqlc.arg(otorgada))
ON CONFLICT (usuario_id, medalla_id) DO NOTHING;
