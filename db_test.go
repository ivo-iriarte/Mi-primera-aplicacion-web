package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlc "canchapp/db/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func abrirDBPrueba(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open(
		"pgx",
		"user=prueba password=prueba dbname=CanchAppprueba host=localhost port=5433 sslmode=disable",
	)
	if err != nil {
		t.Fatalf("error al abrir la base de datos: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("error al conectar con la base de datos: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func limpiarTablasPrueba(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		"TRUNCATE TABLE reservas, cancha_servicio, cancha, servicios, estados_reservas, usuario, institucion, deporte RESTART IDENTITY CASCADE",
	)
	if err != nil {
		t.Fatalf("error al limpiar las tablas de prueba: %v", err)
	}
}

func TestQueries_CRUD(t *testing.T) {
	db := abrirDBPrueba(t)
	limpiarTablasPrueba(t, db)

	ctx := context.Background()
	queries := sqlc.New(db)

	t.Run("Institucion_CRUD", func(t *testing.T) {
		inst, err := queries.CreateInstitucion(ctx, sqlc.CreateInstitucionParams{
			Nombre:      "Club Temporal",
			Direccion:   "Av. 2 1234",
			Telefono:    "2262-400000",
			Email:       "temporal@club.com",
			Descripcion: "Descripción inicial",
			Ciudad:      "Necochea",
			Provincia:   "Buenos Aires",
		})
		if err != nil {
			t.Fatalf("error al crear institución: %v", err)
		}
		if inst.IDInstitucion == 0 {
			t.Errorf("la institución creada no recibió ID")
		}

		obtenida, err := queries.GetInstitucion(ctx, inst.IDInstitucion)
		if err != nil {
			t.Fatalf("error al obtener institución: %v", err)
		}
		if obtenida.Nombre != inst.Nombre {
			t.Errorf("nombre esperado %q, obtenido %q", inst.Nombre, obtenida.Nombre)
		}

		actualizada, err := queries.UpdateInstitucion(ctx, sqlc.UpdateInstitucionParams{
			IDInstitucion: inst.IDInstitucion,
			Nombre:        "Club Temporal Actualizado",
			Direccion:     inst.Direccion,
			Telefono:      inst.Telefono,
			Email:         inst.Email,
			Descripcion:   "Descripción actualizada",
			Activo:        true,
			Ciudad:        inst.Ciudad,
			Provincia:     inst.Provincia,
		})
		if err != nil {
			t.Fatalf("error al actualizar institución: %v", err)
		}
		if actualizada.Nombre != "Club Temporal Actualizado" {
			t.Errorf("el nombre de la institución no se actualizó")
		}

		lista, err := queries.ListInstitucionesActivas(ctx)
		if err != nil {
			t.Fatalf("error al listar instituciones activas: %v", err)
		}
		if len(lista) == 0 {
			t.Errorf("se esperaba al menos una institución activa")
		}

		filas, err := queries.DeleteInstitucion(ctx, inst.IDInstitucion)
		if err != nil {
			t.Fatalf("error al eliminar institución: %v", err)
		}
		if filas != 1 {
			t.Errorf("se esperaba eliminar 1 fila, se eliminaron %d", filas)
		}

		_, err = queries.GetInstitucion(ctx, inst.IDInstitucion)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("se esperaba sql.ErrNoRows tras eliminar institución, se obtuvo: %v", err)
		}
	})

	t.Run("Deporte_CRUD", func(t *testing.T) {
		dep, err := queries.CreateDeporte(ctx, "Basquet")
		if err != nil {
			t.Fatalf("error al crear deporte: %v", err)
		}

		obtenido, err := queries.GetDeporteByNombre(ctx, "Basquet")
		if err != nil {
			t.Fatalf("error al obtener deporte por nombre: %v", err)
		}
		if obtenido.IDDeporte != dep.IDDeporte {
			t.Errorf("ID de deporte no coincide")
		}

		actualizado, err := queries.UpdateDeporte(ctx, sqlc.UpdateDeporteParams{
			IDDeporte: dep.IDDeporte,
			Nombre:    "Básquet 5v5",
			Activo:    true,
		})
		if err != nil {
			t.Fatalf("error al actualizar deporte: %v", err)
		}
		if actualizado.Nombre != "Básquet 5v5" {
			t.Errorf("nombre no actualizado")
		}

		activos, err := queries.ListDeportesActivos(ctx)
		if err != nil {
			t.Fatalf("error al listar deportes activos: %v", err)
		}
		if len(activos) == 0 {
			t.Errorf("se esperaba al menos un deporte activo")
		}

		filas, err := queries.DeleteDeporte(ctx, dep.IDDeporte)
		if err != nil {
			t.Fatalf("error al eliminar deporte: %v", err)
		}
		if filas != 1 {
			t.Errorf("se esperaba eliminar 1 fila, se eliminaron %d", filas)
		}

		_, err = queries.GetDeporte(ctx, dep.IDDeporte)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("se esperaba sql.ErrNoRows tras eliminar deporte, se obtuvo: %v", err)
		}
	})

	t.Run("Usuario_CRUD", func(t *testing.T) {
		u, err := queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
			Nombre:       "Juan",
			Apellido:     "Pérez",
			Email:        "juan.perez@test.com",
			PasswordHash: "hash_seguro_123",
			Telefono:     sql.NullString{String: "2262-555555", Valid: true},
		})
		if err != nil {
			t.Fatalf("error al crear usuario: %v", err)
		}

		porEmail, err := queries.GetUsuarioByEmail(ctx, "juan.perez@test.com")
		if err != nil {
			t.Fatalf("error al buscar usuario por email: %v", err)
		}
		if porEmail.IDUsuario != u.IDUsuario {
			t.Errorf("ID de usuario no coincide")
		}

		actualizado, err := queries.UpdateUsuario(ctx, sqlc.UpdateUsuarioParams{
			IDUsuario:    u.IDUsuario,
			Nombre:       "Juan Carlos",
			Apellido:     u.Apellido,
			Email:        u.Email,
			PasswordHash: u.PasswordHash,
			Telefono:     u.Telefono,
			Activo:       true,
		})
		if err != nil {
			t.Fatalf("error al actualizar usuario: %v", err)
		}
		if actualizado.Nombre != "Juan Carlos" {
			t.Errorf("el nombre del usuario no se actualizó")
		}

		filas, err := queries.DeleteUsuario(ctx, u.IDUsuario)
		if err != nil {
			t.Fatalf("error al eliminar usuario: %v", err)
		}
		if filas != 1 {
			t.Errorf("se esperaba eliminar 1 fila, se eliminaron %d", filas)
		}

		_, err = queries.GetUsuario(ctx, u.IDUsuario)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("se esperaba sql.ErrNoRows tras eliminar usuario, se obtuvo: %v", err)
		}
	})

	t.Run("EstadosReservas_CRUD", func(t *testing.T) {
		est, err := queries.CreateEstadoReserva(ctx, "Pendiente")
		if err != nil {
			t.Fatalf("error al crear estado de reserva: %v", err)
		}

		porNombre, err := queries.GetEstadoReservaByNombre(ctx, "Pendiente")
		if err != nil {
			t.Fatalf("error al obtener estado por nombre: %v", err)
		}
		if porNombre.IDEstadosReserva != est.IDEstadosReserva {
			t.Errorf("ID de estado no coincide")
		}

		actualizado, err := queries.UpdateEstadoReserva(ctx, sqlc.UpdateEstadoReservaParams{
			IDEstadosReserva: est.IDEstadosReserva,
			Nombre:           "En espera",
		})
		if err != nil {
			t.Fatalf("error al actualizar estado: %v", err)
		}
		if actualizado.Nombre != "En espera" {
			t.Errorf("el nombre del estado no se actualizó")
		}

		filas, err := queries.DeleteEstadoReserva(ctx, est.IDEstadosReserva)
		if err != nil {
			t.Fatalf("error al eliminar estado: %v", err)
		}
		if filas != 1 {
			t.Errorf("se esperaba eliminar 1 fila, se eliminaron %d", filas)
		}
	})

	t.Run("FlujoCompleto_Canchas_Servicios_y_Reservas", func(t *testing.T) {
		limpiarTablasPrueba(t, db)

		inst, err := queries.CreateInstitucion(ctx, sqlc.CreateInstitucionParams{
			Nombre:      "Club de prueba",
			Direccion:   "Calle 123",
			Telefono:    "2262-123456",
			Email:       "club@prueba.com",
			Descripcion: "Institución utilizada para la prueba",
			Ciudad:      "Necochea",
			Provincia:   "Buenos Aires",
		})
		if err != nil {
			t.Fatalf("error al crear institución: %v", err)
		}

		futbol, err := queries.CreateDeporte(ctx, "Fútbol")
		if err != nil {
			t.Fatalf("error al crear deporte: %v", err)
		}

		padel, err := queries.CreateDeporte(ctx, "Pádel")
		if err != nil {
			t.Fatalf("error al crear segundo deporte: %v", err)
		}

		usuario, err := queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
			Nombre:       "María",
			Apellido:     "Gómez",
			Email:        "maria@prueba.com",
			PasswordHash: "hash_maria",
			Telefono:     sql.NullString{String: "2262-999999", Valid: true},
		})
		if err != nil {
			t.Fatalf("error al crear usuario: %v", err)
		}

		estadoConfirmada, err := queries.CreateEstadoReserva(ctx, "Confirmada")
		if err != nil {
			t.Fatalf("error al crear estado Confirmada: %v", err)
		}

		estadoCancelada, err := queries.CreateEstadoReserva(ctx, "Cancelada")
		if err != nil {
			t.Fatalf("error al crear estado Cancelada: %v", err)
		}

		canchaCreada, err := queries.CreateCancha(ctx, sqlc.CreateCanchaParams{
			Nombre:               "Cancha 3",
			Precio:               "15000.00",
			DuracionTurnoMinutos: 60,
			HoraApertura:         "08:00:00",
			HoraCierre:           "23:00:00",
			IDInstitucion:        inst.IDInstitucion,
			IDDeporte:            futbol.IDDeporte,
		})
		if err != nil {
			t.Fatalf("error al crear la cancha: %v", err)
		}

		// Pantalla 2: Instituciones por deporte
		instFutbol, err := queries.ListInstitucionesByDeporte(ctx, futbol.IDDeporte)
		if err != nil {
			t.Fatalf("error al listar instituciones por deporte: %v", err)
		}
		if len(instFutbol) != 1 || instFutbol[0].IDInstitucion != inst.IDInstitucion {
			t.Errorf("se esperaba encontrar 1 institución para Fútbol, se obtuvieron %d", len(instFutbol))
		}

		instPadel, err := queries.ListInstitucionesByDeporte(ctx, padel.IDDeporte)
		if err != nil {
			t.Fatalf("error al listar instituciones para Pádel: %v", err)
		}
		if len(instPadel) != 0 {
			t.Errorf("no deberían existir instituciones para Pádel aún, se obtuvieron %d", len(instPadel))
		}

		// Pantalla 3: Canchas por institución y deporte
		canchasFiltradas, err := queries.ListCanchasByInstitucionYDeporte(ctx, sqlc.ListCanchasByInstitucionYDeporteParams{
			IDInstitucion: inst.IDInstitucion,
			IDDeporte:     futbol.IDDeporte,
		})
		if err != nil {
			t.Fatalf("error al listar canchas por institución y deporte: %v", err)
		}
		if len(canchasFiltradas) != 1 || canchasFiltradas[0].IDCancha != canchaCreada.IDCancha {
			t.Errorf("la cancha filtrada por institución y deporte no coincide")
		}

		// Servicios y relación N:M cancha_servicio
		servicioLuz, err := queries.CreateServicio(ctx, sqlc.CreateServicioParams{
			Servicio:    "Iluminación LED",
			Descripcion: sql.NullString{String: "Reflectores nocturnos", Valid: true},
		})
		if err != nil {
			t.Fatalf("error al crear servicio: %v", err)
		}

		if err := queries.AddServicioToCancha(ctx, sqlc.AddServicioToCanchaParams{
			IDCancha:   canchaCreada.IDCancha,
			IDServicio: servicioLuz.IDServicio,
		}); err != nil {
			t.Fatalf("error al asociar servicio a cancha: %v", err)
		}

		serviciosCancha, err := queries.ListServiciosByCancha(ctx, canchaCreada.IDCancha)
		if err != nil {
			t.Fatalf("error al listar servicios de la cancha: %v", err)
		}
		if len(serviciosCancha) != 1 || serviciosCancha[0].IDServicio != servicioLuz.IDServicio {
			t.Errorf("el servicio asociado a la cancha no coincide")
		}

		canchasConLuz, err := queries.ListCanchasByServicio(ctx, servicioLuz.IDServicio)
		if err != nil {
			t.Fatalf("error al listar canchas por servicio: %v", err)
		}
		if len(canchasConLuz) != 1 {
			t.Errorf("se esperaba 1 cancha con el servicio, se obtuvieron %d", len(canchasConLuz))
		}

		if err := queries.RemoveServicioFromCancha(ctx, sqlc.RemoveServicioFromCanchaParams{
			IDCancha:   canchaCreada.IDCancha,
			IDServicio: servicioLuz.IDServicio,
		}); err != nil {
			t.Fatalf("error al desasociar servicio de cancha: %v", err)
		}

		// Update precio de cancha
		filasActualizadas, err := queries.UpdatePrecioCancha(ctx, sqlc.UpdatePrecioCanchaParams{
			IDCancha: canchaCreada.IDCancha,
			Precio:   "18000.00",
		})
		if err != nil || filasActualizadas != 1 {
			t.Fatalf("error al actualizar precio de cancha: filas=%d err=%v", filasActualizadas, err)
		}

		// Pantalla 4 y 5: Reservas, solapamiento de turnos y cancelación
		inicio := time.Date(2026, time.December, 1, 9, 0, 0, 0, time.UTC)
		fin := time.Date(2026, time.December, 1, 10, 0, 0, 0, time.UTC)

		// Antes de reservar, el turno debe estar disponible
		solapadaAntes, err := queries.ExisteReservaSolapada(ctx, sqlc.ExisteReservaSolapadaParams{
			IDCancha:    canchaCreada.IDCancha,
			FechaInicio: inicio,
			FechaFin:    fin,
		})
		if err != nil {
			t.Fatalf("error al verificar solapamiento inicial: %v", err)
		}
		if solapadaAntes {
			t.Errorf("el turno debería estar disponible antes de reservar")
		}

		reserva, err := queries.CreateReserva(ctx, sqlc.CreateReservaParams{
			IDUsuario:     usuario.IDUsuario,
			IDCancha:      canchaCreada.IDCancha,
			FechaInicio:   inicio,
			FechaFin:      fin,
			IDEstado:      estadoConfirmada.IDEstadosReserva,
			PrecioTotal:   "18000.00",
			Observaciones: sql.NullString{String: "Traemos pelota", Valid: true},
		})
		if err != nil {
			t.Fatalf("error al crear reserva: %v", err)
		}

		// El mismo turno (09:00-10:00) ahora debe figurar ocupado/solapado
		solapadaMismoTurno, err := queries.ExisteReservaSolapada(ctx, sqlc.ExisteReservaSolapadaParams{
			IDCancha:    canchaCreada.IDCancha,
			FechaInicio: inicio,
			FechaFin:    fin,
		})
		if err != nil {
			t.Fatalf("error al verificar solapamiento del mismo turno: %v", err)
		}
		if !solapadaMismoTurno {
			t.Errorf("el turno reservado de 09:00 a 10:00 debería figurar ocupado")
		}

		// Un turno contiguo (10:00-11:00) debe seguir disponible
		solapadaContiguo, err := queries.ExisteReservaSolapada(ctx, sqlc.ExisteReservaSolapadaParams{
			IDCancha:    canchaCreada.IDCancha,
			FechaInicio: fin,
			FechaFin:    fin.Add(time.Hour),
		})
		if err != nil {
			t.Fatalf("error al verificar turno contiguo: %v", err)
		}
		if solapadaContiguo {
			t.Errorf("el turno contiguo de 10:00 a 11:00 debería estar disponible")
		}

		// Listar turnos ocupados del día
		diaInicio := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)
		diaFin := diaInicio.Add(24 * time.Hour)
		ocupados, err := queries.ListReservasActivasByCanchaYRango(ctx, sqlc.ListReservasActivasByCanchaYRangoParams{
			IDCancha:    canchaCreada.IDCancha,
			FechaInicio: diaInicio,
			FechaFin:    diaFin,
		})
		if err != nil {
			t.Fatalf("error al listar reservas activas por rango: %v", err)
		}
		if len(ocupados) != 1 || ocupados[0].IDReserva != reserva.IDReserva {
			t.Errorf("se esperaba 1 turno ocupado en el día, se obtuvieron %d", len(ocupados))
		}

		// Detalle de "Mis reservas" con joins
		misReservas, err := queries.ListReservasDetalleByUsuario(ctx, usuario.IDUsuario)
		if err != nil {
			t.Fatalf("error al listar detalle de reservas del usuario: %v", err)
		}
		if len(misReservas) != 1 {
			t.Fatalf("se esperaba 1 reserva en el detalle, se obtuvieron %d", len(misReservas))
		}
		if misReservas[0].CanchaNombre != "Cancha 3" || misReservas[0].InstitucionNombre != "Club de prueba" || misReservas[0].DeporteNombre != "Fútbol" || misReservas[0].EstadoNombre != "Confirmada" {
			t.Errorf("los datos del detalle de reserva no coinciden: %+v", misReservas[0])
		}

		// Cancelar reserva: debe pasar a estado 'Cancelada' y liberar el turno
		cancelada, err := queries.CancelarReserva(ctx, reserva.IDReserva)
		if err != nil {
			t.Fatalf("error al cancelar reserva: %v", err)
		}
		if cancelada.IDEstado != estadoCancelada.IDEstadosReserva {
			t.Errorf("se esperaba estado Cancelada (%d), se obtuvo %d", estadoCancelada.IDEstadosReserva, cancelada.IDEstado)
		}

		solapadaDespuesCancelar, err := queries.ExisteReservaSolapada(ctx, sqlc.ExisteReservaSolapadaParams{
			IDCancha:    canchaCreada.IDCancha,
			FechaInicio: inicio,
			FechaFin:    fin,
		})
		if err != nil {
			t.Fatalf("error al verificar disponibilidad tras cancelar: %v", err)
		}
		if solapadaDespuesCancelar {
			t.Errorf("tras cancelar la reserva, el turno de 09:00 a 10:00 debería volver a estar disponible")
		}

		// Eliminar reserva y cancha
		filasElimReserva, err := queries.DeleteReserva(ctx, reserva.IDReserva)
		if err != nil || filasElimReserva != 1 {
			t.Fatalf("error al eliminar reserva: filas=%d err=%v", filasElimReserva, err)
		}

		filasElimCancha, err := queries.DeleteCancha(ctx, canchaCreada.IDCancha)
		if err != nil || filasElimCancha != 1 {
			t.Fatalf("error al eliminar cancha: filas=%d err=%v", filasElimCancha, err)
		}

		_, err = queries.GetCancha(ctx, canchaCreada.IDCancha)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("la cancha no se eliminó correctamente; error obtenido: %v", err)
		}
	})
}
