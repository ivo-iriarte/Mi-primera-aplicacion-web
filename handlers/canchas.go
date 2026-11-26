package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	sqlc "canchapp/db/sqlc"
	"canchapp/logic"
)

type CanchasHandler struct {
	queries         *sqlc.Queries
	reservasService *logic.ReservasService
}

func NewCanchasHandler(queries *sqlc.Queries, reservasService *logic.ReservasService) *CanchasHandler {
	return &CanchasHandler{
		queries:         queries,
		reservasService: reservasService,
	}
}

func (h *CanchasHandler) Canchas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getCanchas(w, r)
	case http.MethodPost:
		h.crearCancha(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Cancha maneja /canchas/{id}, /canchas/{id}/turnos y /canchas/{id}/servicios[/{id_servicio}].
func (h *CanchasHandler) Cancha(w http.ResponseWriter, r *http.Request) {
	parts := parsePathSegments(r.URL.Path)
	if len(parts) < 2 || len(parts) > 4 {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	idCancha, err := parsePositiveInt32(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cancha ID")
		return
	}

	// Sub-recurso: /canchas/{id}/turnos?fecha=YYYY-MM-DD
	if len(parts) == 3 && parts[2] == "turnos" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.getTurnosCancha(w, r, idCancha)
		return
	}

	// Sub-recurso: /canchas/{id}/servicios
	if len(parts) == 3 && parts[2] == "servicios" {
		switch r.Method {
		case http.MethodGet:
			h.getServiciosDeCancha(w, r, idCancha)
		case http.MethodPost:
			h.addServicioACancha(w, r, idCancha)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Sub-recurso: /canchas/{id}/servicios/{id_servicio}
	if len(parts) == 4 && parts[2] == "servicios" {
		if r.Method != http.MethodDelete {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		idServicio, err := parsePositiveInt32(parts[3])
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid servicio ID")
			return
		}
		h.removeServicioDeCancha(w, r, idCancha, idServicio)
		return
	}

	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "Recurso no encontrado")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getCancha(w, r, idCancha)
	case http.MethodPut:
		h.updateCancha(w, r, idCancha)
	case http.MethodDelete:
		h.deleteCancha(w, r, idCancha)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *CanchasHandler) getCanchas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	instStr := q.Get("id_institucion")
	depStr := q.Get("id_deporte")

	if instStr != "" && depStr != "" {
		idInst, err1 := parsePositiveInt32(instStr)
		idDep, err2 := parsePositiveInt32(depStr)
		if err1 != nil || err2 != nil {
			writeError(w, http.StatusBadRequest, "id_institucion e id_deporte deben ser enteros positivos")
			return
		}
		canchas, err := h.queries.ListCanchasByInstitucionYDeporte(ctx, sqlc.ListCanchasByInstitucionYDeporteParams{
			IDInstitucion: idInst,
			IDDeporte:     idDep,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener canchas")
			return
		}
		if canchas == nil {
			canchas = []sqlc.ListCanchasByInstitucionYDeporteRow{}
		}
		writeJSON(w, http.StatusOK, canchas)
		return
	}

	if instStr != "" {
		idInst, err := parsePositiveInt32(instStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id_institucion inválido")
			return
		}
		canchas, err := h.queries.ListCanchasByInstitucion(ctx, idInst)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener canchas por institución")
			return
		}
		if canchas == nil {
			canchas = []sqlc.ListCanchasByInstitucionRow{}
		}
		writeJSON(w, http.StatusOK, canchas)
		return
	}

	if depStr != "" {
		idDep, err := parsePositiveInt32(depStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id_deporte inválido")
			return
		}
		canchas, err := h.queries.ListCanchasByDeporte(ctx, idDep)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener canchas por deporte")
			return
		}
		if canchas == nil {
			canchas = []sqlc.ListCanchasByDeporteRow{}
		}
		writeJSON(w, http.StatusOK, canchas)
		return
	}

	if q.Get("activas") == "true" {
		canchas, err := h.queries.ListCanchasActivas(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener canchas activas")
			return
		}
		if canchas == nil {
			canchas = []sqlc.ListCanchasActivasRow{}
		}
		writeJSON(w, http.StatusOK, canchas)
		return
	}

	canchas, err := h.queries.ListCanchas(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener canchas")
		return
	}
	if canchas == nil {
		canchas = []sqlc.ListCanchasRow{}
	}
	writeJSON(w, http.StatusOK, canchas)
}

