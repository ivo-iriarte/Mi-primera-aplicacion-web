# Trabajo Práctico 

Este proyecto culmina la evolución de **CanchApp** (partiendo desde el dominio base en Go, persistencia con PostgreSQL y `sqlc`, exposición de API REST y renderizado SSR con `templ`), incorporando **interactividad reactiva en el frontend mediante HTMX**. De este modo, las acciones del usuario actualizan dinámicamente fragmentos específicos del DOM sin recargar la página completa y manteniendo una política estricta de **cero archivos JavaScript propios**.

---

## Cómo ejecutar el proyecto

Para clonar y poner en marcha la aplicación con su entorno completamente configurado:

```bash
git clone https://github.com/ivo-iriarte/Mi-primera-aplicacion-web.git
cd Mi-primera-aplicacion-web
git checkout tp6
make run-app
```

Una vez que el comando finalice las tareas previas e indique que el servidor está listo, ingresar desde cualquier navegador web a:

**http://localhost:8080/app** (la raíz `http://localhost:8080/` redirige automáticamente a `/app`).

Para detener la aplicación y limpiar el entorno, presionar `Ctrl + C` en la terminal donde se ejecutó el comando.

---

## Requisitos previos

Para poder ejecutar el proyecto es necesario tener instalados:
* **Go** (1.23 o superior)
* **Docker** y **Docker Compose**
* **Make**

*(Nota: Tanto `sqlc` como `templ` son descargados y ejecutados en sus versiones exactas directamente por el `Makefile` a través de `go run`, por lo que no es necesario tenerlos instalados de forma global en el sistema).*

---

## Automatización con Makefile (`make run-app`)

Siguiendo las pautas solicitadas por la cátedra, la ejecución del proyecto está completamente centralizada en un target automatizado que gestiona el ciclo de vida completo de la aplicación:

### 1. Tareas previas:
* **Limpieza de entorno previo (`db_down`):** Detiene y elimina contenedores y volúmenes anteriores (`docker compose down -v`) para asegurar una ejecución limpia sin datos residuales.
* **Levantado de la base de datos (`db_up`):** Inicia el contenedor PostgreSQL (`database_test`) y aguarda su disponibilidad mediante comprobaciones de salud (`--wait`).
* **Migración del esquema y datos de prueba:** Aplica el archivo `db/schema/schema.sql` y carga un conjunto inicial de datos de prueba (*fixtures*: clubes, disciplinas y canchas) para permitir la interacción inmediata con la interfaz.
* **Generación de código (`generate`):** Ejecuta `sqlc generate` para compilar las consultas SQL a código Go, `templ generate` para compilar los componentes de vista `.templ` a código Go nativo y tipado, y sincroniza las dependencias con `go mod tidy`.
* **Compilación (`build`):** Construye el binario ejecutable de la aplicación dentro del directorio `tmp/`.

### 2. Ejecución:
* Inicia el servidor web en el puerto `:8080`, sirviendo tanto la interfaz web reactiva en `/app` como los endpoints REST previamente desarrollados.

### 3. Tareas posteriores (limpieza automática):
* El script captura las señales de interrupción (`SIGINT` por `Ctrl + C`, `SIGTERM` o salida del proceso) mediante la instrucción `trap`. Al detenerse la aplicación, se disparan automáticamente las tareas de apagado y eliminación de contenedores y volúmenes creados (`make clean`), dejando el sistema en su estado original sin intervención manual.

---

## Interactividad con HTMX y Restricción de JavaScript

En estricto cumplimiento con la consigna de cursada, **el repositorio no contiene ningún archivo JavaScript (`.js`)**. La librería HTMX (`v2.0.4`) se carga directamente desde su CDN oficial en la plantilla base `views/layout.templ`.

La interacción se resuelve mediante el intercambio de **fragmentos HTML hipermedia** (*Hypermedia as the Engine of Application State*):

* **Navegación sin recarga (`hx-boost` y `hx-get`):** El menú de navegación y los enlaces entre deportes, canchas y reservas interceptan los clics y reemplazan únicamente el contenedor principal (`#main-content`), actualizando el historial del navegador mediante `hx-push-url="true"`.
* **Alta reactiva de deportes (`hx-post`):** Al registrar una nueva disciplina en `/app/deportes/crear`, el servidor responde únicamente con el componente parcial `DeportesGrid` actualizado (`hx-target="#lista-deportes"`), reseteando el formulario sin refrescar la vista.
* **Gestión dinámica de reservas (`hx-post`, `hx-swap` y `hx-confirm`):**
  * Al confirmar un turno en `/app/reservas/crear`, se actualiza únicamente el componente `ReservasTabla` (`#contenedor-tabla-reservas`) y se muestran validaciones de fechas sin perder los datos ingresados.
  * Al cancelar un turno en `/app/reservas/cancelar`, se solicita confirmación (`hx-confirm`) y se elimina la fila específica de la tabla (`hx-target="closest tr"`, `hx-swap="outerHTML swap:0.3s"`) con una transición suave.
