package logic

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// IntervaloHorario representa un bloque de tiempo ocupado por una reserva activa.
type IntervaloHorario struct {
	Inicio time.Time
	Fin    time.Time
}

// Turno representa un bloque reservable de una cancha en una fecha determinada.
type Turno struct {
	FechaInicio time.Time `json:"fecha_inicio"`
	FechaFin    time.Time `json:"fecha_fin"`
	HoraInicio  string    `json:"hora_inicio"`
	HoraFin     string    `json:"hora_fin"`
	Disponible  bool      `json:"disponible"`
}

// ValidateEmail valida el formato básico de un correo electrónico.
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(strings.TrimSpace(email)) {
		return ErrEmailInvalido
	}
	return nil
}

// ValidateDeporte valida las reglas de entrada de un deporte.
func ValidateDeporte(nombre string) error {
	if strings.TrimSpace(nombre) == "" {
		return ErrNombreVacio
	}
	return nil
}

// ValidateInstitucion valida los campos obligatorios y formato de email de una institución.
func ValidateInstitucion(nombre, direccion, telefono, email, ciudad, provincia string) error {
	if strings.TrimSpace(nombre) == "" {
		return ErrNombreVacio
	}
	if strings.TrimSpace(direccion) == "" ||
		strings.TrimSpace(telefono) == "" ||
		strings.TrimSpace(ciudad) == "" ||
		strings.TrimSpace(provincia) == "" {
		return ErrCampoObligatorio
	}
	if err := ValidateEmail(email); err != nil {
		return err
	}
	return nil
}

// ValidateUsuario valida los campos obligatorios y el email de un usuario.
func ValidateUsuario(nombre, apellido, email, passwordHash string) error {
	if strings.TrimSpace(nombre) == "" || strings.TrimSpace(apellido) == "" || strings.TrimSpace(passwordHash) == "" {
		return ErrCampoObligatorio
	}
	if err := ValidateEmail(email); err != nil {
		return err
	}
	return nil
}

// ValidateServicio valida los campos obligatorios de un servicio.
func ValidateServicio(servicio string) error {
	if strings.TrimSpace(servicio) == "" {
		return ErrNombreVacio
	}
	return nil
}

// ValidateEstadoReserva valida el nombre de un estado de reserva.
func ValidateEstadoReserva(nombre string) error {
	if strings.TrimSpace(nombre) == "" {
		return ErrNombreVacio
	}
	return nil
}

// ParseHoraDia parsea una hora en formato "15:04:05" o "15:04" y devuelve horas y minutos.
func ParseHoraDia(hora string) (int, int, error) {
	hora = strings.TrimSpace(hora)
	for _, layout := range []string{"15:04:05", "15:04"} {
		if t, err := time.Parse(layout, hora); err == nil {
			return t.Hour(), t.Minute(), nil
		}
	}
	return 0, 0, ErrHorarioInvalido
}

// ValidateCancha valida todos los invariantes de dominio de una cancha.
func ValidateCancha(nombre, precioStr string, duracionMinutos int16, horaApertura, horaCierre string, idInstitucion, idDeporte int32) error {
	if strings.TrimSpace(nombre) == "" {
		return ErrNombreVacio
	}
	precio, err := strconv.ParseFloat(strings.TrimSpace(precioStr), 64)
	if err != nil || math.IsNaN(precio) || math.IsInf(precio, 0) || precio < 0 {
		return ErrPrecioInvalido
	}
	if duracionMinutos <= 0 || duracionMinutos > 480 {
		return ErrDuracionInvalida
	}
	hAp, mAp, err := ParseHoraDia(horaApertura)
	if err != nil {
		return err
	}
	hCi, mCi, err := ParseHoraDia(horaCierre)
	if err != nil {
		return err
	}
	if hCi*60+mCi <= hAp*60+mAp {
		return ErrHorarioInvalido
	}
	if idInstitucion <= 0 || idDeporte <= 0 {
		return ErrIDInvalido
	}
	return nil
}