func (h *CanchasHandler) crearCancha(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nombre               string `json:"nombre"`
		Precio               string `json:"precio"`
		DuracionTurnoMinutos int16  `json:"duracion_turno_minutos"`
		HoraApertura         string `json:"hora_apertura"`
		HoraCierre           string `json:"hora_cierre"`
		IDInstitucion        int32  `json:"id_institucion"`
		IDDeporte            int32  `json:"id_deporte"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateCancha(
		req.Nombre,
		req.Precio,
		req.DuracionTurnoMinutos,
		req.HoraApertura,
		req.HoraCierre,
		req.IDInstitucion,
		req.IDDeporte,
	); err != nil {
		writeDomainError(w, err)
		return
	}

	cancha, err := h.queries.CreateCancha(r.Context(), sqlc.CreateCanchaParams{
		Nombre:               strings.TrimSpace(req.Nombre),
		Precio:               strings.TrimSpace(req.Precio),
		DuracionTurnoMinutos: req.DuracionTurnoMinutos,
		HoraApertura:         strings.TrimSpace(req.HoraApertura),
		HoraCierre:           strings.TrimSpace(req.HoraCierre),
		IDInstitucion:        req.IDInstitucion,
		IDDeporte:            req.IDDeporte,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al crear cancha: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cancha)
}

func (h *CanchasHandler) getCancha(w http.ResponseWriter, r *http.Request, id int32) {
	cancha, err := h.queries.GetCancha(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró la cancha")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cancha)
}

func (h *CanchasHandler) updateCancha(w http.ResponseWriter, r *http.Request, id int32) {
	var req struct {
		Nombre               string `json:"nombre"`
		Precio               string `json:"precio"`
		Activa               bool   `json:"activa"`
		DuracionTurnoMinutos int16  `json:"duracion_turno_minutos"`
		HoraApertura         string `json:"hora_apertura"`
		HoraCierre           string `json:"hora_cierre"`
		IDInstitucion        int32  `json:"id_institucion"`
		IDDeporte            int32  `json:"id_deporte"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateCancha(
		req.Nombre,
		req.Precio,
		req.DuracionTurnoMinutos,
		req.HoraApertura,
		req.HoraCierre,
		req.IDInstitucion,
		req.IDDeporte,
	); err != nil {
		writeDomainError(w, err)
		return
	}

	cancha, err := h.queries.UpdateCancha(r.Context(), sqlc.UpdateCanchaParams{
		IDCancha:             id,
		Nombre:               strings.TrimSpace(req.Nombre),
		Precio:               strings.TrimSpace(req.Precio),
		Activa:               req.Activa,
		DuracionTurnoMinutos: req.DuracionTurnoMinutos,
		HoraApertura:         strings.TrimSpace(req.HoraApertura),
		HoraCierre:           strings.TrimSpace(req.HoraCierre),
		IDInstitucion:        req.IDInstitucion,
		IDDeporte:            req.IDDeporte,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró la cancha")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cancha)
}

func (h *CanchasHandler) deleteCancha(w http.ResponseWriter, r *http.Request, id int32) {
	filas, err := h.queries.DeleteCancha(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filas == 0 {
		writeError(w, http.StatusNotFound, "No se encontró la cancha")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CanchasHandler) getTurnosCancha(w http.ResponseWriter, r *http.Request, idCancha int32) {
	fechaStr := strings.TrimSpace(r.URL.Query().Get("fecha"))
	var (
		fecha time.Time
		err   error
	)
	if fechaStr == "" {
		fecha = time.Now().UTC()
	} else {
		fecha, err = time.Parse("2006-01-02", fechaStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "fecha inválida, use formato YYYY-MM-DD")
			return
		}
	}

	turnos, err := h.reservasService.ObtenerTurnosDelDia(r.Context(), idCancha, fecha)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if turnos == nil {
		turnos = []logic.Turno{}
	}
	writeJSON(w, http.StatusOK, turnos)
}

func (h *CanchasHandler) getServiciosDeCancha(w http.ResponseWriter, r *http.Request, idCancha int32) {
	servicios, err := h.queries.ListServiciosByCancha(r.Context(), idCancha)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener servicios de la cancha")
		return
	}
	if servicios == nil {
		servicios = []sqlc.Servicio{}
	}
	writeJSON(w, http.StatusOK, servicios)
}

func (h *CanchasHandler) addServicioACancha(w http.ResponseWriter, r *http.Request, idCancha int32) {
	var req struct {
		IDServicio int32 `json:"id_servicio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.IDServicio <= 0 {
		writeError(w, http.StatusBadRequest, "id_servicio inválido")
		return
	}

	if err := h.queries.AddServicioToCancha(r.Context(), sqlc.AddServicioToCanchaParams{
		IDCancha:   idCancha,
		IDServicio: req.IDServicio,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "Error al asociar servicio a cancha: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int32{
		"id_cancha":   idCancha,
		"id_servicio": req.IDServicio,
	})
}

func (h *CanchasHandler) removeServicioDeCancha(w http.ResponseWriter, r *http.Request, idCancha, idServicio int32) {
	if err := h.queries.RemoveServicioFromCancha(r.Context(), sqlc.RemoveServicioFromCanchaParams{
		IDCancha:   idCancha,
		IDServicio: idServicio,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "Error al desasociar servicio: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
