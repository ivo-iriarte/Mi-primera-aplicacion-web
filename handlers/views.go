package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	sqlc "canchapp/db/sqlc"
	"canchapp/views"

	"github.com/a-h/templ"
)

type ViewsHandler struct {
	queries *sqlc.Queries
}

func NewViewsHandler(queries *sqlc.Queries) *ViewsHandler {
	return &ViewsHandler{queries: queries}
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// isPartialHTMX distingue una petición parcial (ej. hx-get apuntando a #main-content)
// de una navegación completa mejorada con hx-boost sobre <body> o restauración de historial.
func isPartialHTMX(r *http.Request) bool {
	return isHTMX(r) &&
		r.Header.Get("HX-Request-Type") != "full" &&
		r.Header.Get("HX-History-Restore-Request") != "true"
}

func renderComponents(w http.ResponseWriter, r *http.Request, components ...templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx := r.Context()
	for _, c := range components {
		if err := c.Render(ctx, w); err != nil {
			log.Printf("error al renderizar componente templ: %v", err)
			return
		}
	}
}

func (vh *ViewsHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	deportes, err := vh.queries.ListDeportes(ctx)
	if err != nil {
		http.Error(w, "Error al cargar deportes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isPartialHTMX(r) {
		renderComponents(w, r,
			views.DeportesContent(deportes),
			views.FlashMessage("", "", true),
		)
		return
	}

	renderComponents(w, r, views.DeportesPage(deportes))
}

func (vh *ViewsHandler) DeportesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		vh.Home(w, r)
	case http.MethodPost:
		vh.CrearDeporte(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (vh *ViewsHandler) DeportesSubroutes(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Rutas soportadas:
	// - DELETE /app/deportes/{id}
	// - GET    /app/deportes/{id}/canchas
	if len(parts) < 3 {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}

	if parts[2] == "crear" {
		vh.CrearDeporte(w, r)
		return
	}

	idDeporte, err := strconv.Atoi(parts[2])
	if err != nil || idDeporte <= 0 {
		http.Error(w, "ID de deporte inválido", http.StatusBadRequest)
		return
	}

	if len(parts) == 3 {
		if r.Method != http.MethodDelete {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		vh.EliminarDeporte(w, r, idDeporte)
		return
	}

	if len(parts) == 4 && parts[3] == "canchas" {
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		vh.CanchasPorDeporte(w, r, idDeporte)
		return
	}

	http.NotFound(w, r)
}

func (vh *ViewsHandler) CrearDeporte(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	nombre := strings.TrimSpace(r.FormValue("nombre"))

	if isHTMX(r) {
		if nombre == "" {
			deportes, err := vh.queries.ListDeportes(ctx)
			if err != nil {
				http.Error(w, "Error al consultar deportes", http.StatusInternalServerError)
				return
			}
			renderComponents(w, r,
				views.DeportesGrid(deportes),
				views.FlashMessage("El nombre del deporte no puede estar vacío.", "error", true),
			)
			return
		}

		_, errCreate := vh.queries.CreateDeporte(ctx, nombre)
		deportes, errList := vh.queries.ListDeportes(ctx)
		if errList != nil {
			http.Error(w, "Error al consultar deportes", http.StatusInternalServerError)
			return
		}

		if errCreate != nil {
			// En caso de error (ej. nombre duplicado), no reseteamos el formulario para conservar lo ingresado.
			renderComponents(w, r,
				views.DeportesGrid(deportes),
				views.FlashMessage("No se pudo registrar el deporte (verificá que no exista ya).", "error", true),
			)
			return
		}

		// Doble/triple swap: lista actualizada + formulario vacío OOB + mensaje de éxito OOB
		renderComponents(w, r,
			views.DeportesGrid(deportes),
			views.DeporteFormOOB(),
			views.FlashMessage("Deporte '"+nombre+"' agregado correctamente.", "success", true),
		)
		return
	}

	if nombre == "" {
		http.Error(w, "El nombre del deporte es obligatorio", http.StatusBadRequest)
		return
	}
	if _, err := vh.queries.CreateDeporte(ctx, nombre); err != nil {
		http.Error(w, "Error al registrar el deporte: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (vh *ViewsHandler) EliminarDeporte(w http.ResponseWriter, r *http.Request, idDeporte int) {
	ctx := r.Context()
	filas, err := vh.queries.DeleteDeporte(ctx, int32(idDeporte))
	if err != nil {
		if isHTMX(r) {
			w.Header().Set("HX-Reswap", "none")
			renderComponents(w, r,
				views.FlashMessage("No se puede eliminar el deporte porque tiene canchas asociadas.", "error", true),
			)
			return
		}
		http.Error(w, "No se puede eliminar el deporte porque tiene canchas asociadas", http.StatusConflict)
		return
	}

	if filas == 0 {
		if isHTMX(r) {
			w.Header().Set("HX-Reswap", "none")
			renderComponents(w, r,
				views.FlashMessage("El deporte indicado no existe.", "error", true),
			)
			return
		}
		http.Error(w, "Deporte no encontrado", http.StatusNotFound)
		return
	}

	// Respuesta vacía con 200 OK: HTMX reemplaza el elemento objetivo (outerHTML) con nada.
	w.WriteHeader(http.StatusOK)
}

func (vh *ViewsHandler) CanchasPorDeporte(w http.ResponseWriter, r *http.Request, idDeporte int) {
	ctx := r.Context()

	deporte, err := vh.queries.GetDeporte(ctx, int32(idDeporte))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Deporte no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error al obtener el deporte", http.StatusInternalServerError)
		return
	}

	canchas, err := vh.queries.ListCanchasActivasByDeporte(ctx, int32(idDeporte))
	if err != nil {
		http.Error(w, "Error al obtener canchas", http.StatusInternalServerError)
		return
	}

	if isPartialHTMX(r) {
		renderComponents(w, r,
			views.CanchasPartial(deporte, canchas),
			views.FlashMessage("", "", true),
		)
		return
	}

	renderComponents(w, r, views.CanchasPage(deporte, canchas))
}

func (vh *ViewsHandler) ReservasCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		vh.Reservas(w, r)
	case http.MethodPost:
		vh.CrearReserva(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (vh *ViewsHandler) ReservasSubroutes(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}

	if parts[2] == "crear" {
		vh.CrearReserva(w, r)
		return
	}

	idReserva, err := strconv.Atoi(parts[2])
	if err != nil || idReserva <= 0 {
		http.Error(w, "ID de reserva inválido", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	vh.EliminarReserva(w, r, idReserva)
}

func (vh *ViewsHandler) cargarDatosFormularioReserva(r *http.Request) ([]sqlc.ListCanchasRow, []sqlc.Usuario, error) {
	ctx := r.Context()
	canchas, err := vh.queries.ListCanchas(ctx)
	if err != nil {
		return nil, nil, err
	}

	usuarios, err := vh.queries.ListUsuarios(ctx)
	if err != nil {
		return nil, nil, err
	}

	if len(usuarios) == 0 {
		u, errCreate := vh.queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
			Nombre:       "Usuario",
			Apellido:     "Demo",
			Email:        "demo@canchapp.com",
			PasswordHash: "secret_hash",
			Telefono:     sql.NullString{String: "2494112233", Valid: true},
		})
		if errCreate == nil {
			usuarios = append(usuarios, u)
		}
	}

	return canchas, usuarios, nil
}

func (vh *ViewsHandler) Reservas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	var idPreseleccionada int32
	if idParam := r.URL.Query().Get("id_cancha"); idParam != "" {
		if val, err := strconv.Atoi(idParam); err == nil && val > 0 {
			idPreseleccionada = int32(val)
		}
	}

	reservas, err := vh.queries.ListReservas(ctx)
	if err != nil {
		http.Error(w, "Error al cargar reservas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	canchas, usuarios, err := vh.cargarDatosFormularioReserva(r)
	if err != nil {
		http.Error(w, "Error al cargar datos de reserva: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isPartialHTMX(r) {
		renderComponents(w, r,
			views.ReservasContent(reservas, canchas, usuarios, idPreseleccionada),
			views.FlashMessage("", "", true),
		)
		return
	}

	renderComponents(w, r, views.ReservasPage(reservas, canchas, usuarios, idPreseleccionada))
}

func (vh *ViewsHandler) CrearReserva(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	idUsuario, errU := strconv.Atoi(r.FormValue("id_usuario"))
	idCancha, errC := strconv.Atoi(r.FormValue("id_cancha"))
	inicioStr := r.FormValue("fecha_inicio")
	finStr := r.FormValue("fecha_fin")
	observaciones := strings.TrimSpace(r.FormValue("observaciones"))

	if errU != nil || idUsuario <= 0 || errC != nil || idCancha <= 0 {
		if isHTMX(r) {
			reservas, err := vh.queries.ListReservas(ctx)
			if err != nil {
				http.Error(w, "Error al consultar reservas", http.StatusInternalServerError)
				return
			}
			renderComponents(w, r,
				views.ReservasTabla(reservas),
				views.FlashMessage("Seleccioná un usuario y una cancha válidos.", "error", true),
			)
			return
		}
		http.Error(w, "Usuario o cancha inválidos", http.StatusBadRequest)
		return
	}

	layout := "2006-01-02T15:04"
	fechaInicio, err1 := time.Parse(layout, inicioStr)
	fechaFin, err2 := time.Parse(layout, finStr)

	if err1 != nil || err2 != nil || !fechaFin.After(fechaInicio) {
		if isHTMX(r) {
			reservas, err := vh.queries.ListReservas(ctx)
			if err != nil {
				http.Error(w, "Error al consultar reservas", http.StatusInternalServerError)
				return
			}
			renderComponents(w, r,
				views.ReservasTabla(reservas),
				views.FlashMessage("Rango de fechas inválido: la fecha de fin debe ser posterior al inicio.", "error", true),
			)
			return
		}
		http.Error(w, "Rango de fechas inválido", http.StatusBadRequest)
		return
	}

	cancha, err := vh.queries.GetCancha(ctx, int32(idCancha))
	if err != nil {
		if isHTMX(r) {
			reservas, errList := vh.queries.ListReservas(ctx)
			if errList != nil {
				http.Error(w, "Error al consultar reservas", http.StatusInternalServerError)
				return
			}
			renderComponents(w, r,
				views.ReservasTabla(reservas),
				views.FlashMessage("La cancha seleccionada no existe.", "error", true),
			)
			return
		}
		http.Error(w, "Cancha inexistente", http.StatusBadRequest)
		return
	}

	estados, err := vh.queries.ListEstadosReservas(ctx)
	if err != nil {
		http.Error(w, "Error al consultar estados de reserva", http.StatusInternalServerError)
		return
	}
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

	_, errCreate := vh.queries.CreateReserva(ctx, sqlc.CreateReservaParams{
		IDUsuario:     int32(idUsuario),
		IDCancha:      int32(idCancha),
		FechaInicio:   fechaInicio,
		FechaFin:      fechaFin,
		IDEstado:      estadoID,
		PrecioTotal:   cancha.Precio,
		Observaciones: obsNull,
	})

	if isHTMX(r) {
		reservas, errList := vh.queries.ListReservas(ctx)
		if errList != nil {
			http.Error(w, "Error al consultar reservas", http.StatusInternalServerError)
			return
		}
		if errCreate != nil {
			renderComponents(w, r,
				views.ReservasTabla(reservas),
				views.FlashMessage("Error al registrar la reserva: "+errCreate.Error(), "error", true),
			)
			return
		}

		canchas, usuarios, errForm := vh.cargarDatosFormularioReserva(r)
		if errForm != nil {
			renderComponents(w, r,
				views.ReservasTabla(reservas),
				views.FlashMessage("Turno reservado con éxito para "+cancha.Nombre+".", "success", true),
			)
			return
		}

		// Lista actualizada + formulario limpio OOB + notificación de éxito OOB
		renderComponents(w, r,
			views.ReservasTabla(reservas),
			views.ReservaFormOOB(canchas, usuarios, 0),
			views.FlashMessage("Turno reservado con éxito para "+cancha.Nombre+".", "success", true),
		)
		return
	}

	if errCreate != nil {
		http.Error(w, "Error al registrar la reserva: "+errCreate.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/app/reservas", http.StatusSeeOther)
}

func (vh *ViewsHandler) EliminarReserva(w http.ResponseWriter, r *http.Request, idReserva int) {
	ctx := r.Context()
	filas, err := vh.queries.DeleteReserva(ctx, int32(idReserva))
	if err != nil {
		if isHTMX(r) {
			w.Header().Set("HX-Reswap", "none")
			renderComponents(w, r,
				views.FlashMessage("Error al cancelar la reserva: "+err.Error(), "error", true),
			)
			return
		}
		http.Error(w, "Error al cancelar la reserva", http.StatusInternalServerError)
		return
	}

	if filas == 0 {
		if isHTMX(r) {
			w.Header().Set("HX-Reswap", "none")
			renderComponents(w, r,
				views.FlashMessage("La reserva indicada no existe.", "error", true),
			)
			return
		}
		http.Error(w, "Reserva no encontrada", http.StatusNotFound)
		return
	}

	// Respuesta vacía con código 200 OK: HTMX reemplaza la fila (closest tr, outerHTML) con nada.
	w.WriteHeader(http.StatusOK)
}
