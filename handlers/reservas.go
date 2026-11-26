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

type ReservasHandler struct {
	queries *sqlc.Queries
	service *logic.ReservasService
}

func NewReservasHandler(queries *sqlc.Queries, service *logic.ReservasService) *ReservasHandler {
	return &ReservasHandler{
		queries: queries,
		service: service,
	}
}

func (h *ReservasHandler) Reservas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getReservas(w, r)
	case http.MethodPost:
		h.crearReserva(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Reserva maneja /reservas/{id} y /reservas/{id}/cancelar.
func (h *ReservasHandler) Reserva(w http.ResponseWriter, r *http.Request) {
	parts := parsePathSegments(r.URL.Path)
	if len(parts) < 2 || len(parts) > 3 {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	id, err := parsePositiveInt32(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	// Acción de negocio: POST /reservas/{id}/cancelar
	if len(parts) == 3 && parts[2] == "cancelar" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.cancelarReserva(w, r, id)
		return
	}

	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "Recurso no encontrado")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getReserva(w, r, id)
	case http.MethodPut:
		h.updateReserva(w, r, id)
	case http.MethodDelete:
		h.deleteReserva(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *ReservasHandler) getReservas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	if idUsuarioStr := q.Get("id_usuario"); idUsuarioStr != "" {
		idUsuario, err := parsePositiveInt32(idUsuarioStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id_usuario inválido")
			return
		}
		reservas, err := h.queries.ListReservasDetalleByUsuario(ctx, idUsuario)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener reservas del usuario")
			return
		}
		if reservas == nil {
			reservas = []sqlc.ListReservasDetalleByUsuarioRow{}
		}
		writeJSON(w, http.StatusOK, reservas)
		return
	}

	if idCanchaStr := q.Get("id_cancha"); idCanchaStr != "" {
		idCancha, err := parsePositiveInt32(idCanchaStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id_cancha inválido")
			return
		}
		reservas, err := h.queries.ListReservasByCancha(ctx, idCancha)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener reservas de la cancha")
			return
		}
		if reservas == nil {
			reservas = []sqlc.Reserva{}
		}
		writeJSON(w, http.StatusOK, reservas)
		return
	}

	if q.Get("detalle") == "true" {
		reservas, err := h.queries.ListReservasDetalle(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener detalle de reservas")
			return
		}
		if reservas == nil {
			reservas = []sqlc.ListReservasDetalleRow{}
		}
		writeJSON(w, http.StatusOK, reservas)
		return
	}

	reservas, err := h.queries.ListReservas(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener reservas")
		return
	}
	if reservas == nil {
		reservas = []sqlc.Reserva{}
	}
	writeJSON(w, http.StatusOK, reservas)
}

func (h *ReservasHandler) crearReserva(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDUsuario     int32  `json:"id_usuario"`
		IDCancha      int32  `json:"id_cancha"`
		FechaInicio   string `json:"fecha_inicio"`
		FechaFin      string `json:"fecha_fin"`
		Observaciones string `json:"observaciones"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	fechaInicio, err1 := parseFechaHora(req.FechaInicio)
	fechaFin, err2 := parseFechaHora(req.FechaFin)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "fecha_inicio y fecha_fin deben tener formato RFC3339 o YYYY-MM-DDTHH:MM")
		return
	}

	reserva, err := h.service.CrearReserva(r.Context(), logic.CrearReservaInput{
		IDUsuario:     req.IDUsuario,
		IDCancha:      req.IDCancha,
		FechaInicio:   fechaInicio,
		FechaFin:      fechaFin,
		Observaciones: req.Observaciones,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, reserva)
}

func (h *ReservasHandler) getReserva(w http.ResponseWriter, r *http.Request, id int32) {
	reserva, err := h.queries.GetReserva(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró la reserva")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, reserva)
}

func (h *ReservasHandler) updateReserva(w http.ResponseWriter, r *http.Request, id int32) {
	var req struct {
		IDUsuario     int32  `json:"id_usuario"`
		IDCancha      int32  `json:"id_cancha"`
		FechaInicio   string `json:"fecha_inicio"`
		FechaFin      string `json:"fecha_fin"`
		IDEstado      int32  `json:"id_estado"`
		PrecioTotal   string `json:"precio_total"`
		Observaciones string `json:"observaciones"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	fechaInicio, err1 := parseFechaHora(req.FechaInicio)
	fechaFin, err2 := parseFechaHora(req.FechaFin)
	if err1 != nil || err2 != nil || !fechaFin.After(fechaInicio) {
		writeDomainError(w, logic.ErrRangoFechasInvalido)
		return
	}
	if req.IDUsuario <= 0 || req.IDCancha <= 0 || req.IDEstado <= 0 {
		writeDomainError(w, logic.ErrIDInvalido)
		return
	}

	obs := strings.TrimSpace(req.Observaciones)
	reserva, err := h.queries.UpdateReserva(r.Context(), sqlc.UpdateReservaParams{
		IDReserva:     id,
		IDUsuario:     req.IDUsuario,
		IDCancha:      req.IDCancha,
		FechaInicio:   fechaInicio,
		FechaFin:      fechaFin,
		IDEstado:      req.IDEstado,
		PrecioTotal:   strings.TrimSpace(req.PrecioTotal),
		Observaciones: sql.NullString{String: obs, Valid: obs != ""},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró la reserva")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, reserva)
}

func (h *ReservasHandler) cancelarReserva(w http.ResponseWriter, r *http.Request, id int32) {
	cancelada, err := h.service.CancelarReserva(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cancelada)
}

func (h *ReservasHandler) deleteReserva(w http.ResponseWriter, r *http.Request, id int32) {
	filas, err := h.queries.DeleteReserva(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filas == 0 {
		writeError(w, http.StatusNotFound, "No se encontró la reserva")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
