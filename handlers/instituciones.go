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

type InstitucionesHandler struct {
	queries *sqlc.Queries
}

func NewInstitucionesHandler(queries *sqlc.Queries) *InstitucionesHandler {
	return &InstitucionesHandler{queries: queries}
}

func (h *InstitucionesHandler) Instituciones(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getInstituciones(w, r)
	case http.MethodPost:
		h.crearInstitucion(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *InstitucionesHandler) Institucion(w http.ResponseWriter, r *http.Request) {
	parts := parsePathSegments(r.URL.Path)
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	id, err := parsePositiveInt32(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid institution ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getInstitucion(w, r, id)
	case http.MethodPut:
		h.updateInstitucion(w, r, id)
	case http.MethodDelete:
		h.deleteInstitucion(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *InstitucionesHandler) getInstituciones(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if idDeporteStr := r.URL.Query().Get("id_deporte"); idDeporteStr != "" {
		idDeporte, err := parsePositiveInt32(idDeporteStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id_deporte inválido")
			return
		}
		instituciones, err := h.queries.ListInstitucionesByDeporte(ctx, idDeporte)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error al obtener instituciones por deporte")
			return
		}
		if instituciones == nil {
			instituciones = []sqlc.Institucion{}
		}
		writeJSON(w, http.StatusOK, instituciones)
		return
	}

	var (
		instituciones []sqlc.Institucion
		err           error
	)
	if r.URL.Query().Get("activas") == "true" {
		instituciones, err = h.queries.ListInstitucionesActivas(ctx)
	} else {
		instituciones, err = h.queries.ListInstituciones(ctx)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener instituciones")
		return
	}
	if instituciones == nil {
		instituciones = []sqlc.Institucion{}
	}
	writeJSON(w, http.StatusOK, instituciones)
}

func (h *InstitucionesHandler) crearInstitucion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nombre      string `json:"nombre"`
		Direccion   string `json:"direccion"`
		Telefono    string `json:"telefono"`
		Email       string `json:"email"`
		Descripcion string `json:"descripcion"`
		Ciudad      string `json:"ciudad"`
		Provincia   string `json:"provincia"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateInstitucion(req.Nombre, req.Direccion, req.Telefono, req.Email, req.Ciudad, req.Provincia); err != nil {
		writeDomainError(w, err)
		return
	}

	inst, err := h.queries.CreateInstitucion(r.Context(), sqlc.CreateInstitucionParams{
		Nombre:      strings.TrimSpace(req.Nombre),
		Direccion:   strings.TrimSpace(req.Direccion),
		Telefono:    strings.TrimSpace(req.Telefono),
		Email:       strings.TrimSpace(req.Email),
		Descripcion: strings.TrimSpace(req.Descripcion),
		Ciudad:      strings.TrimSpace(req.Ciudad),
		Provincia:   strings.TrimSpace(req.Provincia),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al crear institución: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, inst)
}

func (h *InstitucionesHandler) getInstitucion(w http.ResponseWriter, r *http.Request, id int32) {
	inst, err := h.queries.GetInstitucion(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró la institución")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inst)
}

func (h *InstitucionesHandler) updateInstitucion(w http.ResponseWriter, r *http.Request, id int32) {
	var req struct {
		Nombre      string `json:"nombre"`
		Direccion   string `json:"direccion"`
		Telefono    string `json:"telefono"`
		Email       string `json:"email"`
		Descripcion string `json:"descripcion"`
		Activo      bool   `json:"activo"`
		Ciudad      string `json:"ciudad"`
		Provincia   string `json:"provincia"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateInstitucion(req.Nombre, req.Direccion, req.Telefono, req.Email, req.Ciudad, req.Provincia); err != nil {
		writeDomainError(w, err)
		return
	}

	inst, err := h.queries.UpdateInstitucion(r.Context(), sqlc.UpdateInstitucionParams{
		IDInstitucion: id,
		Nombre:        strings.TrimSpace(req.Nombre),
		Direccion:     strings.TrimSpace(req.Direccion),
		Telefono:      strings.TrimSpace(req.Telefono),
		Email:         strings.TrimSpace(req.Email),
		Descripcion:   strings.TrimSpace(req.Descripcion),
		Activo:        req.Activo,
		Ciudad:        strings.TrimSpace(req.Ciudad),
		Provincia:     strings.TrimSpace(req.Provincia),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró la institución")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, inst)
}

func (h *InstitucionesHandler) deleteInstitucion(w http.ResponseWriter, r *http.Request, id int32) {
	filas, err := h.queries.DeleteInstitucion(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filas == 0 {
		writeError(w, http.StatusNotFound, "No se encontró la institución")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
