# requests.hurl
# Pruebas reproducibles de la API REST completa de CanchApp (TP3)
# Ejecutar con: hurl --test requests.hurl

# ==========================================================
# 1. DEPORTES (CRUD)
# ==========================================================
POST http://localhost:8080/deportes
Content-Type: application/json
{
  "nombre": "Voley Test"
}
HTTP 201
[Asserts]
jsonpath "$.id_deporte" isInteger
jsonpath "$.nombre" == "Voley Test"
jsonpath "$.activo" == true
[Captures]
deporte_id: jsonpath "$.id_deporte"

GET http://localhost:8080/deportes?activos=true
HTTP 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$[*].nombre" contains "Voley Test"

GET http://localhost:8080/deportes/{{deporte_id}}
HTTP 200
[Asserts]
jsonpath "$.id_deporte" == {{deporte_id}}

PUT http://localhost:8080/deportes/{{deporte_id}}
Content-Type: application/json
{
  "nombre": "Voley Indoor",
  "activo": true
}
HTTP 200
[Asserts]
jsonpath "$.nombre" == "Voley Indoor"

# ==========================================================
# 2. INSTITUCIONES (CRUD)
# ==========================================================
POST http://localhost:8080/instituciones
Content-Type: application/json
{
  "nombre": "Polideportivo Hurl",
  "direccion": "Av. 59 1234",
  "telefono": "2262-421111",
  "email": "poli.hurl@club.com",
  "descripcion": "Complejo deportivo de prueba",
  "ciudad": "Necochea",
  "provincia": "Buenos Aires"
}
HTTP 201
[Asserts]
jsonpath "$.id_institucion" isInteger
jsonpath "$.nombre" == "Polideportivo Hurl"
[Captures]
institucion_id: jsonpath "$.id_institucion"

# ==========================================================
# 3. USUARIOS (CRUD)
# ==========================================================
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Lucia",
  "apellido": "Fernandez",
  "email": "lucia.hurl@test.com",
  "password_hash": "hash_demo_123",
  "telefono": "2262-500000"
}
HTTP 201
[Asserts]
jsonpath "$.id_usuario" isInteger
jsonpath "$.email" == "lucia.hurl@test.com"
[Captures]
usuario_id: jsonpath "$.id_usuario"

# ==========================================================
# 4. SERVICIOS Y CANCHAS (CRUD + FILTROS POR DEPORTE/INSTITUCION)
# ==========================================================
POST http://localhost:8080/servicios
Content-Type: application/json
{
  "servicio": "Estacionamiento Hurl",
  "descripcion": "Estacionamiento propio"
}
HTTP 201
[Captures]
servicio_id: jsonpath "$.id_servicio"

POST http://localhost:8080/canchas
Content-Type: application/json
{
  "nombre": "Cancha 3 Hurl",
  "precio": "15000.00",
  "duracion_turno_minutos": 60,
  "hora_apertura": "08:00:00",
  "hora_cierre": "12:00:00",
  "id_institucion": {{institucion_id}},
  "id_deporte": {{deporte_id}}
}
HTTP 201
[Asserts]
jsonpath "$.id_cancha" isInteger
jsonpath "$.nombre" == "Cancha 3 Hurl"
[Captures]
cancha_id: jsonpath "$.id_cancha"

# Asociar servicio a la cancha
POST http://localhost:8080/canchas/{{cancha_id}}/servicios
Content-Type: application/json
{
  "id_servicio": {{servicio_id}}
}
HTTP 201

GET http://localhost:8080/canchas/{{cancha_id}}/servicios
HTTP 200
[Asserts]
jsonpath "$[*].servicio" contains "Estacionamiento Hurl"

# Flujo de pantallas: elegir institución por deporte y luego cancha por institución y deporte
GET http://localhost:8080/instituciones?id_deporte={{deporte_id}}
HTTP 200
[Asserts]
jsonpath "$[*].id_institucion" contains {{institucion_id}}

GET http://localhost:8080/canchas?id_institucion={{institucion_id}}&id_deporte={{deporte_id}}
HTTP 200
[Asserts]
jsonpath "$[*].id_cancha" contains {{cancha_id}}

# ==========================================================
# 5. RESERVAS, DISPONIBILIDAD DE TURNOS Y CANCELACION
# ==========================================================
# Consultar turnos disponibles antes de reservar (09:00 debe estar disponible)
GET http://localhost:8080/canchas/{{cancha_id}}/turnos?fecha=2026-12-01
HTTP 200
[Asserts]
jsonpath "$[?(@.hora_inicio == '09:00')].disponible" nth 0 == true

# Crear reserva para el martes 2026-12-01 de 09:00 a 10:00
POST http://localhost:8080/reservas
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_cancha": {{cancha_id}},
  "fecha_inicio": "2026-12-01T09:00:00Z",
  "fecha_fin": "2026-12-01T10:00:00Z",
  "observaciones": "Turno de prueba Hurl"
}
HTTP 201
[Asserts]
jsonpath "$.id_reserva" isInteger
jsonpath "$.precio_total" == "15000.00"
[Captures]
reserva_id: jsonpath "$.id_reserva"

# Intentar reservar el mismo turno debe responder 409 Conflict
POST http://localhost:8080/reservas
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_cancha": {{cancha_id}},
  "fecha_inicio": "2026-12-01T09:00:00Z",
  "fecha_fin": "2026-12-01T10:00:00Z"
}
HTTP 409

# Verificar que el turno de 09:00 ya NO está disponible
GET http://localhost:8080/canchas/{{cancha_id}}/turnos?fecha=2026-12-01
HTTP 200
[Asserts]
jsonpath "$[?(@.hora_inicio == '09:00')].disponible" nth 0 == false

# Pantalla "Mis reservas" (con detalle de cancha, institución, deporte y estado)
GET http://localhost:8080/reservas?id_usuario={{usuario_id}}
HTTP 200
[Asserts]
jsonpath "$[0].cancha_nombre" == "Cancha 3 Hurl"
jsonpath "$[0].institucion_nombre" == "Polideportivo Hurl"
jsonpath "$[0].deporte_nombre" == "Voley Indoor"
jsonpath "$[0].estado_nombre" == "Confirmada"

# Cancelar la reserva (cambia estado a Cancelada y libera el turno)
POST http://localhost:8080/reservas/{{reserva_id}}/cancelar
HTTP 200

# Cancelar por segunda vez debe responder 409 Conflict
POST http://localhost:8080/reservas/{{reserva_id}}/cancelar
HTTP 409

# Verificar que el turno de 09:00 volvió a estar disponible
GET http://localhost:8080/canchas/{{cancha_id}}/turnos?fecha=2026-12-01
HTTP 200
[Asserts]
jsonpath "$[?(@.hora_inicio == '09:00')].disponible" nth 0 == true

# ==========================================================
# 6. LIMPIEZA (DELETE)
# ==========================================================
DELETE http://localhost:8080/reservas/{{reserva_id}}
HTTP 204

DELETE http://localhost:8080/canchas/{{cancha_id}}/servicios/{{servicio_id}}
HTTP 204

DELETE http://localhost:8080/canchas/{{cancha_id}}
HTTP 204

DELETE http://localhost:8080/servicios/{{servicio_id}}
HTTP 204

DELETE http://localhost:8080/usuarios/{{usuario_id}}
HTTP 204

DELETE http://localhost:8080/instituciones/{{institucion_id}}
HTTP 204

DELETE http://localhost:8080/deportes/{{deporte_id}}
HTTP 204

GET http://localhost:8080/deportes/{{deporte_id}}
HTTP 404
