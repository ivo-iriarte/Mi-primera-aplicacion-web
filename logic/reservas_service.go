package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sqlc "canchapp/db/sqlc"
)

// ReservasRepository define el contrato mínimo de persistencia que necesita ReservasService.
// *sqlc.Queries implementa esta interfaz de forma implícita, y en pruebas unitarias
// puede sustituirse por un repositorio en memoria sin levantar PostgreSQL.
type ReservasRepository interface {
	GetCancha(ctx context.Context, idCancha int32) (sqlc.GetCanchaRow, error)
	GetUsuario(ctx context.Context, idUsuario int32) (sqlc.Usuario, error)
	GetEstadoReservaByNombre(ctx context.Context, nombre string) (sqlc.EstadosReserva, error)
	ExisteReservaSolapada(ctx context.Context, arg sqlc.ExisteReservaSolapadaParams) (bool, error)
	ListReservasActivasByCanchaYRango(ctx context.Context, arg sqlc.ListReservasActivasByCanchaYRangoParams) ([]sqlc.Reserva, error)
	CreateReserva(ctx context.Context, arg sqlc.CreateReservaParams) (sqlc.Reserva, error)
	GetReserva(ctx context.Context, idReserva int32) (sqlc.Reserva, error)
	CancelarReserva(ctx context.Context, idReserva int32) (sqlc.Reserva, error)
}

// CrearReservaInput contiene los datos de entrada para el caso de uso de creación de reserva.
type CrearReservaInput struct {
	IDUsuario     int32
	IDCancha      int32
	FechaInicio   time.Time
	FechaFin      time.Time
	Observaciones string
}

// ReservasService coordina los casos de uso y reglas de negocio de reservas y disponibilidad.
type ReservasService struct {
	repo ReservasRepository
}

// NewReservasService construye el servicio de reservas inyectando el repositorio.
func NewReservasService(repo ReservasRepository) *ReservasService {
	return &ReservasService{repo: repo}
}

// CrearReserva aplica todas las reglas de negocio antes de persistir una reserva:
// 1. Verifica usuario y cancha existentes y que la cancha esté activa.
// 2. Valida horario de atención y duración múltiplo del turno.
// 3. Verifica que no exista otra reserva activa solapada en esa cancha.
// 4. Calcula el precio total según la cantidad de turnos.
// 5. Asigna el estado "Confirmada" y persiste la reserva.
func (s *ReservasService) CrearReserva(ctx context.Context, input CrearReservaInput) (sqlc.Reserva, error) {
	if input.IDUsuario <= 0 || input.IDCancha <= 0 {
		return sqlc.Reserva{}, ErrIDInvalido
	}

	if _, err := s.repo.GetUsuario(ctx, input.IDUsuario); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.Reserva{}, fmt.Errorf("usuario %d: %w", input.IDUsuario, ErrNoEncontrado)
		}
		return sqlc.Reserva{}, fmt.Errorf("obtener usuario: %w", err)
	}

	cancha, err := s.repo.GetCancha(ctx, input.IDCancha)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.Reserva{}, fmt.Errorf("cancha %d: %w", input.IDCancha, ErrNoEncontrado)
		}
		return sqlc.Reserva{}, fmt.Errorf("obtener cancha: %w", err)
	}

	if !cancha.Activa {
		return sqlc.Reserva{}, ErrCanchaInactiva
	}

	if err := ValidateHorarioReserva(
		input.FechaInicio,
		input.FechaFin,
		cancha.HoraApertura,
		cancha.HoraCierre,
		cancha.DuracionTurnoMinutos,
	); err != nil {
		return sqlc.Reserva{}, err
	}

	solapada, err := s.repo.ExisteReservaSolapada(ctx, sqlc.ExisteReservaSolapadaParams{
		IDCancha:    input.IDCancha,
		FechaInicio: input.FechaInicio,
		FechaFin:    input.FechaFin,
	})
	if err != nil {
		return sqlc.Reserva{}, fmt.Errorf("verificar disponibilidad: %w", err)
	}
	if solapada {
		return sqlc.Reserva{}, ErrTurnoNoDisponible
	}

	precioTotal, err := CalcularPrecioTotal(
		cancha.Precio,
		cancha.DuracionTurnoMinutos,
		input.FechaInicio,
		input.FechaFin,
	)
	if err != nil {
		return sqlc.Reserva{}, err
	}

	estadoConfirmada, err := s.repo.GetEstadoReservaByNombre(ctx, "Confirmada")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.Reserva{}, fmt.Errorf("estado Confirmada: %w", ErrEstadoNoConfigurado)
		}
		return sqlc.Reserva{}, fmt.Errorf("obtener estado Confirmada: %w", err)
	}

	obs := strings.TrimSpace(input.Observaciones)
	reserva, err := s.repo.CreateReserva(ctx, sqlc.CreateReservaParams{
		IDUsuario:     input.IDUsuario,
		IDCancha:      input.IDCancha,
		FechaInicio:   input.FechaInicio,
		FechaFin:      input.FechaFin,
		IDEstado:      estadoConfirmada.IDEstadosReserva,
		PrecioTotal:   precioTotal,
		Observaciones: sql.NullString{String: obs, Valid: obs != ""},
	})
	if err != nil {
		return sqlc.Reserva{}, fmt.Errorf("crear reserva: %w", err)
	}

	return reserva, nil
}

