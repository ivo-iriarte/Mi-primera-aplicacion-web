package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	sqlc "canchapp/db/sqlc"
	"canchapp/views"
)

type ViewsHandler struct {
	queries *sqlc.Queries
}

func NewViewsHandler(queries *sqlc.Queries) *ViewsHandler {
	return &ViewsHandler{queries: queries}
}

func (vh *ViewsHandler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deportes, err := vh.queries.ListDeportes(ctx)
	if err != nil {
		http.Error(w, "Error al cargar deportes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	views.DeportesPage(deportes).Render(ctx, w)
}

func (vh *ViewsHandler) CrearDeporte(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	nombre := strings.TrimSpace(r.FormValue("nombre"))
	if nombre != "" {
		_, _ = vh.queries.CreateDeporte(r.Context(), nombre)
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (vh *ViewsHandler) CanchasPorDeporte(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	idDeporte, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "ID de deporte inválido", http.StatusBadRequest)
		return
	}

	deporte, err := vh.queries.GetDeporte(ctx, int32(idDeporte))
	if err != nil {
		http.Error(w, "Deporte no encontrado", http.StatusNotFound)
		return
	}

	todasCanchas, err := vh.queries.ListCanchasActivas(ctx)
	if err != nil {
		http.Error(w, "Error al obtener canchas", http.StatusInternalServerError)
		return
	}

	var canchasFiltradas []sqlc.ListCanchasRow
	for _, c := range todasCanchas {
		if c.IDDeporte == int32(idDeporte) {
			canchasFiltradas = append(canchasFiltradas, sqlc.ListCanchasRow(c))
		}
	}

	views.CanchasPage(deporte, canchasFiltradas).Render(ctx, w)
}

func (vh *ViewsHandler) Reservas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var idPreseleccionada int32
	if idParam := r.URL.Query().Get("id_cancha"); idParam != "" {
		if val, err := strconv.Atoi(idParam); err == nil {
			idPreseleccionada = int32(val)
		}
	}

	reservas, _ := vh.queries.ListReservas(ctx)
	canchas, _ := vh.queries.ListCanchas(ctx)
	usuarios, _ := vh.queries.ListUsuarios(ctx)

	if len(usuarios) == 0 {
		u, err := vh.queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
			Nombre:       "Usuario",
			Apellido:     "Demo",
			Email:        "demo@canchapp.com",
			PasswordHash: "secret_hash",
			Telefono:     sql.NullString{String: "2494112233", Valid: true},
		})
		if err == nil {
			usuarios = append(usuarios, u)
		}
	}

	views.ReservasPage(reservas, canchas, usuarios, idPreseleccionada).Render(ctx, w)
}

func (vh *ViewsHandler) CrearReserva(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	idUsuario, _ := strconv.Atoi(r.FormValue("id_usuario"))
	idCancha, _ := strconv.Atoi(r.FormValue("id_cancha"))
	inicioStr := r.FormValue("fecha_inicio")
	finStr := r.FormValue("fecha_fin")
	observaciones := strings.TrimSpace(r.FormValue("observaciones"))

	layout := "2006-01-02T15:04"
	fechaInicio, err1 := time.Parse(layout, inicioStr)
	fechaFin, err2 := time.Parse(layout, finStr)

	if err1 != nil || err2 != nil || !fechaFin.After(fechaInicio) {
		http.Error(w, "Rango de fechas inválido", http.StatusBadRequest)
		return
	}

	cancha, err := vh.queries.GetCancha(ctx, int32(idCancha))
	if err != nil {
		http.Error(w, "Cancha inexistente", http.StatusBadRequest)
		return
	}

	estados, _ := vh.queries.ListEstadosReservas(ctx)
	var estadoID int32 = 1
	if len(estados) > 0 {
		estadoID = estados[0].IDEstadosReserva
	} else {
		nuevoEstado, errEst := vh.queries.CreateEstadoReserva(ctx, "Confirmada")
		if errEst == nil {
			estadoID = nuevoEstado.IDEstadosReserva
		}
	}

	obsNull := sql.NullString{
		String: observaciones,
		Valid:  observaciones != "",
	}

	_, err = vh.queries.CreateReserva(ctx, sqlc.CreateReservaParams{
		IDUsuario:     int32(idUsuario),
		IDCancha:      int32(idCancha),
		FechaInicio:   fechaInicio,
		FechaFin:      fechaFin,
		IDEstado:      estadoID,
		PrecioTotal:   cancha.Precio,
		Observaciones: obsNull,
	})

	if err != nil {
		http.Error(w, "Error al registrar la reserva: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/app/reservas", http.StatusSeeOther)
}

func (vh *ViewsHandler) CancelarReserva(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	idReserva, err := strconv.Atoi(r.FormValue("id_reserva"))
	if err == nil {
		_, _ = vh.queries.DeleteReserva(r.Context(), int32(idReserva))
	}
	http.Redirect(w, r, "/app/reservas", http.StatusSeeOther)
}
