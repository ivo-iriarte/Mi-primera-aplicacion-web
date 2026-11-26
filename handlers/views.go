package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sqlc "canchapp/db/sqlc"
	"canchapp/logic"
	"canchapp/views"
)

type ViewsHandler struct {
	queries         *sqlc.Queries
	reservasService *logic.ReservasService
}

func NewViewsHandler(queries *sqlc.Queries, reservasService *logic.ReservasService) *ViewsHandler {
	return &ViewsHandler{
		queries:         queries,
		reservasService: reservasService,
	}
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// isHTMXPartial devuelve true cuando HTMX pide actualizar un fragmento (no en restauración de historial completo).
func isHTMXPartial(r *http.Request) bool {
	return isHTMX(r) && r.Header.Get("HX-History-Restore-Request") != "true"
}

func (vh *ViewsHandler) asegurarEstadosYUsuarioDemo(ctx context.Context) []sqlc.Usuario {
	estados, _ := vh.queries.ListEstadosReservas(ctx)
	tieneConfirmada := false
	tieneCancelada := false
	for _, e := range estados {
		if e.Nombre == "Confirmada" {
			tieneConfirmada = true
		}
		if e.Nombre == "Cancelada" {
			tieneCancelada = true
		}
	}
	if !tieneConfirmada {
		_, _ = vh.queries.CreateEstadoReserva(ctx, "Confirmada")
	}
	if !tieneCancelada {
		_, _ = vh.queries.CreateEstadoReserva(ctx, "Cancelada")
	}

	usuarios, _ := vh.queries.ListUsuariosActivos(ctx)
	if len(usuarios) == 0 {
		u, err := vh.queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
			Nombre:       "Martín",
			Apellido:     "Santos",
			Email:        "martin@canchapp.com",
			PasswordHash: "demo_hash",
			Telefono:     sql.NullString{String: "2262-451122", Valid: true},
		})
		if err == nil {
			usuarios = append(usuarios, u)
		}
	}
	return usuarios
}