// ValidateHorarioReserva verifica que el rango de reserva sea válido, caiga dentro del horario
// de apertura y cierre de la cancha en un mismo día y respete la duración de turno configurada.
func ValidateHorarioReserva(fechaInicio, fechaFin time.Time, horaApertura, horaCierre string, duracionMinutos int16) error {
	if !fechaFin.After(fechaInicio) {
		return ErrRangoFechasInvalido
	}
	if duracionMinutos <= 0 {
		return ErrDuracionInvalida
	}

	hAp, mAp, err := ParseHoraDia(horaApertura)
	if err != nil {
		return err
	}
	hCi, mCi, err := ParseHoraDia(horaCierre)
	if err != nil {
		return err
	}

	loc := fechaInicio.Location()
	apertura := time.Date(fechaInicio.Year(), fechaInicio.Month(), fechaInicio.Day(), hAp, mAp, 0, 0, loc)
	cierre := time.Date(fechaInicio.Year(), fechaInicio.Month(), fechaInicio.Day(), hCi, mCi, 0, 0, loc)

	if fechaInicio.Before(apertura) || fechaFin.After(cierre) {
		return ErrFueraDeHorario
	}

	duracionReserva := fechaFin.Sub(fechaInicio)
	duracionTurno := time.Duration(duracionMinutos) * time.Minute
	if duracionReserva < duracionTurno || duracionReserva%duracionTurno != 0 {
		return ErrDuracionTurnoIncorrect
	}

	return nil
}

// CalcularPrecioTotal calcula el precio total de una reserva en función del precio por turno
// de la cancha y la cantidad de turnos que abarca [fechaInicio, fechaFin).
func CalcularPrecioTotal(precioTurnoStr string, duracionTurnoMinutos int16, fechaInicio, fechaFin time.Time) (string, error) {
	if !fechaFin.After(fechaInicio) {
		return "", ErrRangoFechasInvalido
	}
	if duracionTurnoMinutos <= 0 {
		return "", ErrDuracionInvalida
	}
	precioTurno, err := strconv.ParseFloat(strings.TrimSpace(precioTurnoStr), 64)
	if err != nil || math.IsNaN(precioTurno) || math.IsInf(precioTurno, 0) || precioTurno < 0 {
		return "", ErrPrecioInvalido
	}

	minutos := fechaFin.Sub(fechaInicio).Minutes()
	cantidadTurnos := minutos / float64(duracionTurnoMinutos)
	total := math.Round(precioTurno*cantidadTurnos*100) / 100
	return fmt.Sprintf("%.2f", total), nil
}

// HaySolapamiento indica si dos intervalos semiabiertos [inicioA, finA) y [inicioB, finB) se superponen.
func HaySolapamiento(inicioA, finA, inicioB, finB time.Time) bool {
	return inicioA.Before(finB) && finA.After(inicioB)
}

// GenerarTurnosDisponibles construye la grilla de turnos de una cancha para una fecha dada
// y marca como no disponibles aquellos bloques que se solapan con alguna reserva activa.
func GenerarTurnosDisponibles(fecha time.Time, horaApertura, horaCierre string, duracionMinutos int16, reservasActivas []IntervaloHorario) ([]Turno, error) {
	if duracionMinutos <= 0 {
		return nil, ErrDuracionInvalida
	}
	hAp, mAp, err := ParseHoraDia(horaApertura)
	if err != nil {
		return nil, err
	}
	hCi, mCi, err := ParseHoraDia(horaCierre)
	if err != nil {
		return nil, err
	}
	if hCi*60+mCi <= hAp*60+mAp {
		return nil, ErrHorarioInvalido
	}

	loc := fecha.Location()
	inicioDia := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), hAp, mAp, 0, 0, loc)
	cierreDia := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), hCi, mCi, 0, 0, loc)
	paso := time.Duration(duracionMinutos) * time.Minute

	var turnos []Turno
	for actual := inicioDia; !actual.Add(paso).After(cierreDia); actual = actual.Add(paso) {
		finTurno := actual.Add(paso)
		disponible := true
		for _, r := range reservasActivas {
			if HaySolapamiento(actual, finTurno, r.Inicio, r.Fin) {
				disponible = false
				break
			}
		}
		turnos = append(turnos, Turno{
			FechaInicio: actual,
			FechaFin:    finTurno,
			HoraInicio:  actual.Format("15:04"),
			HoraFin:     finTurno.Format("15:04"),
			Disponible:  disponible,
		})
	}

	return turnos, nil
}
