APP_NAME := canchapp

.PHONY: all generate build db_down db_up test clean

all: test

# Ejecutamos el sqlc generate
generate:
	@sqlc generate

# Construimos el binario de la aplicacion
build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

# Borramos los contenedores y volumenes previos
db_down:
	@docker compose -f docker-compose.test.yml down -v

# Levantamos contenedores y como se nos indico esperamos a que la base este corriendo para luego cargar el esquema
db_up: db_down
	@docker compose -f docker-compose.test.yml up -d --wait --wait-timeout 60
	@docker compose -f docker-compose.test.yml exec -T database_test \
		psql -v ON_ERROR_STOP=1 -U prueba -d CanchAppprueba < db/schema/schema.sql

# Aca lo que hacemos son las siguientes cosas
# Ejecutamos los tests (continuamos con los test del TP2)
# Pruobamos los endpoints con Hurl
test: build db_up
	@./tmp/$(APP_NAME) & APP_PID=$$!; \
	trap "kill $$APP_PID 2>/dev/null; $(MAKE) db_down" EXIT; \
	sleep 2; \
	go test -v -count=1 ./...; \
	hurl --test requests.hurl

# Limpiamos los contenedores y volumenes
clean: db_down
	@rm -rf tmp
