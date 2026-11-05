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

### D1 · El caso
El caso es destacar un comentario (`CambiarEstado` hacia `destacado`). Escribe en cuatro lugares: cambia el estado del comentario, registra una fila en `puntos`, suma la reputación del usuario y otorga la medalla que corresponda. Si fallara a mitad de camino, quedaría un comentario destacado sin puntos, o puntos sin reputación, o reputación sin medalla: el usuario vería una reputación que no coincide con su historial.

### D2 · Opciones que consideramos
1. **`EnTransaccion(fn)` en la interfaz del repositorio.** El servicio llama `Repo.EnTransaccion(func(tx Repositorio) error {...})` y adentro usa `tx`. Cada adaptador abre su propia transacción (GORM con `db.Transaction`, sqlc con `BeginTx` y `WithTx`). Cumple las cinco restricciones: el servicio no importa gorm ni database/sql, el repositorio falso solo ejecuta `fn`, los manejadores no cambian y cada adaptador resuelve la suya.
2. **Pasar un `*gorm.DB` o un `*sql.Tx` al servicio.** Es lo más corto, pero el servicio tendría que importar gorm o database/sql (rompe la restricción 1) y no funcionaría con el otro adaptador.
3. **Un método compuesto `DestacarComentario(...)` en el repositorio** que haga las cuatro escrituras. Cumple la atomicidad, pero las reglas (cuántos puntos, qué medalla toca) quedarían escritas dentro de cada adaptador, dos veces, y no se podrían probar con el repositorio falso.

### D3 · La que elegimos y por qué
La opción 1. En nuestro proyecto las reglas de puntos y medallas (`PuntosDestacado`, `MedallaMayor`) viven en el servicio y se prueban sin base de datos. Con `EnTransaccion` esas reglas se quedan donde están y la transacción es solo un detalle del adaptador. Además, desde el hito 2 la interfaz ya tenía `EnTransaccion` para GORM, así que el servicio no cambió.

### D4 · La que descartamos y por qué
La opción 2 (pasar `tx` al servicio). El servicio tendría que importar `gorm` y, al cambiar a sqlc, habría que reescribirlo para `database/sql`. Perderíamos las pruebas con repositorio falso, que son las que hacen que `go test ./...` corra sin PostgreSQL.

### D5 · Con sqlc
Hubo que implementar `EnTransaccion` en el adaptador sqlc: `Conexion.BeginTx`, crear una copia del adaptador con `q.WithTx(tx)` y hacer `Commit` o `Rollback` según el resultado de `fn`. También hubo que reemplazar en sqlc los cinco métodos que el servicio usa dentro de la transacción (guardar comentario, registrar punto, sumar reputación, listar medallas y otorgar medalla); si no, los métodos heredados de GORM habrían corrido fuera de la transacción. Límite que anotamos: dentro de `fn` solo deben usarse métodos reescritos por el adaptador. El diseño (interfaz y servicio) no cambió; se agregaron cuatro consultas más en `db/consultas.sql`.

### D6 · Archivos
Tocamos: `internal/comentarios/repositorio/sqlc.go` (adaptador y `EnTransaccion`), `db/consultas.sql`, `db/schema.sql`, `sqlc.yaml`, `internal/comentarios/repositorio/filas/` (generado por sqlc), `internal/comentarios/repositorio/transaccion_test.go` (prueba con fallo forzado contra PostgreSQL con los dos adaptadores) y `main.go` (se elige el adaptador). Quedaron igual: `servicio.go`, `reglas.go`, el repositorio falso, los manejadores y el router.

**Qué operación de más necesitaría esto si el proyecto creciera:** el pago mensual del negocio (`estado_pago` de `Negocio`) junto con un registro de pago; hoy se verifica a mano, pero al automatizarlo habría que actualizar dos tablas a la vez.
