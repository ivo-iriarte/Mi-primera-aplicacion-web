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

-- name: DeleteCancha :execrows
DELETE FROM cancha
WHERE id_cancha = $1;


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

-- name: CreateDeporte :one
INSERT INTO deporte (nombre)
VALUES ($1)
RETURNING *;
