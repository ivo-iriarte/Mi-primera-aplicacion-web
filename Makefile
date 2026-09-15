SQLC_VERSION := v1.31.1

.PHONY: test generate build db_test

test:
	@trap 'docker compose -f docker-compose.test.yml rm -sf database_test' EXIT; \
	$(MAKE) generate && go test -count=1 ./...

generate: db_test
	@go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

build: 
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

db_test: 
	@docker compose -f docker-compose.test.yml up -d --wait --wait-timeout 60 && \
 	docker compose -f docker-compose.test.yml exec -T database_test \
    psql -v ON_ERROR_STOP=1 -U prueba -d CanchAppprueba < db/schema/schema.sql
