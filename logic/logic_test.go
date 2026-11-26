package logic

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlc "canchapp/db/sqlc"
)

func TestValidacionesPuras(t *testing.T) {
	t.Run("ValidateEmail", func(t *testing.T) {
		if err := ValidateEmail("usuario@club.com"); err != nil {
			t.Errorf("email válido rechazado: %v", err)
		}
		if err := ValidateEmail("invalido-sin-arroba"); !errors.Is(err, ErrEmailInvalido) {
			t.Errorf("se esperaba ErrEmailInvalido, se obtuvo: %v", err)
		}
	})

	t.Run("ValidateCancha", func(t *testing.T) {
		if err := ValidateCancha("Cancha 1", "15000.00", 60, "08:00", "23:00", 1, 1); err != nil {
			t.Errorf("cancha válida rechazada: %v", err)
		}
		if err := ValidateCancha("Cancha 1", "-100", 60, "08:00", "23:00", 1, 1); !errors.Is(err, ErrPrecioInvalido) {
			t.Errorf("se esperaba ErrPrecioInvalido, se obtuvo: %v", err)
		}
		if err := ValidateCancha("Cancha 1", "15000", 60, "23:00", "08:00", 1, 1); !errors.Is(err, ErrHorarioInvalido) {
			t.Errorf("se esperaba ErrHorarioInvalido, se obtuvo: %v", err)
		}
	})

	t.Run("ValidateHorarioReserva_y_CalcularPrecioTotal", func(t *testing.T) {
		inicio := time.Date(2026, time.December, 1, 9, 0, 0, 0, time.UTC)
		fin := time.Date(2026, time.December, 1, 11, 0, 0, 0, time.UTC) // 2 turnos de 60 min

		if err := ValidateHorarioReserva(inicio, fin, "08:00:00", "23:00:00", 60); err != nil {
			t.Fatalf("horario válido rechazado: %v", err)
		}

		total, err := CalcularPrecioTotal("15000.00", 60, inicio, fin)
		if err != nil {
			t.Fatalf("error al calcular precio: %v", err)
		}
		if total != "30000.00" {
			t.Errorf("precio total esperado 30000.00, obtenido %s", total)
		}

		// Fuera de horario (07:00 antes de apertura 08:00)
		temprano := time.Date(2026, time.December, 1, 7, 0, 0, 0, time.UTC)
		if err := ValidateHorarioReserva(temprano, inicio, "08:00:00", "23:00:00", 60); !errors.Is(err, ErrFueraDeHorario) {
			t.Errorf("se esperaba ErrFueraDeHorario, se obtuvo: %v", err)
		}

		// Duración no múltiplo (90 min en cancha de turnos de 60 min)
		finNoMultiplo := inicio.Add(90 * time.Minute)
		if err := ValidateHorarioReserva(inicio, finNoMultiplo, "08:00:00", "23:00:00", 60); !errors.Is(err, ErrDuracionTurnoIncorrect) {
			t.Errorf("se esperaba ErrDuracionTurnoIncorrect, se obtuvo: %v", err)
		}
	})

	t.Run("GenerarTurnosDisponibles", func(t *testing.T) {
		fecha := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)
		ocupado := IntervaloHorario{
			Inicio: time.Date(2026, time.December, 1, 9, 0, 0, 0, time.UTC),
			Fin:    time.Date(2026, time.December, 1, 10, 0, 0, 0, time.UTC),
		}

		turnos, err := GenerarTurnosDisponibles(fecha, "08:00:00", "11:00:00", 60, []IntervaloHorario{ocupado})
		if err != nil {
			t.Fatalf("error al generar turnos: %v", err)
		}
		if len(turnos) != 3 {
			t.Fatalf("se esperaban 3 turnos (08-09, 09-10, 10-11), se obtuvieron %d", len(turnos))
		}
		if !turnos[0].Disponible || turnos[1].Disponible || !turnos[2].Disponible {
			t.Errorf("disponibilidad incorrecta: %+v", turnos)
		}
	})
}

