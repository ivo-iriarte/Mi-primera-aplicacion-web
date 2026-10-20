package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// "github.com/joho/godotenv"

	sqlc "canchapp/db/sqlc"
	"canchapp/handlers"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// func abrirDB() (*sql.DB, error) {
// 	user := os.Getenv("POSTGRES_USER")
// 	password := os.Getenv("POSTGRES_PASSWORD")
// 	dbname := os.Getenv("POSTGRES_DB")

// 	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=localhost port=5432 sslmode=disable", user, password, dbname)
// 	db, err := sql.Open("pgx", connStr)

// 	if err != nil {
// 		return nil, fmt.Errorf("error al iniciar la base de datos : %w", err)
// 	}

// 	if err := db.Ping(); err != nil {
// 		db.Close()
// 		return nil, fmt.Errorf("error en la conexion con la BD: %w", err)
// 	}

// 	db.SetMaxOpenConns(25)
// 	return db, nil
// }

// todo esto era del original y tomaba las variables del .env pero por ahora usamos las variables del de testeo incrustadas eb el codigo ya que el .env no lo subimos a ds

func abrirDB() (*sql.DB, error) {

	db, err := sql.Open("pgx", "user=prueba password=prueba dbname=CanchAppprueba host=localhost port=5433 sslmode=disable")

	if err != nil {
		return nil, fmt.Errorf("error al iniciar la base de datos : %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error en la conexion con la BD: %w", err)
	}

	db.SetMaxOpenConns(25)
	return db, nil
}

func main() {

	// godotenv.Load()

	database, err := abrirDB()
	if err != nil {
		log.Fatalf("error al abrir la base de datos: %v", err)
	}
	defer database.Close()

	// ctx := context.Background()
	//esta me quedo vieja creo porque los handlers usan el contexto propio de cada request

	queries := sqlc.New(database)

	deportesHandler := handlers.NewDeportesHandler(queries)

	http.HandleFunc("/deportes", deportesHandler.Deportes)
	http.HandleFunc("/deportes/", deportesHandler.Deporte)
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
