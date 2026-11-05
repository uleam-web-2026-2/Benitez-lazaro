# Problemas en Windows (Semana 10)

El taller se probó solo en macOS. Estos fueron los tropiezos en Windows (PowerShell) y qué hicimos:

1. **`Expand-Archive` / Explorador dejó la carpeta duplicada (`taller10\taller10`).** El zip ya trae una carpeta con su nombre. Se arregló usando la ruta con la carpeta repetida al copiar (`Copy-Item`).
2. **`go run .` falló con «Only one usage of each socket address».** Quedaba un servidor anterior usando el puerto 8080. Se buscó el proceso con `Get-NetTCPConnection -LocalPort 8080` y se cerró con `Stop-Process -Id <id> -Force`.
3. **`npx` bloqueado por la política de ejecución de PowerShell.** Se usa `npx.cmd` en lugar de `npx`.
4. **Redocly escribe en el canal de errores,** así que `> archivo` dejaba `redocly.txt` vacío. Se redirigió con `2>&1`.
5. **Avisos «LF will be replaced by CRLF» de git.** Son normales en Windows; no cambian el código.
