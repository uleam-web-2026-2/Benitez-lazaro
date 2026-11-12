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


## Semana 11 · el contenedor

### D1 · Qué entra en la imagen
La imagen final (`comunitario:latest`) solo tiene el binario `/servidor` sobre distroless: pesa 38.7 MB (11.6 MB de contenido). Todo lo demás se quedó en la etapa de compilación (`comunitario:compilacion`), que pesa 1.63 GB porque trae Go, las herramientas y las librerías. Sin `.dockerignore`, el contexto de la primera construcción fue de 2.45 MB y con él bajó a 6.78 kB: la diferencia era el `.env` y la carpeta `.git`. Sin `.dockerignore`, el `.env` no llegaría a la imagen final, porque la segunda etapa solo copia `/servidor`; pero sí llegaría a la etapa de compilación (`COPY . .` lo dejaría en `/src/.env`) y quedaría guardado en las capas de `comunitario:compilacion` y en la caché de Docker. Si alguien publica o comparte esa imagen, se lleva la clave.

### D2 · Dónde está la base, vista desde el contenedor
Con `DB_HOST=localhost` falló con `127.0.0.1:5432 ... connection refused` (`2_local.txt`): dentro del contenedor `localhost` es el propio contenedor, y ahí no hay ninguna base. Pusimos `DB_HOST=host.docker.internal` y `DB_PORT=5433`, que es la máquina donde publicamos el contenedor `pg` de PostgreSQL. Usamos un contenedor porque el PostgreSQL de Windows rechaza las conexiones que vienen de Docker. Cuando la API corra en un servidor que no es nuestra máquina, `host.docker.internal` ya no sirve: habrá que poner la dirección real de la base (o el nombre de su servicio si están en la misma red).

### D3 · Los puertos
Usamos `-p 8080:8080`. El número de la izquierda es el de la máquina (el que usa Bruno, `localhost:8080`); el de la derecha es el del contenedor, donde escucha el servidor. El de la derecha tiene que coincidir con `PORT` de `.env.docker`, porque el servidor escucha en ese puerto. Si no coincide, el puerto publicado no tiene nada escuchando y la petición se corta (`Connection reset by peer`) o no recibe respuesta.

### D4 · La hora
Con `go run` nuestras fechas salen como `2026-11-11T19:45:40.796347-05:00` y en el contenedor como `2026-11-12T01:00:46.615151Z`. Es el mismo momento (UTC-5 frente a UTC), solo escrito distinto. Cambió porque la imagen distroless no trae la base de zonas horarias: dentro del contenedor la hora local es UTC. El código y la base son los mismos. Decidimos no cambiar el código y dejar que la API devuelva UTC (`Z`), que es una forma estándar y sin ambigüedad; el cliente la convierte a su zona. También los logs del contenedor salen en UTC.

### Nota sobre el código
Para que el log del contenedor mostrara una línea por petición (paso 2.4) agregamos `r.Use(middleware.Logger)` en `internal/api/router.go`. No cambia ninguna respuesta de la API.


## Semana 11 · reto: la API y la base, juntas

### D5 · Cómo encuentra la API a la base
La API usa `DB_HOST=db` y `DB_PORT=5432`. `db` es el nombre del servicio en `compose.yaml`: Compose crea una red interna y resuelve ese nombre a la dirección del contenedor de PostgreSQL. Usamos el 5432 y no el 5433 porque el 5433 era el puerto que habíamos publicado en nuestra máquina para el contenedor `pg` suelto; entre contenedores de la misma red se usa el puerto real de PostgreSQL, que es el 5432. Con Compose no necesitamos `host.docker.internal`.

### D6 · El orden de arranque
Con la base recién creada hay dos riesgos: que la API arranque antes de que PostgreSQL acepte conexiones, y que la base esté vacía. No llegamos a ver el primer fallo en nuestras pruebas porque lo previmos desde el principio: `depends_on` solo espera a que el contenedor arranque, no a que PostgreSQL esté listo, así que agregamos un `healthcheck` con `pg_isready` y `condition: service_healthy`, y la API arranca solo cuando la base está `healthy`. El `healthcheck` usa `-h 127.0.0.1` (TCP) porque en el primer arranque PostgreSQL crea la base con un servidor temporal que solo escucha por socket, y sin `-h` la prueba daría «listo» demasiado pronto. En cuanto a la base vacía, no es problema: la API ejecuta `AutoMigrate` y la semilla al arrancar y crea las tablas y los datos. Lo comprobamos tres veces seguidas con `docker compose down -v` y `docker compose up -d --build` (los tres intentos terminaron con `db` healthy, `api` Up y `401` sin token).

### D7 · Dónde quedan los datos
La base guarda sus archivos en el volumen `datos_pg` (montado en `/var/lib/postgresql/data`). `docker compose down` elimina los contenedores y la red, pero no el volumen, y por eso los registros siguen al volver a subir: comprobamos con `psql` que el comentario 4, aprobado antes del `down`, seguía `aprobado` después. Lo que borra los datos es `docker compose down -v`, que además elimina el volumen.

### D8 · Los secretos
`compose.yaml` no tiene ninguna contraseña. Usa `${DB_PASSWORD}` y `${JWT_SECRET}`, que Compose toma del archivo `.env` de la carpeta (el mismo de `go run .`, que Git ignora). Con `${DB_PASSWORD:?...}` Compose se detiene con un mensaje si falta alguno. En el repositorio solo está `.env.example`, sin valores reales: quien clone el proyecto copia ese archivo a `.env` y completa los valores. `.env.docker` es solo para el `docker run` suelto y Compose no lo usa.

### D9 · Lo que se publica en la máquina
Solo se publica el puerto de la API, `8080:8080` (el valor de `PORT`), que es al que llama Bruno. La base no se publica: así no choca con el PostgreSQL de Windows (5432) ni con el contenedor `pg` (5433), y no queda expuesta fuera de la red de Compose. Cuando necesitemos consultarla usamos `docker compose exec db psql -U postgres -d APP-COMUNITARIA`, y así revisamos los datos tras el `down`.
