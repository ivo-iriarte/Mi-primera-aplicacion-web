package logic

import "errors"

// Errores de dominio y validación de la capa de lógica de negocio.
// Los handlers los inspeccionan con errors.Is para mapearlos al código HTTP adecuado
// (400 Bad Request, 404 Not Found, 409 Conflict, etc.).
var (
	ErrNombreVacio            = errors.New("el nombre no puede estar vacío")
	ErrCampoObligatorio       = errors.New("faltan campos obligatorios")
	ErrEmailInvalido          = errors.New("formato de email inválido")
	ErrPrecioInvalido         = errors.New("el precio debe ser un número válido mayor o igual a cero")
	ErrDuracionInvalida       = errors.New("la duración del turno debe estar entre 1 y 480 minutos")
	ErrHorarioInvalido        = errors.New("horario inválido: el formato debe ser HH:MM o HH:MM:SS y la hora de cierre debe ser posterior a la de apertura")
	ErrIDInvalido             = errors.New("el identificador debe ser mayor a cero")
	ErrRangoFechasInvalido    = errors.New("rango de fechas inválido: la fecha de fin debe ser posterior a la de inicio")
	ErrFueraDeHorario         = errors.New("el turno solicitado está fuera del horario de atención de la cancha")
	ErrDuracionTurnoIncorrect = errors.New("la duración de la reserva debe ser un múltiplo de la duración de turno de la cancha")
	ErrCanchaInactiva         = errors.New("la cancha no se encuentra activa")
	ErrTurnoNoDisponible      = errors.New("el turno solicitado ya está reservado")
	ErrReservaYaCancelada     = errors.New("la reserva ya se encuentra cancelada")
	ErrEstadoNoConfigurado    = errors.New("el estado de reserva requerido no está configurado en la base de datos")
	ErrNoEncontrado           = errors.New("recurso no encontrado")
)