// Home renderiza la pantalla principal con los deportes disponibles (GET /app).
func (vh *ViewsHandler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deportes, err := vh.queries.ListDeportesActivos(ctx)
	if err != nil {
		http.Error(w, "Error al cargar deportes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXPartial(r) {
		views.FlashMessage("", "", true).Render(ctx, w)
		views.DeportesPartial(deportes).Render(ctx, w)
		return
	}
	views.DeportesPage(deportes).Render(ctx, w)
}

// CrearDeporte procesa el alta de un deporte con actualización parcial + OOB reset en HTMX o PRG clásico.
func (vh *ViewsHandler) CrearDeporte(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	nombre := strings.TrimSpace(r.FormValue("nombre"))

	if isHTMX(r) {
		if err := logic.ValidateDeporte(nombre); err != nil {
			deportes, _ := vh.queries.ListDeportesActivos(ctx)
			views.DeportesGrid(deportes).Render(ctx, w)
			views.FlashMessage(err.Error(), "error", true).Render(ctx, w)
			return
		}

		_, err := vh.queries.CreateDeporte(ctx, nombre)
		deportes, _ := vh.queries.ListDeportesActivos(ctx)
		views.DeportesGrid(deportes).Render(ctx, w)

		if err != nil {
			views.FlashMessage("Error al registrar el deporte: "+err.Error(), "error", true).Render(ctx, w)
		} else {
			views.DeporteForm(true).Render(ctx, w)
			views.FlashMessage("Deporte '"+nombre+"' agregado correctamente.", "success", true).Render(ctx, w)
		}
		return
	}

	if logic.ValidateDeporte(nombre) == nil {
		_, _ = vh.queries.CreateDeporte(ctx, nombre)
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

// InstitucionesPorDeporte muestra las instituciones con canchas activas para un deporte (GET /app/deportes/{id}).
func (vh *ViewsHandler) InstitucionesPorDeporte(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := parsePathSegments(r.URL.Path)
	// Soporta /app/deportes/{id} y también enlaces antiguos /app/deportes/{id}/canchas
	if len(parts) < 3 {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	idDeporte, err := parsePositiveInt32(parts[2])
	if err != nil {
		http.Error(w, "ID de deporte inválido", http.StatusBadRequest)
		return
	}

	deporte, err := vh.queries.GetDeporte(ctx, idDeporte)
	if err != nil {
		http.Error(w, "Deporte no encontrado", http.StatusNotFound)
		return
	}

	instituciones, err := vh.queries.ListInstitucionesByDeporte(ctx, idDeporte)
	if err != nil {
		http.Error(w, "Error al cargar instituciones: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXPartial(r) {
		views.FlashMessage("", "", true).Render(ctx, w)
		views.InstitucionesPorDeportePartial(deporte, instituciones).Render(ctx, w)
		return
	}
	views.InstitucionesPorDeportePage(deporte, instituciones).Render(ctx, w)
}

// InstitucionDetalle muestra la institución, sus deportes disponibles y sus canchas filtrables (GET /app/instituciones/{id}).
func (vh *ViewsHandler) InstitucionDetalle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := parsePathSegments(r.URL.Path)
	if len(parts) < 3 {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	idInstitucion, err := parsePositiveInt32(parts[2])
	if err != nil {
		http.Error(w, "ID de institución inválido", http.StatusBadRequest)
		return
	}

	institucion, err := vh.queries.GetInstitucion(ctx, idInstitucion)
	if err != nil {
		http.Error(w, "Institución no encontrada", http.StatusNotFound)
		return
	}

	var idDeporteFiltro int32
	if depStr := r.URL.Query().Get("id_deporte"); depStr != "" {
		if val, errConv := parsePositiveInt32(depStr); errConv == nil {
			idDeporteFiltro = val
		}
	}

	deportesClub, _ := vh.queries.ListDeportesByInstitucion(ctx, idInstitucion)
	mapaDeportes := make(map[int32]string, len(deportesClub))
	for _, d := range deportesClub {
		mapaDeportes[d.IDDeporte] = d.Nombre
	}

	var canchasView []views.CanchaCardView
	if idDeporteFiltro > 0 {
		rows, errList := vh.queries.ListCanchasByInstitucionYDeporte(ctx, sqlc.ListCanchasByInstitucionYDeporteParams{
			IDInstitucion: idInstitucion,
			IDDeporte:     idDeporteFiltro,
		})
		if errList != nil {
			http.Error(w, "Error al cargar canchas", http.StatusInternalServerError)
			return
		}
		for _, c := range rows {
			servicios, _ := vh.queries.ListServiciosByCancha(ctx, c.IDCancha)
			canchasView = append(canchasView, views.CanchaCardView{
				IDCancha:             c.IDCancha,
				Nombre:               c.Nombre,
				Precio:               c.Precio,
				Activa:               c.Activa,
				DuracionTurnoMinutos: c.DuracionTurnoMinutos,
				HoraApertura:         c.HoraApertura,
				HoraCierre:           c.HoraCierre,
				IDInstitucion:        c.IDInstitucion,
				InstitucionNombre:    institucion.Nombre,
				IDDeporte:            c.IDDeporte,
				DeporteNombre:        mapaDeportes[c.IDDeporte],
				Servicios:            servicios,
			})
		}
	} else {
		rows, errList := vh.queries.ListCanchasByInstitucion(ctx, idInstitucion)
		if errList != nil {
			http.Error(w, "Error al cargar canchas", http.StatusInternalServerError)
			return
		}
		for _, c := range rows {
			servicios, _ := vh.queries.ListServiciosByCancha(ctx, c.IDCancha)
			canchasView = append(canchasView, views.CanchaCardView{
				IDCancha:             c.IDCancha,
				Nombre:               c.Nombre,
				Precio:               c.Precio,
				Activa:               c.Activa,
				DuracionTurnoMinutos: c.DuracionTurnoMinutos,
				HoraApertura:         c.HoraApertura,
				HoraCierre:           c.HoraCierre,
				IDInstitucion:        c.IDInstitucion,
				InstitucionNombre:    institucion.Nombre,
				IDDeporte:            c.IDDeporte,
				DeporteNombre:        mapaDeportes[c.IDDeporte],
				Servicios:            servicios,
			})
		}
	}

	if isHTMXPartial(r) {
		if r.Header.Get("HX-Target") == "seccion-canchas-institucion" {
			views.CanchasInstitucionSection(institucion, deportesClub, idDeporteFiltro, canchasView).Render(ctx, w)
			return
		}
		views.FlashMessage("", "", true).Render(ctx, w)
		views.InstitucionDetallePartial(institucion, deportesClub, idDeporteFiltro, canchasView).Render(ctx, w)
		return
	}

	views.InstitucionDetallePage(institucion, deportesClub, idDeporteFiltro, canchasView).Render(ctx, w)
}

func (vh *ViewsHandler) armarDisponibilidadView(ctx context.Context, idCancha int32, fecha time.Time, idUsuarioSel int32) (views.DisponibilidadView, error) {
	cancha, err := vh.queries.GetCancha(ctx, idCancha)
	if err != nil {
		return views.DisponibilidadView{}, err
	}

	institucion, err := vh.queries.GetInstitucion(ctx, cancha.IDInstitucion)
	if err != nil {
		return views.DisponibilidadView{}, err
	}

	deporte, err := vh.queries.GetDeporte(ctx, cancha.IDDeporte)
	if err != nil {
		return views.DisponibilidadView{}, err
	}

	servicios, _ := vh.queries.ListServiciosByCancha(ctx, cancha.IDCancha)
	usuarios := vh.asegurarEstadosYUsuarioDemo(ctx)
	if idUsuarioSel == 0 && len(usuarios) > 0 {
		idUsuarioSel = usuarios[0].IDUsuario
	}

	turnos, err := vh.reservasService.ObtenerTurnosDelDia(ctx, cancha.IDCancha, fecha)
	if err != nil {
		return views.DisponibilidadView{}, err
	}

	hayDisponibles := false
	for _, t := range turnos {
		if t.Disponible {
			hayDisponibles = true
			break
		}
	}

	hoyUTC := time.Now().UTC()
	return views.DisponibilidadView{
		Cancha: views.CanchaCardView{
			IDCancha:             cancha.IDCancha,
			Nombre:               cancha.Nombre,
			Precio:               cancha.Precio,
			Activa:               cancha.Activa,
			DuracionTurnoMinutos: cancha.DuracionTurnoMinutos,
			HoraApertura:         cancha.HoraApertura,
			HoraCierre:           cancha.HoraCierre,
			IDInstitucion:        cancha.IDInstitucion,
			InstitucionNombre:    institucion.Nombre,
			IDDeporte:            cancha.IDDeporte,
			DeporteNombre:        deporte.Nombre,
			Servicios:            servicios,
		},
		FechaSeleccionada:     fecha.Format("2006-01-02"),
		FechaLegible:          views.FormatearFechaLegible(fecha),
		Dias:                  views.ConstruirDiasSelector(hoyUTC, fecha),
		Turnos:                turnos,
		HayTurnosDisponibles:  hayDisponibles,
		Usuarios:              usuarios,
		IDUsuarioSeleccionado: idUsuarioSel,
	}, nil
}

// CanchaDetalle muestra la información de la cancha, selector de fechas y horarios disponibles (GET /app/canchas/{id}).
func (vh *ViewsHandler) CanchaDetalle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := parsePathSegments(r.URL.Path)
	if len(parts) < 3 {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	idCancha, err := parsePositiveInt32(parts[2])
	if err != nil {
		http.Error(w, "ID de cancha inválido", http.StatusBadRequest)
		return
	}

	hoy := time.Now().UTC()
	fecha := time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 0, 0, 0, 0, time.UTC)
	if fechaParam := strings.TrimSpace(r.URL.Query().Get("fecha")); fechaParam != "" {
		if parsed, errParse := time.Parse("2006-01-02", fechaParam); errParse == nil {
			fecha = parsed
		}
	}

	vm, err := vh.armarDisponibilidadView(ctx, idCancha, fecha, 0)
	if err != nil {
		http.Error(w, "Cancha no encontrada: "+err.Error(), http.StatusNotFound)
		return
	}

	if isHTMXPartial(r) {
		views.FlashMessage("", "", true).Render(ctx, w)
		if r.Header.Get("HX-Target") == "seccion-disponibilidad" {
			views.DisponibilidadSection(vm).Render(ctx, w)
			return
		}
		views.CanchaDetallePartial(vm).Render(ctx, w)
		return
	}

	views.CanchaDetallePage(vm).Render(ctx, w)
}

// CrearReserva procesa la confirmación de un turno desde la pantalla de cancha (POST /app/reservas/crear).
func (vh *ViewsHandler) CrearReserva(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	idUsuario, _ := strconv.Atoi(r.FormValue("id_usuario"))
	idCancha, _ := strconv.Atoi(r.FormValue("id_cancha"))
	fechaStr := strings.TrimSpace(r.FormValue("fecha"))
	horaInicioStr := strings.TrimSpace(r.FormValue("hora_inicio"))
	observaciones := strings.TrimSpace(r.FormValue("observaciones"))

	cancha, err := vh.queries.GetCancha(ctx, int32(idCancha))
	if err != nil {
		http.Error(w, "La cancha seleccionada no existe", http.StatusBadRequest)
		return
	}

	vh.asegurarEstadosYUsuarioDemo(ctx)

	hoy := time.Now().UTC()
	fechaDia := time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 0, 0, 0, 0, time.UTC)
	if parsedDia, errDia := time.Parse("2006-01-02", fechaStr); errDia == nil {
		fechaDia = parsedDia
	}

	var fechaInicio, fechaFin time.Time
	var errHorario error

	if horaInicioStr != "" {
		h, m, errParse := logic.ParseHoraDia(horaInicioStr)
		if errParse != nil {
			errHorario = errParse
		} else {
			fechaInicio = time.Date(fechaDia.Year(), fechaDia.Month(), fechaDia.Day(), h, m, 0, 0, time.UTC)
			fechaFin = fechaInicio.Add(time.Duration(cancha.DuracionTurnoMinutos) * time.Minute)
		}
	} else {
		// Respaldo por compatibilidad si se envían fecha_inicio y fecha_fin completos
		fi, err1 := parseFechaHora(r.FormValue("fecha_inicio"))
		ff, err2 := parseFechaHora(r.FormValue("fecha_fin"))
		if err1 != nil || err2 != nil {
			errHorario = logic.ErrRangoFechasInvalido
		} else {
			fechaInicio = fi.UTC()
			fechaFin = ff.UTC()
			fechaDia = time.Date(fechaInicio.Year(), fechaInicio.Month(), fechaInicio.Day(), 0, 0, 0, 0, time.UTC)
		}
	}

	var errReserva error
	if errHorario != nil {
		errReserva = errHorario
	} else {
		_, errReserva = vh.reservasService.CrearReserva(ctx, logic.CrearReservaInput{
			IDUsuario:     int32(idUsuario),
			IDCancha:      int32(idCancha),
			FechaInicio:   fechaInicio,
			FechaFin:      fechaFin,
			Observaciones: observaciones,
		})
	}

	if isHTMX(r) {
		vm, errVM := vh.armarDisponibilidadView(ctx, int32(idCancha), fechaDia, int32(idUsuario))
		if errVM == nil {
			views.DisponibilidadSection(vm).Render(ctx, w)
		}
		if errReserva != nil {
			views.FlashMessage("No se pudo reservar: "+errReserva.Error(), "error", true).Render(ctx, w)
		} else {
			msg := fmt.Sprintf(
				"¡Reserva confirmada en %s para el %s de %s a %s! Ya podés verla en Mis Reservas.",
				cancha.Nombre,
				fechaInicio.Format("02/01/2006"),
				fechaInicio.Format("15:04"),
				fechaFin.Format("15:04"),
			)
			views.FlashMessage(msg, "success", true).Render(ctx, w)
		}
		return
	}

	if errReserva != nil {
		http.Redirect(w, r, fmt.Sprintf("/app/canchas/%d?fecha=%s", idCancha, fechaDia.Format("2006-01-02")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/app/reservas", http.StatusSeeOther)
}

func (vh *ViewsHandler) armarMisReservasView(ctx context.Context, filtro string, idUsuario int32) (views.MisReservasView, error) {
	if filtro != "historial" {
		filtro = "activas"
	}
	usuarios := vh.asegurarEstadosYUsuarioDemo(ctx)

	todas, err := vh.queries.ListReservasDetalle(ctx)
	if err != nil {
		return views.MisReservasView{}, err
	}

	var activas []sqlc.ListReservasDetalleRow
	var historial []sqlc.ListReservasDetalleRow

	for _, r := range todas {
		if idUsuario > 0 && r.IDUsuario != idUsuario {
			continue
		}
		if r.EstadoNombre == "Cancelada" {
			historial = append(historial, r)
		} else {
			activas = append(activas, r)
		}
	}

	seleccionadas := activas
	if filtro == "historial" {
		seleccionadas = historial
	}

	return views.MisReservasView{
		Filtro:        filtro,
		IDUsuario:     idUsuario,
		Usuarios:      usuarios,
		Reservas:      seleccionadas,
		CantActivas:   len(activas),
		CantHistorial: len(historial),
	}, nil
}

// Reservas muestra la pantalla de Mis Reservas con pestañas Activas e Historial (GET /app/reservas).
func (vh *ViewsHandler) Reservas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Si llega ?id_cancha=X desde algún marcador previo, redirige a la vista de disponibilidad de esa cancha.
	if idCanchaParam := r.URL.Query().Get("id_cancha"); idCanchaParam != "" {
		if idCancha, err := parsePositiveInt32(idCanchaParam); err == nil {
			http.Redirect(w, r, fmt.Sprintf("/app/canchas/%d", idCancha), http.StatusSeeOther)
			return
		}
	}

	filtro := r.URL.Query().Get("filtro")
	var idUsuario int32
	if uParam := r.URL.Query().Get("id_usuario"); uParam != "" {
		if val, err := strconv.Atoi(uParam); err == nil && val > 0 {
			idUsuario = int32(val)
		}
	}

	vm, err := vh.armarMisReservasView(ctx, filtro, idUsuario)
	if err != nil {
		http.Error(w, "Error al cargar reservas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXPartial(r) {
		views.FlashMessage("", "", true).Render(ctx, w)
		views.MisReservasSection(vm).Render(ctx, w)
		return
	}

	views.ReservasPage(vm).Render(ctx, w)
}

// CancelarReserva cambia el estado de la reserva a Cancelada y libera el turno (POST /app/reservas/cancelar).
func (vh *ViewsHandler) CancelarReserva(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	idReserva, errID := parsePositiveInt32(r.FormValue("id_reserva"))
	filtro := r.FormValue("filtro")
	idUsuarioInt, _ := strconv.Atoi(r.FormValue("id_usuario"))
	idUsuario := int32(0)
	if idUsuarioInt > 0 {
		idUsuario = int32(idUsuarioInt)
	}

	var errCancel error
	if errID != nil {
		errCancel = errID
	} else {
		_, errCancel = vh.reservasService.CancelarReserva(ctx, idReserva)
	}

	if isHTMX(r) {
		vm, errVM := vh.armarMisReservasView(ctx, filtro, idUsuario)
		if errVM == nil {
			views.MisReservasSection(vm).Render(ctx, w)
		}
		if errCancel != nil {
			views.FlashMessage("Error al cancelar la reserva: "+errCancel.Error(), "error", true).Render(ctx, w)
		} else {
			views.FlashMessage("Reserva cancelada correctamente. El horario volvió a quedar disponible.", "success", true).Render(ctx, w)
		}
		return
	}

	http.Redirect(w, r, "/app/reservas", http.StatusSeeOther)
}

// EliminarReserva mantiene soporte para DELETE /app/reservas/{id} cancelando la reserva para liberar el turno.
func (vh *ViewsHandler) EliminarReserva(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/app/reservas/")
	idReserva, err := parsePositiveInt32(idStr)
	if err != nil {
		http.Error(w, "ID de reserva inválido", http.StatusBadRequest)
		return
	}
	if _, err := vh.reservasService.CancelarReserva(ctx, idReserva); err != nil {
		http.Error(w, "Error al cancelar la reserva: "+err.Error(), http.StatusBadRequest)
		return
	}
	if isHTMX(r) {
		views.FlashMessage("Reserva cancelada correctamente.", "success", true).Render(ctx, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}
