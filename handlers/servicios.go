package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	sqlc "canchapp/db/sqlc"
	"canchapp/logic"
)

type ServiciosHandler struct {
	queries *sqlc.Queries
}

func NewServiciosHandler(queries *sqlc.Queries) *ServiciosHandler {
	return &ServiciosHandler{queries: queries}
}

func (h *ServiciosHandler) Servicios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getServicios(w, r)
	case http.MethodPost:
		h.crearServicio(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *ServiciosHandler) Servicio(w http.ResponseWriter, r *http.Request) {
	parts := parsePathSegments(r.URL.Path)
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	id, err := parsePositiveInt32(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid service ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getServicio(w, r, id)
	case http.MethodPut:
		h.updateServicio(w, r, id)
	case http.MethodDelete:
		h.deleteServicio(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *ServiciosHandler) getServicios(w http.ResponseWriter, r *http.Request) {
	var (
		servicios []sqlc.Servicio
		err       error
	)
	if r.URL.Query().Get("activos") == "true" {
		servicios, err = h.queries.ListServiciosActivos(r.Context())
	} else {
		servicios, err = h.queries.ListServicios(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener servicios")
		return
	}
	if servicios == nil {
		servicios = []sqlc.Servicio{}
	}
	writeJSON(w, http.StatusOK, servicios)
}

func (h *ServiciosHandler) crearServicio(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Servicio    string `json:"servicio"`
		Descripcion string `json:"descripcion"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateServicio(req.Servicio); err != nil {
		writeDomainError(w, err)
		return
	}

	desc := strings.TrimSpace(req.Descripcion)
	servicio, err := h.queries.CreateServicio(r.Context(), sqlc.CreateServicioParams{
		Servicio:    strings.TrimSpace(req.Servicio),
		Descripcion: sql.NullString{String: desc, Valid: desc != ""},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al crear servicio: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, servicio)
}

func (h *ServiciosHandler) getServicio(w http.ResponseWriter, r *http.Request, id int32) {
	servicio, err := h.queries.GetServicio(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró el servicio")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, servicio)
}

func (h *ServiciosHandler) updateServicio(w http.ResponseWriter, r *http.Request, id int32) {
	var req struct {
		Servicio    string `json:"servicio"`
		Descripcion string `json:"descripcion"`
		Activo      bool   `json:"activo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateServicio(req.Servicio); err != nil {
		writeDomainError(w, err)
		return
	}

	desc := strings.TrimSpace(req.Descripcion)
	servicio, err := h.queries.UpdateServicio(r.Context(), sqlc.UpdateServicioParams{
		IDServicio:  id,
		Servicio:    strings.TrimSpace(req.Servicio),
		Descripcion: sql.NullString{String: desc, Valid: desc != ""},
		Activo:      req.Activo,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró el servicio")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, servicio)
}

func (h *ServiciosHandler) deleteServicio(w http.ResponseWriter, r *http.Request, id int32) {
	filas, err := h.queries.DeleteServicio(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filas == 0 {
		writeError(w, http.StatusNotFound, "No se encontró el servicio")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