* **Notificaciones Out-of-Band (`hx-swap-oob`):** El componente `FlashMessage` inyecta alertas visuales de éxito o error en `#toast-container` de forma paralela a la actualización principal del DOM.
* **Degradación elegante (*Progressive Enhancement*):** Todos los handlers inspeccionan el encabezado HTTP `HX-Request: true`. Si la petición proviene de HTMX, devuelven fragmentos parciales; si se accede de forma tradicional sin HTMX, responden con páginas completas o redirecciones estándar (`303 See Other`).

---

## Arquitectura y Documentación del Proyecto

El sistema mantiene una arquitectura desacoplada en tres capas (*Three-Tier*), consolidando el desarrollo de toda la cursada:

### 1. Capa de Presentación (`views/`):
* **`layout.templ`:** Estructura base HTML5, carga de HTMX vía CDN, estilos CSS responsivos (Grid y Flexbox), indicadores de carga (`.htmx-indicator`) y el componente global `FlashMessage`.
* **`deportes.templ`:** Define la vista completa `DeportesPage` y el fragmento parcial `DeportesGrid` para el listado y alta reactiva de disciplinas deportivas.
* **`canchas.templ`:** Define `CanchasPage` y el fragmento `CanchasPartial` con las canchas disponibles por deporte, precios, duración de turnos y enlace directo para reservar.
* **`reservas.templ`:** Define `ReservasPage` y el componente parcial `ReservasTabla` para el alta y cancelación reactiva de turnos.

### 2. Capa de Lógica del Negocio / Controladores (`handlers/`):
* **`handlers/views.go`:** Controlador web que discrimina peticiones HTMX (`isHTMX(r)`) de peticiones estándar, valida reglas de negocio (como consistencia de rangos horarios), consulta la base de datos vía `sqlc` y renderiza vistas completas o fragmentos parciales.
* **`handlers/deportes.go`:** Expone los endpoints de la API REST (`GET /deportes` y `GET /deportes/{id}`) en formato JSON.

### 3. Capa de Persistencia (`db/`):
* **`db/schema/schema.sql`:** Modelo relacional en PostgreSQL compuesto por las entidades principales del dominio: `institucion`, `deporte`, `cancha`, `usuario`, `reserva` y `estados_reservas`.
* **`db/queries/` y `db/sqlc/`:** Consultas SQL tipadas y compiladas a código Go mediante `sqlc` utilizando el driver `pgx/v5`.

---

## Endpoints y Rutas Disponibles

### Interfaz Web (SSR + HTMX)
* `GET /` → Redirige a `/app`.
* `GET /app` → Pantalla principal con grilla de deportes y formulario de alta.
* `POST /app/deportes/crear` → Crea un deporte (devuelve fragmento `DeportesGrid` + notificación OOB en peticiones HTMX).
* `GET /app/deportes/{id}/canchas` → Lista las canchas del deporte seleccionado (devuelve `CanchasPartial` en peticiones HTMX).
* `GET /app/reservas` → Panel de reservas con formulario (soporta preselección de cancha vía `?id_cancha=X`) y listado de turnos.
* `POST /app/reservas/crear` → Registra una nueva reserva (devuelve fragmento `ReservasTabla` + notificación OOB en peticiones HTMX).
* `POST /app/reservas/cancelar` → Elimina una reserva existente (elimina la fila `<tr>` del DOM y envía notificación OOB en peticiones HTMX).

### API REST
* `GET /deportes` → Retorna el listado de deportes en formato JSON.
* `GET /deportes/{id}` → Retorna el detalle de un deporte en formato JSON.

---

## Comandos del Makefile

* `make run-app`: Flujo principal automatizado (limpia entorno previo, levanta PostgreSQL, aplica esquema y fixtures, genera código de `sqlc` y `templ`, compila, ejecuta el servidor y limpia al salir).
* `make generate`: Sincroniza módulos Go y genera el código tipado a partir de `sqlc` y `templ`.
* `make build`: Genera el código y compila el binario ejecutable en `tmp/canchapp`.
* `make db_up`: Levanta exclusivamente el contenedor de PostgreSQL con el esquema y los datos iniciales cargados.
* `make db_down`: Detiene y elimina contenedores y volúmenes de Docker.
* `make clean`: Ejecuta `db_down` y elimina los archivos compilados en `tmp/`.
