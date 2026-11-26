package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware registra método HTTP, ruta, dirección IP remota y duración de cada petición.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		log.Printf("Request: %s %s %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("Completed: %s %s (%v)", r.Method, r.URL.RequestURI(), time.Since(inicio))
	})
}