// fakeReservasRepo implementa ReservasRepository en memoria para probar el servicio sin PostgreSQL.
type fakeReservasRepo struct {
	canchas  map[int32]sqlc.GetCanchaRow
	usuarios map[int32]sqlc.Usuario
	estados  map[string]sqlc.EstadosReserva
	reservas map[int32]sqlc.Reserva
	nextID   int32
}

func newFakeReservasRepo() *fakeReservasRepo {
	return &fakeReservasRepo{
		canchas: map[int32]sqlc.GetCanchaRow{
			3: {
				IDCancha:             3,
				Nombre:               "Cancha 3",
				Precio:               "18000.00",
				Activa:               true,
				DuracionTurnoMinutos: 60,
				HoraApertura:         "08:00:00",
				HoraCierre:           "12:00:00",
				IDInstitucion:        1,
				IDDeporte:            1,
			},
		},
		usuarios: map[int32]sqlc.Usuario{
			1: {IDUsuario: 1, Nombre: "Ana", Apellido: "López", Email: "ana@test.com", Activo: true},
		},
		estados: map[string]sqlc.EstadosReserva{
			"Confirmada": {IDEstadosReserva: 1, Nombre: "Confirmada"},
			"Cancelada":  {IDEstadosReserva: 2, Nombre: "Cancelada"},
		},
		reservas: make(map[int32]sqlc.Reserva),
		nextID:   1,
	}
}

func (f *fakeReservasRepo) GetCancha(_ context.Context, id int32) (sqlc.GetCanchaRow, error) {
	c, ok := f.canchas[id]
	if !ok {
		return sqlc.GetCanchaRow{}, sql.ErrNoRows
	}
	return c, nil
}

func (f *fakeReservasRepo) GetUsuario(_ context.Context, id int32) (sqlc.Usuario, error) {
	u, ok := f.usuarios[id]
	if !ok {
		return sqlc.Usuario{}, sql.ErrNoRows
	}
	return u, nil
}

func (f *fakeReservasRepo) GetEstadoReservaByNombre(_ context.Context, nombre string) (sqlc.EstadosReserva, error) {
	e, ok := f.estados[nombre]
	if !ok {
		return sqlc.EstadosReserva{}, sql.ErrNoRows
	}
	return e, nil
}

