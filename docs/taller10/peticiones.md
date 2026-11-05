# Peticiones de la Semana 10 (con el servidor corriendo)

Mismas seis peticiones del Hito 2 en Bruno. Lo nuevo que se nota:
- La respuesta de un comentario ya no trae `eliminado` (viene del DTO).
- El POST ignora cualquier `estado` o `usuario_id` que se mande en el cuerpo.
- La petición 5 (aprobar/destacar) ahora corre por sqlc en una transacción.

Pegue aquí las capturas de las respuestas.