// CancelarReserva cambia el estado de una reserva existente a "Cancelada" para liberar el turno.
func (s *ReservasService) CancelarReserva(ctx context.Context, idReserva int32) (sqlc.Reserva, error) {
	if idReserva <= 0 {
		return sqlc.Reserva{}, ErrIDInvalido
	}

	reserva, err := s.repo.GetReserva(ctx, idReserva)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.Reserva{}, fmt.Errorf("reserva %d: %w", idReserva, ErrNoEncontrado)
		}
		return sqlc.Reserva{}, fmt.Errorf("obtener reserva: %w", err)
	}

	estadoCancelada, err := s.repo.GetEstadoReservaByNombre(ctx, "Cancelada")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.Reserva{}, fmt.Errorf("estado Cancelada: %w", ErrEstadoNoConfigurado)
		}
		return sqlc.Reserva{}, fmt.Errorf("obtener estado Cancelada: %w", err)
	}

	if reserva.IDEstado == estadoCancelada.IDEstadosReserva {
		return sqlc.Reserva{}, ErrReservaYaCancelada
	}

	cancelada, err := s.repo.CancelarReserva(ctx, idReserva)
	if err != nil {
		return sqlc.Reserva{}, fmt.Errorf("cancelar reserva: %w", err)
	}

	return cancelada, nil
}

// ObtenerTurnosDelDia devuelve todos los turnos del día para una cancha indicando
// cuáles siguen disponibles y cuáles están ocupados por reservas activas.
func (s *ReservasService) ObtenerTurnosDelDia(ctx context.Context, idCancha int32, fecha time.Time) ([]Turno, error) {
	if idCancha <= 0 {
		return nil, ErrIDInvalido
	}

	cancha, err := s.repo.GetCancha(ctx, idCancha)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cancha %d: %w", idCancha, ErrNoEncontrado)
		}
		return nil, fmt.Errorf("obtener cancha: %w", err)
	}

	if !cancha.Activa {
		return nil, ErrCanchaInactiva
	}

	loc := fecha.Location()
	inicioDia := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, loc)
	finDia := inicioDia.Add(24 * time.Hour)

	reservas, err := s.repo.ListReservasActivasByCanchaYRango(ctx, sqlc.ListReservasActivasByCanchaYRangoParams{
		IDCancha:    idCancha,
		FechaInicio: inicioDia,
		FechaFin:    finDia,
	})
	if err != nil {
		return nil, fmt.Errorf("listar reservas activas del día: %w", err)
	}

	ocupados := make([]IntervaloHorario, 0, len(reservas))
	for _, r := range reservas {
		ocupados = append(ocupados, IntervaloHorario{
			Inicio: r.FechaInicio,
			Fin:    r.FechaFin,
		})
	}

	return GenerarTurnosDisponibles(fecha, cancha.HoraApertura, cancha.HoraCierre, cancha.DuracionTurnoMinutos, ocupados)
}
