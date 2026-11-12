# Semana 11 · las seis peticiones del hito 2 (AyudaComún)

| Petición | Con go run | En el contenedor | Con Compose |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | 401 |
| 2 login | 200 | 200 | 200 |
| 3 con token | 200 | 200 | 200 |
| 4 rol sin permiso | 403 | 403 | 403 |
| 5 cambio permitido | 200 | 200 | 200 |
| 6 transición prohibida | 422 | 422 | 422 |

Petición 3 con go run (respuesta completa):

```json
[
  { "id": 8, "usuario_id": 1, "categoria_id": 1, "texto": "Comentario nuevo para la demostracion", "estado": "pendiente", "moderador_id": null, "creado": "2026-11-11T19:45:40.796347-05:00", "actualizado": "2026-11-11T19:45:40.796347-05:00" },
  { "id": 7, "usuario_id": 1, "categoria_id": null, "texto": "prueba paso 2.1", "estado": "pendiente", "moderador_id": null, "creado": "2026-11-04T19:59:39.737141-05:00", "actualizado": "2026-11-04T19:59:39.737141-05:00" },
  { "id": 4, "usuario_id": 1, "categoria_id": 1, "texto": "Comentario nuevo para la demostracion", "estado": "aprobado", "moderador_id": 2, "creado": "2026-11-04T19:41:24.111079-05:00", "actualizado": "2026-11-04T19:42:04.957653-05:00" },
  { "id": 3, "usuario_id": 2, "categoria_id": 3, "texto": "Puedes filtrar los comentarios por estado.", "estado": "pendiente", "moderador_id": null, "creado": "2026-11-04T19:39:09.740489-05:00", "actualizado": "2026-11-04T19:39:09.740489-05:00" },
  { "id": 2, "usuario_id": 1, "categoria_id": 2, "texto": "La transferencia se verifica en 24 horas hábiles.", "estado": "aprobado", "moderador_id": 2, "creado": "2026-11-04T19:39:09.740489-05:00", "actualizado": "2026-11-04T19:39:09.740489-05:00" },
  { "id": 1, "usuario_id": 1, "categoria_id": 1, "texto": "Para reiniciar el router espera 30 segundos antes de encenderlo.", "estado": "destacado", "moderador_id": 2, "creado": "2026-11-04T19:39:09.740489-05:00", "actualizado": "2026-11-04T19:39:09.740489-05:00" }
]
```

La misma, en el contenedor (paso 2.5):

```json
[
  { "id": 4, "usuario_id": 1, "categoria_id": 1, "texto": "Comentario nuevo para la demostracion", "estado": "pendiente", "moderador_id": null, "creado": "2026-11-12T01:00:46.615151Z", "actualizado": "2026-11-12T01:00:46.615151Z" },
  { "id": 3, "usuario_id": 2, "categoria_id": 3, "texto": "Puedes filtrar los comentarios por estado.", "estado": "pendiente", "moderador_id": null, "creado": "2026-11-12T00:55:25.053107Z", "actualizado": "2026-11-12T00:55:25.053107Z" },
  { "id": 2, "usuario_id": 1, "categoria_id": 2, "texto": "La transferencia se verifica en 24 horas hábiles.", "estado": "aprobado", "moderador_id": 2, "creado": "2026-11-12T00:55:25.053107Z", "actualizado": "2026-11-12T00:55:25.053107Z" },
  { "id": 1, "usuario_id": 1, "categoria_id": 1, "texto": "Para reiniciar el router espera 30 segundos antes de encenderlo.", "estado": "destacado", "moderador_id": 2, "creado": "2026-11-12T00:55:25.053107Z", "actualizado": "2026-11-12T00:55:25.053107Z" }
]
```

Notas: el contenedor usa su propia base (`pg`, en el 5433), por eso los ids empiezan de cero y los datos son otros. Las peticiones 4 y 5 se hicieron con el registro nuevo (id 4). Todos los códigos coinciden con `go run`. Lo único que cambió es el formato de las fechas: `-05:00` con `go run` y `Z` (UTC) en el contenedor.

