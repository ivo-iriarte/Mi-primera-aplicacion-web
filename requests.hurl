# requests.hurl

# 1. Crear un nuevo deporte (POST /deportes)
POST http://localhost:8080/deportes
Content-Type: application/json
{
  "nombre": "Padel"
}
HTTP 201
[Asserts]
jsonpath "$.id_deporte" isInteger
jsonpath "$.nombre" == "Padel"
jsonpath "$.activo" == true
[Captures]
deporte_id: jsonpath "$.id_deporte"

# 2. Obtener todos los deportes (GET /deportes)
GET http://localhost:8080/deportes
HTTP 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$[*].nombre" contains "Padel"

# 3. Obtener el deporte creado por ID (GET /deportes/{id})
GET http://localhost:8080/deportes/{{deporte_id}}
HTTP 200
[Asserts]
jsonpath "$.id_deporte" == {{deporte_id}}
jsonpath "$.nombre" == "Padel"

# 4. Actualizar el deporte (PUT /deportes/{id})
PUT http://localhost:8080/deportes/{{deporte_id}}
Content-Type: application/json
{
  "nombre": "Padel Indoor",
  "activo": false
}
HTTP 200
[Asserts]
jsonpath "$.id_deporte" == {{deporte_id}}
jsonpath "$.nombre" == "Padel Indoor"
jsonpath "$.activo" == false

# 5. Eliminar el deporte (DELETE /deportes/{id})
DELETE http://localhost:8080/deportes/{{deporte_id}}
HTTP 204

# 6. Verificar que ya no existe y devuelve 404 (GET /deportes/{id})
GET http://localhost:8080/deportes/{{deporte_id}}
HTTP 404
