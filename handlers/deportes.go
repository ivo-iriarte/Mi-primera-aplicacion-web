package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	sqlc "canchapp/db/sqlc"
)

type DeportesHandler struct {
	queries *sqlc.Queries
}

func NewDeportesHandler(queries *sqlc.Queries) *DeportesHandler {
	return &DeportesHandler{queries: queries}
}

//obviamente se tuvo que hacer una adaptacion del ejemplo del profe para que funcione
//con la BD en postgrees pero mantuvimos la logica de una funcion manejadora, para deportes y otra
// funcion manejadora para deporte por ID

func (h *DeportesHandler) Deportes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getDeportes(w, r)
	case http.MethodPost:
		h.crearDeporte(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DeportesHandler) Deporte(w http.ResponseWriter, r *http.Request) {
	// el id se nos comento que se puede extraer en esta version de GO mediante r.PathValue("id"), igualmente lo hicimos como el profe con split

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[2] == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid sport ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getDeporte(w, r, id)
	case http.MethodPut:
		h.updateDeporte(w, r, id)
	case http.MethodDelete:
		h.deleteDeporte(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

//-----------------------------------obtener todos los deportes-------------------------------------------------

func (h *DeportesHandler) getDeportes(w http.ResponseWriter, r *http.Request) {

	deportes, err := h.queries.ListDeportes(r.Context())
	if err != nil {
		http.Error(w, "Error al obtener deportes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deportes)
}

//--------------------------------crear un deporte----------------------------------------------

func (h *DeportesHandler) crearDeporte(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nombre string `json:"nombre"`
	}

	//decodificamos el json como hace el profe con el ejemplo de producto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// nos fijamos que el nombre no sea vacio (quiza deberiamos chequear que sea un deporte de verdad cargando un catalogo de deportes o asi, o quiza en restricciones de dominio de la BD)
	if req.Nombre == "" {
		http.Error(w, "El nombre esta vacio", http.StatusBadRequest)
		return
	}
	// si pasa los errores llamamos a sqlc
	deporte, err := h.queries.CreateDeporte(r.Context(), req.Nombre)
	if err != nil {
		http.Error(w, "Error al crear deporte: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// devolvemos el 201 osea el status created con el objeto en JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deporte)
}

//----------------------obtener un solo deporte por su id----------------------------------------------------

//lo hacemos con ID para resolver por el momento, pero la logica en realidad quiza seria que se busque por nombre
//el cual definimos con unique, ya que tiene mas sentido conocer el nombre del deporte (el cual es unico), y no el id
// luego se reveeera

func (h *DeportesHandler) getDeporte(w http.ResponseWriter, r *http.Request, id int) {
	deporte, err := h.queries.GetDeporte(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "No se encontro el deporte", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deporte)
}

func (h *DeportesHandler) updateDeporte(w http.ResponseWriter, r *http.Request, id int) {
	var req struct {
		Nombre string `json:"nombre"`
		Activo bool   `json:"activo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Nombre == "" {
		http.Error(w, "El deporte no tiene nombre", http.StatusBadRequest)
		return
	}

	//--------------------------------------actualizar un deporte ---------------------------------------

	//como nota recordar que UpdateDeportParams es el tipo que genero el SQLC y bueno sus argumentos que con estas pavadas nos trabamos un par

	//esto de la linea 147 a la 155 es el equivalente a la modificacion que hace el alejandro en la filmina con la diferencia de que usamos la conexion con la BD y el usaba un slice
	//osea aca manejamos una conexion real por lo tanto usamos las funciones de SQLC

	//no sabemos si esta bien que el error que tire sea el 500 y no el 404 pero lo dehjamos asi como lo hizo el profe

	deporte, err := h.queries.UpdateDeporte(r.Context(), sqlc.UpdateDeporteParams{
		IDDeporte: int32(id),
		Nombre:    req.Nombre,
		Activo:    req.Activo,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deporte)
}

// ---------------------------------------Eliminar un deporte-------------------------------------------------
func (h *DeportesHandler) deleteDeporte(w http.ResponseWriter, r *http.Request, id int) {
	filas, err := h.queries.DeleteDeporte(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if filas == 0 {
		http.Error(w, "No se elimino el deporte", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
