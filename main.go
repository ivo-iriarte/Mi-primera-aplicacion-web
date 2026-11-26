package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sqlc "canchapp/db/sqlc"
	"canchapp/handlers"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func obtenerDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	user := envOrDefault("DB_USER", "prueba")
	password := envOrDefault("DB_PASSWORD", "prueba")
	dbname := envOrDefault("DB_NAME", "CanchAppprueba")
	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "5433")
	sslmode := envOrDefault("DB_SSLMODE", "disable")

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		user, password, dbname, host, port, sslmode)
}

func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func abrirDB() (*sql.DB, error) {
	db, err := sql.Open("pgx", obtenerDSN())
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
	mux := http.NewServeMux()

	// Archivos estáticos (CSS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Endpoints API REST (TP3)
	deportesHandler := handlers.NewDeportesHandler(queries)
	mux.HandleFunc("/deportes", deportesHandler.Deportes)
	mux.HandleFunc("/deportes/", deportesHandler.Deporte)

	// Vistas SSR + HTMX (TP5 / TP6)
	viewsHandler := handlers.NewViewsHandler(queries)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/app/" {
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/app", viewsHandler.Home)
	mux.HandleFunc("/app/deportes", viewsHandler.DeportesCollection)
	mux.HandleFunc("/app/deportes/", viewsHandler.DeportesSubroutes)
	mux.HandleFunc("/app/reservas", viewsHandler.ReservasCollection)
	mux.HandleFunc("/app/reservas/", viewsHandler.ReservasSubroutes)

	addr := ":" + envOrDefault("PORT", "8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("CanchApp corriendo en http://localhost%s (UI: /app)", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error en el servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("error durante el apagado del servidor: %v", err)
	}
}
