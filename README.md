# Trabajo Práctico 6 - Interfaces Dinámicas con HTMX

Este proyecto culmina la evolución de **CanchApp** (partiendo desde el dominio base en Go, persistencia con PostgreSQL y `sqlc`, exposición de API REST y renderizado SSR con `templ`), incorporando **interactividad reactiva en el frontend mediante HTMX 4.0**. De este modo, las acciones del usuario actualizan dinámicamente fragmentos específicos del DOM sin recargar la página completa y manteniendo una política estricta de **cero JavaScript propio** (sin archivos `.js` ni scripts inline).

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
* **Migración del esquema y datos de prueba:** Aplica el archivo `db/schema/schema.sql` y carga un conjunto inicial de datos de prueba (*fixtures*: institución, deportes, canchas y estados de reserva) para permitir la interacción inmediata con la interfaz.
* **Verificación de ausencia de JS (`check-no-js`):** Comprueba automáticamente que no existan archivos `.js` en el repositorio.
* **Generación de código (`generate`):** Ejecuta `sqlc generate` para compilar las consultas SQL a código Go, `templ generate` para compilar los componentes de vista `.templ` a código Go nativo y tipado, y sincroniza las dependencias con `go mod tidy`.
* **Análisis y compilación (`build`):** Ejecuta `go vet ./...` y construye el binario ejecutable de la aplicación dentro del directorio `tmp/`.

### 2. Ejecución:
* Inicia el servidor web en el puerto `:8080`, sirviendo los estilos estáticos en `/static/`, la interfaz web reactiva en `/app` y los endpoints REST previamente desarrollados.

### 3. Tareas posteriores (limpieza automática):
* El servidor implementa *graceful shutdown* ante señales `SIGINT`/`SIGTERM`, y el `Makefile` captura la salida mediante `trap` para disparar automáticamente el apagado y eliminación de contenedores y volúmenes (`make clean`), dejando el sistema en su estado original sin intervención manual.

---

## Interactividad con HTMX 4.0 y Cero JavaScript

En estricto cumplimiento con la consigna y la teoría de la cátedra, **el proyecto no contiene archivos JavaScript (`.js`) ni código JS inline**. La biblioteca **HTMX (`v4.0.0`)** se incluye con verificación de integridad (`integrity` SHA-384) desde su CDN en el `<head>` de `views/layout.templ`.

La interacción se resuelve íntegramente mediante **intercambio declarativo de fragmentos HTML**:

* **Navegación fluida (`hx-boost:inherited="true"` y `hx-get`):** En `layout.templ` se habilita la herencia explícita de `hx-boost` en HTMX 4.0 sobre `<body>`, mientras que los enlaces internos entre deportes, canchas y reservas actualizan dinámicamente `#main-content` y sincronizan la URL del navegador con `hx-push-url="true"`.
* **Creación sin recarga + limpieza automática de formularios con Out-of-Band Swap (`hx-post` + `hx-swap-oob="true"`):**
  * Los formularios están desacoplados en sus propios componentes (`DeporteForm` y `ReservaForm` en `views/forms.templ`).
  * Al crear un deporte (`POST /app/deportes`) o una reserva (`POST /app/reservas`), se elimina la redirección y el servidor responde con la lista actualizada (`DeportesGrid` / `ReservasTabla`), un formulario limpio con `hx-swap-oob="true"` (`DeporteFormOOB` / `ReservaFormOOB`) y una notificación `FlashMessage` OOB. Si ocurre un error de validación, no se envía el swap del formulario para preservar los datos ingresados por el usuario.
* **Borrado sin recarga (`hx-delete` + `200 OK` vacío):**
  * Tanto en las tarjetas de deportes (`DELETE /app/deportes/{id}`) como en las filas de la tabla de reservas (`DELETE /app/reservas/{id}`), los botones utilizan `hx-delete`, `hx-target` (`closest .deporte-card` / `closest tr`), `hx-swap="outerHTML swap:300ms"` y confirmación con `hx-confirm`.
  * El servidor extrae el `{id}` de la URL, ejecuta el `Delete...` generado por `sqlc` y devuelve una respuesta vacía con estado `200 OK`, eliminando el elemento del DOM con una transición CSS suave. Si existe una restricción de integridad referencial (ej. deporte con canchas asociadas), responde con `HX-Reswap: none` y muestra la alerta OOB correspondiente.
