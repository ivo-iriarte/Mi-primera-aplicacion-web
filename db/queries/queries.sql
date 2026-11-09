-- CANCHA

-- name: CreateCancha :one
INSERT into cancha(nombre,precio,duracion_turno_minutos,hora_apertura,hora_cierre,id_institucion,id_deporte)
VALUES($1, $2, $3, $4, $5, $6,$7)
RETURNING *;

-- name: GetCancha :one
SELECT id_cancha,nombre, precio, activa, duracion_turno_minutos, hora_apertura,hora_cierre,id_institucion,id_deporte
FROM cancha
WHERE id_cancha = $1;

-- name: ListCanchas :many
SELECT id_cancha,nombre, precio, activa, duracion_turno_minutos, hora_apertura,hora_cierre,id_institucion,id_deporte
FROM cancha;

-- name: ListCanchasActivas :many
SELECT id_cancha,nombre, precio, activa, duracion_turno_minutos, hora_apertura,hora_cierre,id_institucion,id_deporte
FROM cancha
WHERE activa = TRUE;

-- name: UpdatePrecioCancha :execrows
UPDATE cancha 
SET precio = $2, updated_at = NOW()
WHERE id_cancha = $1;

-- name: UpdateCancha :one
UPDATE cancha
SET nombre = $2,
    precio = $3,
    activa = $4,
    duracion_turno_minutos = $5,
    hora_apertura = $6,
    hora_cierre = $7,
    id_institucion = $8,
    id_deporte = $9,
    updated_at = NOW()
WHERE id_cancha = $1
RETURNING *;

-- name: DeleteCancha :execrows
DELETE FROM cancha
WHERE id_cancha = $1;

-- INSTITUCION

-- name: CreateInstitucion :one
INSERT INTO institucion (
    nombre,
    direccion,
    telefono,
    email,
    descripcion,
    ciudad,
    provincia
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetInstitucion :one
SELECT * FROM institucion
WHERE id_institucion = $1;

-- name: ListInstituciones :many
SELECT * FROM institucion
ORDER BY id_institucion;

-- name: UpdateInstitucion :one
UPDATE institucion
SET nombre = $2,
    direccion = $3,
    telefono = $4,
    email = $5,
    descripcion = $6,
    activo = $7,
    ciudad = $8,
    provincia = $9,
    updated_at = NOW()
WHERE id_institucion = $1
RETURNING *;

-- name: DeleteInstitucion :execrows
DELETE FROM institucion
WHERE id_institucion = $1;

-- DEPORTE

-- name: CreateDeporte :one
INSERT INTO deporte (nombre)
VALUES ($1)
RETURNING *;

-- name: GetDeporte :one
SELECT * FROM deporte
WHERE id_deporte = $1;

-- name: ListDeportes :many
SELECT * FROM deporte
ORDER BY id_deporte;

-- name: UpdateDeporte :one
UPDATE deporte
SET nombre = $2,
    activo = $3,
    updated_at = NOW()
WHERE id_deporte = $1
RETURNING *;

-- name: DeleteDeporte :execrows
DELETE FROM deporte
WHERE id_deporte = $1;

-- USUARIO

-- name: CreateUsuario :one
INSERT INTO usuario (
    nombre,
    apellido,
    email,
    password_hash,
    telefono
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUsuario :one
SELECT * FROM usuario
WHERE id_usuario = $1;

-- name: GetUsuarioByEmail :one
SELECT * FROM usuario
WHERE email = $1;

-- name: ListUsuarios :many
SELECT * FROM usuario
ORDER BY id_usuario;

-- name: UpdateUsuario :one
UPDATE usuario
SET nombre = $2,
    apellido = $3,
    email = $4,
    password_hash = $5,
    telefono = $6,
    activo = $7,
    updated_at = NOW()
WHERE id_usuario = $1
RETURNING *;

-- name: DeleteUsuario :execrows
DELETE FROM usuario
WHERE id_usuario = $1;

-- ESTADOS_RESERVAS

-- name: CreateEstadoReserva :one
INSERT INTO estados_reservas (nombre)
VALUES ($1)
RETURNING *;

-- name: GetEstadoReserva :one
SELECT * FROM estados_reservas
WHERE id_estados_reserva = $1;

-- name: ListEstadosReservas :many
SELECT * FROM estados_reservas
ORDER BY id_estados_reserva;

-- name: UpdateEstadoReserva :one
UPDATE estados_reservas
SET nombre = $2
WHERE id_estados_reserva = $1
RETURNING *;

-- name: DeleteEstadoReserva :execrows
DELETE FROM estados_reservas
WHERE id_estados_reserva = $1;

-- SERVICIOS

-- name: CreateServicio :one
INSERT INTO servicios (servicio, descripcion)
VALUES ($1, $2)
RETURNING *;

-- name: GetServicio :one
SELECT * FROM servicios
WHERE id_servicio = $1;

-- name: ListServicios :many
SELECT * FROM servicios
ORDER BY id_servicio;

-- name: UpdateServicio :one
UPDATE servicios
SET servicio = $2,
    descripcion = $3,
    activo = $4
WHERE id_servicio = $1
RETURNING *;

-- name: DeleteServicio :execrows
DELETE FROM servicios
WHERE id_servicio = $1;

-- CANCHA_SERVICIO 

-- name: AddServicioToCancha :exec
INSERT INTO cancha_servicio (id_cancha, id_servicio)
VALUES ($1, $2);

-- name: RemoveServicioFromCancha :exec
DELETE FROM cancha_servicio
WHERE id_cancha = $1 AND id_servicio = $2;

-- name: ListServiciosByCancha :many
SELECT s.*
FROM servicios s
JOIN cancha_servicio cs ON s.id_servicio = cs.id_servicio
WHERE cs.id_cancha = $1;

-- RESERVAS

-- name: CreateReserva :one
INSERT INTO reservas (
    id_usuario,
    id_cancha,
    fecha_inicio,
    fecha_fin,
    id_estado,
    precio_total,
    observaciones
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetReserva :one
SELECT * FROM reservas
WHERE id_reserva = $1;

-- name: ListReservas :many
SELECT * FROM reservas
ORDER BY fecha_inicio DESC;

-- name: ListReservasByCancha :many
SELECT * FROM reservas
WHERE id_cancha = $1
ORDER BY fecha_inicio;

-- name: ListReservasByUsuario :many
SELECT * FROM reservas
WHERE id_usuario = $1
ORDER BY fecha_inicio DESC;

-- name: UpdateReservaEstado :one
UPDATE reservas
SET id_estado = $2,
    updated_at = NOW()
WHERE id_reserva = $1
RETURNING *;

-- name: UpdateReserva :one
UPDATE reservas
SET id_usuario = $2,
    id_cancha = $3,
    fecha_inicio = $4,
    fecha_fin = $5,
    id_estado = $6,
    precio_total = $7,
    observaciones = $8,
    updated_at = NOW()
WHERE id_reserva = $1
RETURNING *;

-- name: DeleteReserva :execrows
DELETE FROM reservas
WHERE id_reserva = $1;
