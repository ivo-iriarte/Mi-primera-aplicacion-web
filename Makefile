APP_NAME := canchapp
TEMPL_VERSION := v0.3.833
SQLC_VERSION := v1.31.1

.PHONY: all run-app generate build db_down db_up clean check-no-js

all: run-app

# Verificación de consigna: sin archivos .js
check-no-js:
	@echo "Verificando ausencia de archivos .js..."
	@if find . -name "*.js" -not -path "./vendor/*" | grep -q .; then \
		echo "ERROR: Se encontraron archivos .js en el repositorio:"; \
		find . -name "*.js" -not -path "./vendor/*"; \
		exit 1; \
	fi
	@echo "OK: Sin archivos .js."

# Tareas previas: sqlc + templ generate
generate:
	@go mod tidy
	@echo "Generando código de sqlc y templ..."
	@go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate
	@go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION) generate

# Compilación
build: check-no-js generate
	@echo "Compilando aplicación..."
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

# Limpieza de Docker
db_down:
	@echo "Deteniendo y eliminando contenedores y volúmenes..."
	@docker compose -f docker-compose.test.yml down -v --remove-orphans 2>/dev/null || true

# Tareas previas: Levantar Postgres, esperar healthcheck y precargar esquema y fixtures
db_up: db_down
	@echo "Levantando PostgreSQL..."
	@docker compose -f docker-compose.test.yml up -d --wait --wait-timeout 60
	@echo "Cargando esquema..."
	@docker compose -f docker-compose.test.yml exec -T database_test \
		psql -v ON_ERROR_STOP=1 -U prueba -d CanchAppprueba < db/schema/schema.sql
	@echo "Insertando datos iniciales (fixtures)..."
	@docker compose -f docker-compose.test.yml exec -T database_test \
		psql -v ON_ERROR_STOP=1 -U prueba -d CanchAppprueba -c " \
		INSERT INTO institucion (nombre, direccion, telefono, email, descripcion, ciudad, provincia) VALUES \
		('Club Náutico Necochea', 'Ribera Río Quequén', '2262-420000', 'nautico@club.com', 'Club con canchas e instalaciones completas', 'Necochea', 'Buenos Aires') ON CONFLICT DO NOTHING; \
		INSERT INTO deporte (nombre) VALUES ('Fútbol'), ('Padel'), ('Tenis') ON CONFLICT DO NOTHING; \
		INSERT INTO cancha (nombre, precio, duracion_turno_minutos, hora_apertura, hora_cierre, id_institucion, id_deporte) VALUES \
		('Cancha Sintético 1', 18000.00, 60, '08:00:00', '23:00:00', 1, 1), \
		('Cancha Padel Cristal', 14000.00, 90, '09:00:00', '23:00:00', 1, 2) ON CONFLICT DO NOTHING; \
		INSERT INTO estados_reservas (nombre) VALUES ('Confirmada'), ('Cancelada') ON CONFLICT DO NOTHING;"

# TARGET TP5: Tareas previas -> Ejecución -> Tareas posteriores con trap
run-app: db_up build
	@echo "=================================================="
	@echo " CanchApp iniciada en: http://localhost:8080/app"
	@echo " Presione Ctrl+C para finalizar y limpiar."
	@echo "=================================================="
	@trap 'echo "\nLimpiando contenedores y volúmenes..."; $(MAKE) clean; exit 0' INT TERM EXIT; \
	./tmp/$(APP_NAME)

clean: db_down
	@rm -rf tmp
