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

type EstadosReservasHandler struct {
	queries *sqlc.Queries
}

func NewEstadosReservasHandler(queries *sqlc.Queries) *EstadosReservasHandler {
	return &EstadosReservasHandler{queries: queries}
}

func (h *EstadosReservasHandler) EstadosReservas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getEstados(w, r)
	case http.MethodPost:
		h.crearEstado(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *EstadosReservasHandler) EstadoReserva(w http.ResponseWriter, r *http.Request) {
	parts := parsePathSegments(r.URL.Path)
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	id, err := parsePositiveInt32(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid state ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getEstado(w, r, id)
	case http.MethodPut:
		h.updateEstado(w, r, id)
	case http.MethodDelete:
		h.deleteEstado(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *EstadosReservasHandler) getEstados(w http.ResponseWriter, r *http.Request) {
	estados, err := h.queries.ListEstadosReservas(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener estados de reserva")
		return
	}
	if estados == nil {
		estados = []sqlc.EstadosReserva{}
	}
	writeJSON(w, http.StatusOK, estados)
}

func (h *EstadosReservasHandler) crearEstado(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nombre string `json:"nombre"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateEstadoReserva(req.Nombre); err != nil {
		writeDomainError(w, err)
		return
	}

	estado, err := h.queries.CreateEstadoReserva(r.Context(), strings.TrimSpace(req.Nombre))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al crear estado de reserva: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, estado)
}

func (h *EstadosReservasHandler) getEstado(w http.ResponseWriter, r *http.Request, id int32) {
	estado, err := h.queries.GetEstadoReserva(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró el estado de reserva")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, estado)
}

func (h *EstadosReservasHandler) updateEstado(w http.ResponseWriter, r *http.Request, id int32) {
	var req struct {
		Nombre string `json:"nombre"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateEstadoReserva(req.Nombre); err != nil {
		writeDomainError(w, err)
		return
	}

	estado, err := h.queries.UpdateEstadoReserva(r.Context(), sqlc.UpdateEstadoReservaParams{
		IDEstadosReserva: id,
		Nombre:           strings.TrimSpace(req.Nombre),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró el estado de reserva")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, estado)
}

func (h *EstadosReservasHandler) deleteEstado(w http.ResponseWriter, r *http.Request, id int32) {
	filas, err := h.queries.DeleteEstadoReserva(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filas == 0 {
		writeError(w, http.StatusNotFound, "No se encontró el estado de reserva")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
