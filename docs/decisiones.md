## D1 · ¿Quién pone el estado al crear?

*Opciones:* 
Que el cliente envíe el estado al crear el comentario, o que el servidor lo asigne automáticamente por seguridad.

*Qué elegimos:* 
Que el servidor imponga el estado inicial de forma automática.

*Por qué, en nuestro negocio:* 
En nuestra plataforma de soporte comunitario, un usuario no debe tener la capacidad de autoasignarse un estado de moderación avanzado (como "aprobado" o "destacado") al momento de publicar. Por control de calidad y seguridad, el servidor fuerza obligatoriamente que todo nuevo comentario nazca como "pendiente".

*Qué pasaría con la otra opción:* 
Si dejáramos que el cliente envíe el estado desde el formulario, cualquier usuario malintencionado podría crear registros marcados directamente como "destacados" o "aprobados", evadiendo por completo el filtro de la comunidad y alterando el sistema de reputación y medallas.