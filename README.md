# **Trabajo práctico**

Este proyecto evoluciona las etapas anteriores (persistencia con PostgreSQL y sqlc, y exposición de API REST) incorporando una capa de presentación interactiva renderizada del lado del servidor (SSR) mediante la librería templ.

## **Cómo ejecutar el proyecto**

Para clonar y poner en marcha la aplicación con su entorno completamente configurado:

git clone https://github.com/ivo-iriarte/Mi-primera-aplicacion-web.git
cd Mi-primera-aplicacion-web
git checkout tp5
make run-app

Una vez que el comando finalice las tareas previas e indique que el servidor está listo, ingresar desde cualquier navegador web a:
http://localhost:8080/app (la raíz http://localhost:8080/ redirige automáticamente a /app).
Para detener la aplicación y limpiar el entorno, presionar Ctrl + C en la terminal donde se ejecutó el comando.

## **Automatización con Makefile (make run-app)**

Siguiendo las pautas solicitadas por la cátedra, la ejecución del proyecto está completamente centralizada en un target automatizado que gestiona el ciclo de vida completo de la aplicación:
Tareas previas:
  Limpieza de entorno previo (db_down): Detiene y elimina contenedores y volúmenes anteriores (docker compose down -v) para asegurar una ejecución limpia sin datos residuales.
  Levantado de la base de datos (db_up): Inicia el contenedor PostgreSQL (database_test) y aguarda su disponibilidad mediante comprobaciones de salud (--wait).
  Migración del esquema y datos de prueba: Aplica el archivo db/schema/schema.sql y carga un conjunto inicial de datos de prueba (fixtures: clubes, disciplinas y canchas) para permitir la interacción    inmediata con la interfaz.
  Verificación de restricciones (check-no-js): Valida de manera automática que no existan archivos .js en el repositorio antes de compilar.
  Generación de código (generate): Ejecuta sqlc generate para compilar las consultas SQL a código Go, templ generate para compilar los componentes de vista .templ a código Go nativo y tipado, y          sincroniza las dependencias con go mod tidy.
  Compilación (build): Construye el binario ejecutable de la aplicación dentro del directorio tmp/.

Ejecución:
  Inicia el servidor web en el puerto :8080, sirviendo tanto la interfaz web en /app como los endpoints REST previamente desarrollados.

Tareas posteriores (limpieza automática):
  El script captura las señales de interrupción (SIGINT por Ctrl + C, SIGTERM o salida del proceso) mediante la instrucción trap. Al detenerse la aplicación, se disparan automáticamente las tareas de    apagado y eliminación de contenedores y volúmenes creados (make clean), dejando el sistema en su estado original sin intervención manual.

## **Desarrollo y Arquitectura**

El sistema respeta una arquitectura en tres capas (Three-Tier):

Capa de Presentación (views/):
  layout.templ: Define la estructura base HTML5 común, metadatos y la hoja de estilos CSS responsiva (utilizando CSS Grid y Flexbox).
  deportes.templ: Vista principal que lista los deportes disponibles y permite registrar nuevas disciplinas mediante formulario.
  canchas.templ: Muestra las canchas asociadas a una disciplina deportiva seleccionada, detallando precios, duración de turnos y franja horaria.
  reservas.templ: Panel de gestión ("Mis Turnos") que permite reservar turnos para una cancha (con soporte de preselección vía ?id_cancha=X) y visualizar/cancelar reservas existentes.

Capa de Lógica del Negocio / Controladores (handlers/):
  handlers/views.go: Controlador encargado de recibir las peticiones web, interactuar con la base de datos a través de sqlc, inyectar datos tipados en los componentes de templ y renderizarlos hacia el   http.ResponseWriter.
  handlers/deportes.go: Mantiene disponibles los endpoints de la API REST implementada en etapas anteriores (GET /deportes y GET /deportes/{id}).

Capa de Persistencia (db/):
  Continúa utilizando PostgreSQL junto con sqlc y el driver pgx/v5 para garantizar consultas SQL seguras, verificadas en tiempo de compilación y libres de la sobrecarga de un ORM tradicional.

