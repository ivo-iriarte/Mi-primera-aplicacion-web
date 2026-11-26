package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"canchapp/logic"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parsePathSegments divide la ruta eliminando barras iniciales y finales.
func parsePathSegments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func parsePositiveInt32(raw string) (int32, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, logic.ErrIDInvalido
	}
	return int32(id), nil
}

// parseFechaHora acepta formatos ISO 8601 / RFC3339 y "2006-01-02T15:04".
func parseFechaHora(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, logic.ErrRangoFechasInvalido
}

// writeDomainError traduce errores de la capa de lógica/datos al código HTTP correspondiente.
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, logic.ErrNoEncontrado), errors.Is(err, sql.ErrNoRows):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, logic.ErrTurnoNoDisponible), errors.Is(err, logic.ErrReservaYaCancelada):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, logic.ErrNombreVacio),
		errors.Is(err, logic.ErrCampoObligatorio),
		errors.Is(err, logic.ErrEmailInvalido),
		errors.Is(err, logic.ErrPrecioInvalido),
		errors.Is(err, logic.ErrDuracionInvalida),
		errors.Is(err, logic.ErrHorarioInvalido),
		errors.Is(err, logic.ErrIDInvalido),
		errors.Is(err, logic.ErrRangoFechasInvalido),
		errors.Is(err, logic.ErrFueraDeHorario),
		errors.Is(err, logic.ErrDuracionTurnoIncorrect),
		errors.Is(err, logic.ErrCanchaInactiva):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
