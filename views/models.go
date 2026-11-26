package views

import (
	"strings"
	"time"

	sqlc "canchapp/db/sqlc"
	"canchapp/logic"
)

// CanchaCardView agrupa los datos de una cancha junto con su institución, deporte y servicios para las vistas.
type CanchaCardView struct {
	IDCancha             int32
	Nombre               string
	Precio               string
	Activa               bool
	DuracionTurnoMinutos int16
	HoraApertura         string
	HoraCierre           string
	IDInstitucion        int32
	InstitucionNombre    string
	IDDeporte            int32
	DeporteNombre        string
	Servicios            []sqlc.Servicio
}

// DiaSelector representa una tarjeta de fecha rápida en la selección de turnos.
type DiaSelector struct {
	FechaISO     string
	DiaSemana    string
	NumeroDia    string
	MesCorto     string
	Seleccionado bool
}

// DisponibilidadView contiene el estado necesario para renderizar la grilla de turnos y el formulario de reserva.
type DisponibilidadView struct {
	Cancha                CanchaCardView
	FechaSeleccionada     string
	FechaLegible          string
	Dias                  []DiaSelector
	Turnos                []logic.Turno
	HayTurnosDisponibles  bool
	Usuarios              []sqlc.Usuario
	IDUsuarioSeleccionado int32
}

// MisReservasView contiene el estado para la pantalla de Mis Reservas (pestañas Activas e Historial).
type MisReservasView struct {
	Filtro        string
	IDUsuario     int32
	Usuarios      []sqlc.Usuario
	Reservas      []sqlc.ListReservasDetalleRow
	CantActivas   int
	CantHistorial int
}

// IconoDeporte devuelve un ícono representativo según el nombre del deporte.
func IconoDeporte(nombre string) string {
	n := strings.ToLower(strings.TrimSpace(nombre))
	switch {
	case strings.Contains(n, "fútbol") || strings.Contains(n, "futbol"):
		return "⚽"
	case strings.Contains(n, "tenis"):
		return "🎾"
	case strings.Contains(n, "pádel") || strings.Contains(n, "padel"):
		return "🏓"
	case strings.Contains(n, "básquet") || strings.Contains(n, "basquet"):
		return "🏀"
	case strings.Contains(n, "voley") || strings.Contains(n, "vóley"):
		return "🏐"
	case strings.Contains(n, "hockey"):
		return "🏑"
	default:
		return "🏅"
	}
}

// IconoServicio devuelve un ícono simple según el tipo de servicio de la cancha.
func IconoServicio(servicio string) string {
	s := strings.ToLower(strings.TrimSpace(servicio))
	switch {
	case strings.Contains(s, "luz") || strings.Contains(s, "ilumin"):
		return "💡"
	case strings.Contains(s, "ducha"):
		return "🚿"
	case strings.Contains(s, "vestuario"):
		return "👕"
	case strings.Contains(s, "estacionamiento") || strings.Contains(s, "parking"):
		return "🚗"
	case strings.Contains(s, "buffet") || strings.Contains(s, "cantina") || strings.Contains(s, "parrilla"):
		return "🥤"
	case strings.Contains(s, "techad") || strings.Contains(s, "cubiert"):
		return "🏟️"
	default:
		return "✓"
	}
}

// FormatearHoraCorta convierte "08:00:00" en "08:00".
func FormatearHoraCorta(hora string) string {
	hora = strings.TrimSpace(hora)
	if len(hora) >= 5 {
		return hora[:5]
	}
	return hora
}

var diasSemanaCorto = [...]string{"DOM", "LUN", "MAR", "MIÉ", "JUE", "VIE", "SÁB"}
var diasSemanaLargo = [...]string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"}
var mesesCorto = [...]string{"", "ENE", "FEB", "MAR", "ABR", "MAY", "JUN", "JUL", "AGO", "SEP", "OCT", "NOV", "DIC"}

// ConstruirDiasSelector genera una tira de 7 días consecutivos resaltando la fecha seleccionada.
func ConstruirDiasSelector(base time.Time, seleccionada time.Time) []DiaSelector {
	inicioBase := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.UTC)
	selDia := time.Date(seleccionada.Year(), seleccionada.Month(), seleccionada.Day(), 0, 0, 0, 0, time.UTC)

	// Si la fecha elegida cae fuera de los próximos 7 días desde hoy, centramos la tira desde la fecha elegida.
	finBase := inicioBase.AddDate(0, 0, 6)
	if selDia.Before(inicioBase) || selDia.After(finBase) {
		inicioBase = selDia
	}

	dias := make([]DiaSelector, 0, 7)
	for i := 0; i < 7; i++ {
		d := inicioBase.AddDate(0, 0, i)
		dias = append(dias, DiaSelector{
			FechaISO:     d.Format("2006-01-02"),
			DiaSemana:    diasSemanaCorto[d.Weekday()],
			NumeroDia:    d.Format("02"),
			MesCorto:     mesesCorto[d.Month()],
			Seleccionado: d.Equal(selDia),
		})
	}
	return dias
}

// FormatearFechaLegible devuelve por ejemplo "Martes 01/12/2026".
func FormatearFechaLegible(fecha time.Time) string {
	return diasSemanaLargo[fecha.Weekday()] + " " + fecha.Format("02/01/2006")
}
