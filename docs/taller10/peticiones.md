# Peticiones de la Semana 10 (con el servidor corriendo)

Mismas seis peticiones del Hito 2 en Bruno. Lo nuevo que se nota:
- La respuesta de un comentario ya no trae `eliminado` (viene del DTO).
- El POST ignora cualquier `estado` o `usuario_id` que se mande en el cuerpo.
- La petición 5 (aprobar/destacar) ahora corre por sqlc en una transacción.

Pegue aquí las capturas de las respuestas.

## Resultados

| Petición | Código |
|---|---|
| 0 crear registro | 201 |
| 1 sin token | 401 |
| 2 login | 200 |
| 3 con token | 200 |
| 4 miembro no puede | 403 |
| 5 moderador sí puede (sqlc + transacción) | 200 |
| 6 transición prohibida | 422 |

![Petición 5](bruno_5.png)

![Petición 6](bruno_6.png)

