-- Esquema que lee sqlc (solo las tablas que consultamos con SQL).
-- Las tablas reales las crea GORM; aquí se declaran con NOT NULL en lo que
-- siempre se llena, para que sqlc genere tipos simples (int64, string, time.Time).
CREATE TABLE usuarios (
    id         bigserial PRIMARY KEY,
    negocio_id bigint NOT NULL,
    nombre     text NOT NULL,
    correo     text NOT NULL,
    rol        text NOT NULL,
    reputacion bigint NOT NULL,
    creado     timestamptz NOT NULL
);

CREATE TABLE comentarios (
    id           bigserial PRIMARY KEY,
    usuario_id   bigint NOT NULL,
    categoria_id bigint,
    texto        text NOT NULL,
    estado       text NOT NULL,
    moderador_id bigint,
    creado       timestamptz NOT NULL,
    actualizado  timestamptz NOT NULL,
    eliminado    timestamptz
);

CREATE TABLE puntos (
    id            bigserial PRIMARY KEY,
    usuario_id    bigint NOT NULL,
    comentario_id bigint,
    cantidad      bigint NOT NULL,
    motivo        text NOT NULL,
    creado        timestamptz NOT NULL
);

CREATE TABLE medallas (
    id             bigserial PRIMARY KEY,
    nombre         text NOT NULL,
    puntos_minimos bigint NOT NULL,
    descripcion    text NOT NULL DEFAULT ''
);

CREATE TABLE usuario_medallas (
    id            bigserial PRIMARY KEY,
    usuario_id    bigint NOT NULL,
    medalla_id    bigint NOT NULL,
    comentario_id bigint,
    otorgada      timestamptz NOT NULL,
    UNIQUE (usuario_id, medalla_id)
);
