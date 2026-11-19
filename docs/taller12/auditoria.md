# Semana 12 · auditoría de inyección

Búsqueda: `git grep -n -E -f docs/taller12/patrones.txt -- "*.go" ":(exclude)*_test.go"`. Salida completa en `auditoria.txt` (31 líneas).

| Línea | ¿Arma SQL? | ¿Llega texto del cliente? | ¿Va aparte? | Qué hicimos |
| --- | --- | --- | --- | --- |
| `repositorio.go:31` `Where("estado = ?", estado)` | Sí | Sí, `?estado=` de la URL | Sí, con `?`. Además `servicio.Listar` rechaza con 400 cualquier estado fuera de la lista (`EstadoValido`) antes de llegar aquí | Nada: es segura. |
| `repositorio.go:29` `Order("id desc").Limit().Offset()` y `:34` `Find` | Sí | `page` y `limit` de la URL, pero ya convertidos a número con `strconv.Atoi` y validados | `Order` es texto fijo; `Limit` y `Offset` reciben enteros | Nada. |
| `repositorio.go:40`, `:61`, `:69` `First(&x, id)` | Sí | El `{id}` de la ruta (solo `:40`); `:61` y `:69` usan ids del token o de la base | `id` es `uint`: `leerID` lo convierte con `strconv.Atoi`; un texto raro da 0 y la respuesta es 404 | Nada. |
| `repositorio.go:95` `Select("reputacion").First(&u, usuarioID)` | Sí | No: `usuarioID` es un `uint` que sale del comentario guardado | `Select` es texto fijo, el id es `uint` | Nada. |
| `repositorio.go:84` y `:89` `Where("id = ?", …)` y `UpdateColumn(… gorm.Expr("reputacion + ?", puntos))` | Sí | No: los ids y los puntos los pone el servidor | Sí, con `?` | Nada. |
| `repositorio.go:78` `Preload(...).Find` y `:105` `Order("puntos_minimos").Find` | Sí | No: son textos fijos | — | Nada. |
| `repositorio.go:122` `Where("login = ?", login)` | Sí | **Sí: el `usuario` del JSON del login** (el único texto libre del cliente que llega a una consulta) | Sí, con `?` | Nada: es segura. |
| `filas/consultas.sql.go:31, 63, 82, 123, 148, 171` y `filas/db.go:13, 15, 16` | Sí | Sí (ids y valores) | Sí: es código generado por sqlc; las consultas están en `db/consultas.sql` con `$1`, `$2`… y los valores van como argumentos | Nada. El código generado no se edita. |
| `handler.go:72` `r.URL.Query()` | No: solo lee los parámetros de la URL | Sí | Después van al servicio, que valida `estado` y convierte `page` y `limit` a números | Nada. |
| `auth/clave.go:26`, `auth/token.go:42, 43, 53` | No: `Sprintf` y `+` arman el hash de la clave y el token | No | — | Nada. |
| `router.go:21, 29` `r.Group(...)` | No: agrupa rutas de chi | — | — | Nada. |
| `main.go:35, 74, 75` | No: arma la cadena de conexión con variables del entorno, un mensaje y la dirección del servidor | No | — | Nada. |

**Resultado:** de las 31 líneas, ninguna arma SQL pegando texto del cliente dentro de la consulta. No cambiamos código, así que no hace falta `auditoria_despues.txt`. La prueba del paso 1.4 (`inyeccion.txt`) lo confirma.