* **Indicadores de carga (`hx-indicator`):** Feedback visual (`.htmx-indicator`) durante el envío de formularios.
* **Degradación elegante (*Progressive Enhancement*):** Los handlers inspeccionan `HX-Request: true` (y `HX-Request-Type`); ante accesos directos sin HTMX devuelven documentos HTML completos o redirecciones `303 See Other` (patrón PRG).

---

## Arquitectura y Estructura del Proyecto

### 1. Capa de Presentación (`views/` y `static/`):
* **`static/styles.css`:** Hoja de estilos externa con variables CSS, diseño responsivo (CSS Grid, Flexbox y `@media` queries), estilos semánticos, transiciones `.htmx-swapping` e indicadores `.htmx-indicator`.
* **`views/layout.templ`:** Documento base HTML5 (`lang="es"`), enlace a `/static/styles.css`, inclusión de HTMX 4.0, navegación `hx-boost:inherited="true"` y componente `FlashMessage` con soporte OOB.
* **`views/forms.templ`:** Componentes reutilizables de formularios (`DeporteForm`, `ReservaForm`) con etiquetas `<label for="...">` accesibles y sus variantes Out-of-Band (`DeporteFormOOB`, `ReservaFormOOB`).
* **`views/deportes.templ`:** Vista `DeportesPage`, contenido parcial `DeportesContent` y grilla `DeportesGrid` con acceso a canchas y borrado reactivo (`hx-delete`).
* **`views/canchas.templ`:** Vista `CanchasPage` y fragmento `CanchasPartial` con las canchas activas filtradas por deporte.
* **`views/reservas.templ`:** Vista `ReservasPage`, contenido parcial `ReservasContent` y tabla `ReservasTabla` con cancelación reactiva de turnos (`hx-delete`).

### 2. Capa de Controladores HTTP (`handlers/`):
* **`handlers/views.go`:** Maneja la interfaz SSR/HTMX, valida formularios (`r.ParseForm()`), coordina consultas `sqlc` manejando errores explícitamente y compone fragmentos parciales + swaps OOB.
* **`handlers/deportes.go`:** Expone los endpoints JSON de la API REST (`/deportes` y `/deportes/{id}`).

### 3. Capa de Persistencia (`db/`):
* **`db/schema/schema.sql`:** Esquema relacional en PostgreSQL (`institucion`, `deporte`, `cancha`, `usuario`, `reservas`, `estados_reservas`, `servicios`, `cancha_servicio`).
* **`db/queries/queries.sql`:** Consultas SQL parametrizadas compiladas a código Go tipado mediante `sqlc` y driver `pgx/v5`.

---

## Endpoints y Rutas Disponibles

### Interfaz Web (SSR + HTMX 4.0)
* `GET /` → Redirige a `/app`.
* `GET /static/*` → Sirve hojas de estilo y recursos estáticos.
* `GET /app` → Pantalla principal con grilla de deportes y formulario de alta.
* `POST /app/deportes` → Crea un deporte (devuelve `DeportesGrid` + `DeporteFormOOB` + `FlashMessage` OOB).
* `DELETE /app/deportes/{id}` → Elimina un deporte sin canchas asociadas (devuelve `200 OK` vacío para remover la tarjeta del DOM).
* `GET /app/deportes/{id}/canchas` → Lista las canchas activas del deporte seleccionado.
* `GET /app/reservas` → Panel de reservas con formulario (soporta `?id_cancha=X`) y tabla de turnos.
* `POST /app/reservas` → Registra una nueva reserva (devuelve `ReservasTabla` + `ReservaFormOOB` + `FlashMessage` OOB).
* `DELETE /app/reservas/{id}` → Elimina una reserva existente (devuelve `200 OK` vacío para remover la fila `<tr>` del DOM).

### API REST
* `GET /deportes`, `POST /deportes`
* `GET /deportes/{id}`, `PUT /deportes/{id}`, `DELETE /deportes/{id}`

---

## Comandos del Makefile

* `make run-app`: Flujo principal automatizado (limpia entorno previo, levanta PostgreSQL, aplica esquema y fixtures, verifica cero JS, genera código de `sqlc` y `templ`, compila, ejecuta el servidor y limpia al salir).
* `make check-no-js`: Verifica que no existan archivos `.js` en el proyecto.
* `make generate`: Ejecuta `sqlc generate`, `templ generate` y `go mod tidy`.
* `make build`: Ejecuta `check-no-js`, `generate`, `go vet` y compila el binario en `tmp/canchapp`.
* `make test`: Levanta PostgreSQL de prueba, genera el código y ejecuta `go vet ./...` y `go test -v ./...`.
* `make db_up` / `make db_down` / `make clean`: Gestión y limpieza de contenedores y artefactos.
