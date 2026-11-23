package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	sqlc "canchapp/db/sqlc"
	"canchapp/handlers"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func abrirDB() (*sql.DB, error) {
	db, err := sql.Open("pgx", "user=prueba password=prueba dbname=CanchAppprueba host=localhost port=5433 sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("error al iniciar la base de datos: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error en la conexion con la BD: %w", err)
	}

	db.SetMaxOpenConns(25)
	return db, nil
}

func main() {
	database, err := abrirDB()
	if err != nil {
		log.Fatalf("error al abrir la base de datos: %v", err)
	}
	defer database.Close()

	queries := sqlc.New(database)

	// Endpoints API REST (TP3)
	deportesHandler := handlers.NewDeportesHandler(queries)
	http.HandleFunc("/deportes", deportesHandler.Deportes)
	http.HandleFunc("/deportes/", deportesHandler.Deporte)

	// Vistas SSR con Templ (TP5)
	viewsHandler := handlers.NewViewsHandler(queries)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})
	http.HandleFunc("/app", viewsHandler.Home)
	http.HandleFunc("/app/deportes/crear", viewsHandler.CrearDeporte)
	http.HandleFunc("/app/deportes/", viewsHandler.CanchasPorDeporte)
	http.HandleFunc("/app/reservas", viewsHandler.Reservas)
	http.HandleFunc("/app/reservas/crear", viewsHandler.CrearReserva)
	http.HandleFunc("/app/reservas/cancelar", viewsHandler.CancelarReserva)

	log.Println("CanchApp corriendo en http://localhost:8080 (UI: /app)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
