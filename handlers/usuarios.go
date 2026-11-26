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

type UsuariosHandler struct {
	queries *sqlc.Queries
}

func NewUsuariosHandler(queries *sqlc.Queries) *UsuariosHandler {
	return &UsuariosHandler{queries: queries}
}

func (h *UsuariosHandler) Usuarios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getUsuarios(w, r)
	case http.MethodPost:
		h.crearUsuario(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *UsuariosHandler) Usuario(w http.ResponseWriter, r *http.Request) {
	parts := parsePathSegments(r.URL.Path)
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	id, err := parsePositiveInt32(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getUsuario(w, r, id)
	case http.MethodPut:
		h.updateUsuario(w, r, id)
	case http.MethodDelete:
		h.deleteUsuario(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *UsuariosHandler) getUsuarios(w http.ResponseWriter, r *http.Request) {
	var (
		usuarios []sqlc.Usuario
		err      error
	)
	if r.URL.Query().Get("activos") == "true" {
		usuarios, err = h.queries.ListUsuariosActivos(r.Context())
	} else {
		usuarios, err = h.queries.ListUsuarios(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener usuarios")
		return
	}
	if usuarios == nil {
		usuarios = []sqlc.Usuario{}
	}
	writeJSON(w, http.StatusOK, usuarios)
}

func (h *UsuariosHandler) crearUsuario(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nombre       string `json:"nombre"`
		Apellido     string `json:"apellido"`
		Email        string `json:"email"`
		PasswordHash string `json:"password_hash"`
		Telefono     string `json:"telefono"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateUsuario(req.Nombre, req.Apellido, req.Email, req.PasswordHash); err != nil {
		writeDomainError(w, err)
		return
	}

	tel := strings.TrimSpace(req.Telefono)
	usuario, err := h.queries.CreateUsuario(r.Context(), sqlc.CreateUsuarioParams{
		Nombre:       strings.TrimSpace(req.Nombre),
		Apellido:     strings.TrimSpace(req.Apellido),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: strings.TrimSpace(req.PasswordHash),
		Telefono:     sql.NullString{String: tel, Valid: tel != ""},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al crear usuario: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, usuario)
}

func (h *UsuariosHandler) getUsuario(w http.ResponseWriter, r *http.Request, id int32) {
	usuario, err := h.queries.GetUsuario(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró el usuario")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, usuario)
}

func (h *UsuariosHandler) updateUsuario(w http.ResponseWriter, r *http.Request, id int32) {
	var req struct {
		Nombre       string `json:"nombre"`
		Apellido     string `json:"apellido"`
		Email        string `json:"email"`
		PasswordHash string `json:"password_hash"`
		Telefono     string `json:"telefono"`
		Activo       bool   `json:"activo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := logic.ValidateUsuario(req.Nombre, req.Apellido, req.Email, req.PasswordHash); err != nil {
		writeDomainError(w, err)
		return
	}

	tel := strings.TrimSpace(req.Telefono)
	usuario, err := h.queries.UpdateUsuario(r.Context(), sqlc.UpdateUsuarioParams{
		IDUsuario:    id,
		Nombre:       strings.TrimSpace(req.Nombre),
		Apellido:     strings.TrimSpace(req.Apellido),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: strings.TrimSpace(req.PasswordHash),
		Telefono:     sql.NullString{String: tel, Valid: tel != ""},
		Activo:       req.Activo,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No se encontró el usuario")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, usuario)
}

func (h *UsuariosHandler) deleteUsuario(w http.ResponseWriter, r *http.Request, id int32) {
	filas, err := h.queries.DeleteUsuario(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filas == 0 {
		writeError(w, http.StatusNotFound, "No se encontró el usuario")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
