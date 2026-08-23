package main

import (
	"fmt"
	"net/http"
)

func main() {
	staticDir := "./static"
	fileServer := http.FileServer(http.Dir(staticDir))
	http.Handle("/", fileServer)

	port := ":8080"
	fmt.Printf("Servidor estatico escuchando en http://localhost%s\n", port)
	fmt.Printf("Servidor sirviendo archivos desde: %s\n", staticDir)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("error al iniciar el servidor: %s\n", err)
	}
}