func (f *fakeReservasRepo) ExisteReservaSolapada(_ context.Context, arg sqlc.ExisteReservaSolapadaParams) (bool, error) {
	idCancelada := f.estados["Cancelada"].IDEstadosReserva
	for _, r := range f.reservas {
		if r.IDCancha == arg.IDCancha && r.IDEstado != idCancelada &&
			HaySolapamiento(r.FechaInicio, r.FechaFin, arg.FechaInicio, arg.FechaFin) {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeReservasRepo) ListReservasActivasByCanchaYRango(_ context.Context, arg sqlc.ListReservasActivasByCanchaYRangoParams) ([]sqlc.Reserva, error) {
	idCancelada := f.estados["Cancelada"].IDEstadosReserva
	var res []sqlc.Reserva
	for _, r := range f.reservas {
		if r.IDCancha == arg.IDCancha && r.IDEstado != idCancelada &&
			HaySolapamiento(r.FechaInicio, r.FechaFin, arg.FechaInicio, arg.FechaFin) {
			res = append(res, r)
		}
	}
	return res, nil
}

func (f *fakeReservasRepo) CreateReserva(_ context.Context, arg sqlc.CreateReservaParams) (sqlc.Reserva, error) {
	r := sqlc.Reserva{
		IDReserva:     f.nextID,
		IDUsuario:     arg.IDUsuario,
		IDCancha:      arg.IDCancha,
		FechaInicio:   arg.FechaInicio,
		FechaFin:      arg.FechaFin,
		IDEstado:      arg.IDEstado,
		PrecioTotal:   arg.PrecioTotal,
		Observaciones: arg.Observaciones,
	}
	f.reservas[r.IDReserva] = r
	f.nextID++
	return r, nil
}

func (f *fakeReservasRepo) GetReserva(_ context.Context, id int32) (sqlc.Reserva, error) {
	r, ok := f.reservas[id]
	if !ok {
		return sqlc.Reserva{}, sql.ErrNoRows
	}
	return r, nil
}

func (f *fakeReservasRepo) CancelarReserva(_ context.Context, id int32) (sqlc.Reserva, error) {
	r, ok := f.reservas[id]
	if !ok {
		return sqlc.Reserva{}, sql.ErrNoRows
	}
	r.IDEstado = f.estados["Cancelada"].IDEstadosReserva
	f.reservas[id] = r
	return r, nil
}

func TestReservasService_FlujoDisponibilidadYCancelacion(t *testing.T) {
	repo := newFakeReservasRepo()
	service := NewReservasService(repo)
	ctx := context.Background()

	martes := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)
	inicio := time.Date(2026, time.December, 1, 9, 0, 0, 0, time.UTC)
	fin := time.Date(2026, time.December, 1, 10, 0, 0, 0, time.UTC)

	// 1. Reservar turno de 9:00 a 10:00 en Cancha 3
	reserva, err := service.CrearReserva(ctx, CrearReservaInput{
		IDUsuario:   1,
		IDCancha:    3,
		FechaInicio: inicio,
		FechaFin:    fin,
	})
	if err != nil {
		t.Fatalf("error inesperado al crear reserva: %v", err)
	}
	if reserva.PrecioTotal != "18000.00" {
		t.Errorf("precio total esperado 18000.00, obtenido %s", reserva.PrecioTotal)
	}

	// 2. Intentar reservar el mismo turno debe fallar con ErrTurnoNoDisponible
	_, err = service.CrearReserva(ctx, CrearReservaInput{
		IDUsuario:   1,
		IDCancha:    3,
		FechaInicio: inicio,
		FechaFin:    fin,
	})
	if !errors.Is(err, ErrTurnoNoDisponible) {
		t.Fatalf("se esperaba ErrTurnoNoDisponible, se obtuvo: %v", err)
	}

	// 3. Consultar turnos del día: el de 09:00-10:00 no debe estar disponible
	turnos, err := service.ObtenerTurnosDelDia(ctx, 3, martes)
	if err != nil {
		t.Fatalf("error al obtener turnos del día: %v", err)
	}
	for _, turno := range turnos {
		if turno.HoraInicio == "09:00" && turno.Disponible {
			t.Errorf("el turno 09:00 debería figurar no disponible")
		}
		if turno.HoraInicio == "10:00" && !turno.Disponible {
			t.Errorf("el turno 10:00 debería seguir disponible")
		}
	}

	// 4. Cancelar la reserva y verificar que el turno vuelve a estar disponible
	cancelada, err := service.CancelarReserva(ctx, reserva.IDReserva)
	if err != nil {
		t.Fatalf("error al cancelar reserva: %v", err)
	}
	if cancelada.IDEstado != 2 {
		t.Errorf("se esperaba estado Cancelada (2), se obtuvo %d", cancelada.IDEstado)
	}

	// Cancelar de nuevo debe devolver ErrReservaYaCancelada
	if _, err := service.CancelarReserva(ctx, reserva.IDReserva); !errors.Is(err, ErrReservaYaCancelada) {
		t.Errorf("se esperaba ErrReservaYaCancelada, se obtuvo: %v", err)
	}

	// Ahora el turno de 09:00 vuelve a poder reservarse
	if _, err := service.CrearReserva(ctx, CrearReservaInput{
		IDUsuario:   1,
		IDCancha:    3,
		FechaInicio: inicio,
		FechaFin:    fin,
	}); err != nil {
		t.Fatalf("el turno cancelado debería poder volver a reservarse: %v", err)
	}
}
