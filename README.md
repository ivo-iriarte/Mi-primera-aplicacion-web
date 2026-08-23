# Mi-primera-aplicacion-web
Trabajo especial para la cursada de la materia programacion web realizado por Martin Santos Abasolo e Ivo Iriarte, a cargo de los profesores Alejandro Zunino y Alfredo Raul Teyseyre

Cabe aclarar que es un proyecto pensado hace tiempo, por lo que se cuenta con un analisis bastante bueno del dominio, asi como tambien con el esquema entidad-relacion del proyecto y los codigos de creacion de tablas en postgree

# Aplicación de Reserva de Canchas

## Descripción del dominio

La aplicación tendrá como objetivo permitir que los usuarios puedan consultar y realizar reservas de canchas deportivas pertenecientes a distintas instituciones.

El sistema administrará información sobre los usuarios, las instituciones, las canchas disponibles, los deportes que se pueden practicar, los servicios ofrecidos y las reservas realizadas.

## Entidades principales

### Usuario

Representa a las personas que utilizarán la aplicación para realizar reservas.

Se almacenará la siguiente información:

* Identificador del usuario.
* Nombre.
* Apellido.
* Email.
* Contraseña almacenada de forma segura.
* Teléfono.
* Estado del usuario (activo o inactivo).
* Fecha de creación.
* Fecha de última actualización.

### Instituciones

Representa a los establecimientos que poseen las canchas disponibles para reservar.

Se almacenará:

* Identificador de la institución.
* Nombre.
* Dirección.
* Teléfono.
* Email.
* Descripción.
* Ciudad.
* Provincia.
* Estado de la institución (activa o inactiva).
* Fecha de creación.
* Fecha de última actualización.

### Canchas

Representa cada una de las canchas disponibles dentro de una institución.

Se almacenará:

* Identificador de la cancha.
* Institución a la que pertenece.
* Nombre.
* Deporte asociado.
* Precio.
* Estado de la cancha (activa o inactiva).
* Duración de los turnos en minutos.
* Hora de apertura.
* Hora de cierre.
* Fecha de creación.
* Fecha de última actualización.

### Deportes

Representa los distintos deportes para los cuales pueden existir canchas.

Se almacenará:

* Identificador del deporte.
* Nombre.
* Estado (activo o inactivo).
* Fecha de creación.
* Fecha de última actualización.

### Reservas

Representa una reserva realizada por un usuario sobre una cancha determinada.

Se almacenará:

* Identificador de la reserva.
* Usuario que realizó la reserva.
* Cancha reservada.
* Fecha y hora de inicio.
* Fecha y hora de finalización.
* Estado de la reserva.
* Precio total.
* Observaciones.
* Fecha de creación.
* Fecha de última actualización.

### Estados de reserva

Representa los posibles estados en los que puede encontrarse una reserva.

Se almacenará:

* Identificador del estado.
* Nombre del estado.

Por ejemplo, una reserva podría encontrarse pendiente, confirmada o cancelada.

### Servicios

Representa los servicios adicionales que pueden ofrecer las distintas canchas.

Se almacenará:

* Identificador del servicio.
* Nombre del servicio.
* Descripción.
* Estado (activo o inactivo).

### Cancha - Servicio

Esta entidad permite relacionar las canchas con los servicios que ofrece cada una.

Se almacenará:

* Identificador de la cancha.
* Identificador del servicio.

De esta manera, una cancha puede disponer de varios servicios y un mismo servicio puede estar disponible en distintas canchas.

## Ejecución del proyecto

### Requisitos

Para ejecutar el proyecto es necesario tener instalado **Go**.

### Ejecutar el servidor
se debe contar con GIT instalado, y GO.
se debera clonar el repositorio, mediante el comando:
```bash
git clone https://github.com/ivo-iriarte/Mi-primera-aplicacion-web.git
```

Desde una terminal, ubicarse dentro de la carpeta del proyecto y ejecutar:

```bash
go run .
```

El servidor web se iniciará en el puerto **8080**.

Luego, abrir un navegador web e ingresar a:

```text
http://localhost:8080
```

Al acceder a esa dirección se mostrará la página `index.html` de la aplicación.
