package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	sqlc "canchapp/db/sqlc"
	"canchapp/handlers"
	"canchapp/logic"
	"canchapp/middleware"

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
	reservasService := logic.NewReservasService(queries)

	mux := http.NewServeMux()

	// Endpoints API REST (TP3)
	deportesHandler := handlers.NewDeportesHandler(queries)
	mux.HandleFunc("/deportes", deportesHandler.Deportes)
	mux.HandleFunc("/deportes/", deportesHandler.Deporte)

	institucionesHandler := handlers.NewInstitucionesHandler(queries)
	mux.HandleFunc("/instituciones", institucionesHandler.Instituciones)
	mux.HandleFunc("/instituciones/", institucionesHandler.Institucion)

	canchasHandler := handlers.NewCanchasHandler(queries, reservasService)
	mux.HandleFunc("/canchas", canchasHandler.Canchas)
	mux.HandleFunc("/canchas/", canchasHandler.Cancha)

	usuariosHandler := handlers.NewUsuariosHandler(queries)
	mux.HandleFunc("/usuarios", usuariosHandler.Usuarios)
	mux.HandleFunc("/usuarios/", usuariosHandler.Usuario)

	serviciosHandler := handlers.NewServiciosHandler(queries)
	mux.HandleFunc("/servicios", serviciosHandler.Servicios)
	mux.HandleFunc("/servicios/", serviciosHandler.Servicio)

	estadosHandler := handlers.NewEstadosReservasHandler(queries)
	mux.HandleFunc("/estados-reservas", estadosHandler.EstadosReservas)
	mux.HandleFunc("/estados-reservas/", estadosHandler.EstadoReserva)

	reservasHandler := handlers.NewReservasHandler(queries, reservasService)
	mux.HandleFunc("/reservas", reservasHandler.Reservas)
	mux.HandleFunc("/reservas/", reservasHandler.Reserva)

	// Vistas SSR con Templ + HTMX (TP5 y TP6)
	viewsHandler := handlers.NewViewsHandler(queries, reservasService)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/app", viewsHandler.Home)
	mux.HandleFunc("/app/deportes/crear", viewsHandler.CrearDeporte)
	mux.HandleFunc("/app/deportes/", viewsHandler.InstitucionesPorDeporte)
	mux.HandleFunc("/app/instituciones/", viewsHandler.InstitucionDetalle)
	mux.HandleFunc("/app/canchas/", viewsHandler.CanchaDetalle)
	mux.HandleFunc("/app/reservas", viewsHandler.Reservas)
	mux.HandleFunc("/app/reservas/crear", viewsHandler.CrearReserva)
	mux.HandleFunc("/app/reservas/cancelar", viewsHandler.CancelarReserva)
	mux.HandleFunc("/app/reservas/", viewsHandler.EliminarReserva)

	log.Println("CanchApp corriendo en http://localhost:8080 (UI: /app)")
	log.Fatal(http.ListenAndServe(":8080", middleware.LoggingMiddleware(mux)))
}
