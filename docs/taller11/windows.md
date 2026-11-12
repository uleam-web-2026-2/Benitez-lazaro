# Problemas en Windows (Semana 11)

El taller se probó solo en macOS. Estos fueron los tropiezos en Windows (PowerShell) y qué hicimos:

1. **`docker run --rm hello-world` falló con `open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified`.** Docker estaba instalado pero Docker Desktop no estaba corriendo. Se abrió Docker Desktop, se esperó a que el motor arrancara y se repitió el comando.
2. **El zip se descomprimió con la carpeta repetida (`taller11\taller11`) y a veces no estaba donde se esperaba.** Se listó con `Get-ChildItem -Recurse -Depth 2` y se usó la ruta real al copiar los archivos.
3. **`> archivo` en PowerShell no sirve para guardar la salida de Docker** (la guarda en otra codificación o con mensajes de error de PowerShell). Se usó `cmd /c "comando > archivo 2>&1"` para los comandos que escriben en el canal de errores (`docker run`, `docker compose logs`), y `| Out-File -Encoding utf8` para los demás. Además, `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8` evitó que `docker compose ps` guardara `sΓÇª` en lugar de `…`.
4. **El PostgreSQL de Windows (servicio `postgresql-x64-18`, puerto 5432) no sirve para el contenedor,** porque rechaza las conexiones que vienen de Docker. Se creó un contenedor de PostgreSQL llamado `pg` publicado en el 5433 y el contenedor de la API se conectó a `host.docker.internal:5433`.
5. **El puerto 8080 estuvo ocupado por un `go run` anterior** al arrancar otro servidor. Se buscó el proceso con `Get-NetTCPConnection -LocalPort 8080` y se cerró con `Stop-Process`.
6. **El log del contenedor no mostraba las peticiones**, porque nuestro router no tenía `middleware.Logger`. Se agregó `r.Use(middleware.Logger)` en `internal/api/router.go` (no cambia ninguna respuesta).
7. **Las respuestas de Bruno se quedan en pantalla** hasta que se vuelve a pulsar Send, lo que confundió la comprobación de que los datos sobreviven a `docker compose down`. Se comprobó directamente con `docker compose exec db psql ...`.
8. **Avisos «LF will be replaced by CRLF» de Git.** Son normales en Windows y no cambian el código.
