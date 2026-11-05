## D1 · ¿Quién pone el estado al crear?

*Opciones:* 
Que el cliente envíe el estado al crear el comentario, o que el servidor lo asigne automáticamente por seguridad.

*Qué elegimos:* 
Que el servidor imponga el estado inicial de forma automática.

*Por qué, en nuestro negocio:* 
En nuestra plataforma de soporte comunitario, un usuario no debe tener la capacidad de autoasignarse un estado de moderación avanzado (como "aprobado" o "destacado") al momento de publicar. Por control de calidad y seguridad, el servidor fuerza obligatoriamente que todo nuevo comentario nazca como "pendiente".

*Qué pasaría con la otra opción:* 
Si dejáramos que el cliente envíe el estado desde el formulario, cualquier usuario malintencionado podría crear registros marcados directamente como "destacados" o "aprobados", evadiendo por completo el filtro de la comunidad y alterando el sistema de reputación y medallas.


## Semana 8 · estructura

**Estructura que elegimos:** una carpeta por responsabilidad, y dentro de la entidad principal (Comentario) una subcarpeta por capa.

```
internal/
├── modelos/                    los 8 structs de las tablas (sin reglas, sin HTTP)
├── auth/                       token firmado, login, claves, middleware de token y de rol
├── api/router.go               aquí se ve qué ruta pide token y qué rol la puede usar
├── semilla/                    datos iniciales (negocio, usuarios, cuentas, medallas...)
├── pruebas/                    base de datos falsa en memoria para probar sin PostgreSQL
└── comentarios/                entidad principal
    ├── handler/                recibe la petición HTTP, lee el JSON y devuelve el código. NO consulta la base.
    ├── servicio/               reglas del negocio (máquina de estados, puntos, medallas). NO sabe nada de HTTP.
    └── repositorio/            única capa que habla con PostgreSQL (GORM)
```

**Por qué esta estructura:**

- `handler` no tiene `gorm` ni `DB`: solo llama al servicio. Así la función que atiende la petición no consulta la base directamente.
- `servicio/reglas.go` y `servicio/servicio.go` no importan `net/http`: la regla «de destacado no se vuelve atrás» se puede probar sin servidor.
- El servicio pide la base de datos mediante una interfaz (`Repositorio`). Por eso las pruebas usan una base falsa en memoria y corren sin PostgreSQL.
- Si mañana cambiamos de base de datos, solo se toca `comentarios/repositorio/` (y `main.go` para conectar). Los handlers, el servicio, las reglas y el router no cambian.
- La decisión de «esta ruta pide token» está en una sola línea de `internal/api/router.go` (`r.Use(auth.Autenticar(secreto))`), y la de «solo moderador» en la siguiente (`r.Use(auth.RequiereRol(...))`).

**Transición prohibida:** responde **422** y no 409 ni 400 ni 403. El JSON está bien formado (no es 400), el usuario sí tiene permiso (no es 403), pero la petición pide algo que la regla del negocio no permite. Se valida en `servicio.CambiarEstado` con `TransicionPermitida`.

   Revisado por Alexander Lazaro: probamos las seis peticiones en Bruno, el validador y go test.


   //


## Semana 10 · transacciones

**D1. Repositorio como interfaz.** El servicio solo conoce `servicio.Repositorio`. Hay dos adaptadores: `Gorm` y `Sqlc` (que incrusta a `Gorm` y reemplaza solo lo que se escribe en SQL). Se cambia de uno a otro en una línea de `main.go`.

**D2. DTO en la entrada y la salida.** `modelos.Comentario` ya no tiene etiquetas `json`. El cliente manda `ComentarioPeticion` (solo texto y categoría) y recibe `ComentarioRespuesta`. Así no puede fijar `estado`, `usuario_id` ni `moderador_id`, y `eliminado` nunca sale.

**D3. sqlc para los casos críticos.** Consultas escritas a mano en `db/consultas.sql` (buscar y guardar comentario, registrar punto, sumar reputación, listar medallas, otorgar medalla). Se genera con `sqlc generate` el paquete `filas`. Se guarda en el repositorio porque el profesor y el compañero deben poder compilar sin instalar sqlc.

**D4. Qué es atómico.** Al destacar un comentario: guardar el estado + registrar el punto + sumar reputación + otorgar medalla. Si algo falla, no queda nada a medias.

**D5. Transacción en cada adaptador.** GORM: `db.Transaction`. sqlc: `BeginTx` + `q.WithTx(tx)` + `Commit`/`Rollback`. Dentro de la transacción solo se usan métodos que el adaptador sqlc reemplaza (los demás irían por GORM, fuera de la transacción).

**D6. Prueba del fallo forzado.** `transaccion_test.go` guarda el comentario, suma reputación y devuelve un error a mitad: comprueba que el estado y la reputación vuelven a como estaban, con los dos adaptadores. Se omite sola si no hay PostgreSQL.
