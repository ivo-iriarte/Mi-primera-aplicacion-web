package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	sqlc "tp2web/db/sqlc"
)

func TestQueries_CRUD(t *testing.T) {
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

	ctx := context.Background()

	_, err = db.ExecContext(
		ctx,
		"TRUNCATE TABLE cancha, institucion, deporte RESTART IDENTITY CASCADE",
	)
	if err != nil {
		t.Fatalf("error al limpiar las tablas de prueba: %v", err)
	}

	queries := sqlc.New(db)

	institucion, err := queries.CreateInstitucion(
		ctx,
		sqlc.CreateInstitucionParams{
			Nombre:      "Club de prueba",
			Direccion:   "Calle 123",
			Telefono:    "2262-123456",
			Email:       "club@prueba.com",
			Descripcion: "Institución utilizada para la prueba",
			Ciudad:      "Necochea",
			Provincia:   "Buenos Aires",
		},
	)
	if err != nil {
		t.Fatalf("error al crear la institución: %v", err)
	}

	if institucion.IDInstitucion == 0 {
		t.Errorf("la institución creada no recibió un ID")
	}

	deporte, err := queries.CreateDeporte(ctx, "Fútbol")
	if err != nil {
		t.Fatalf("error al crear el deporte: %v", err)
	}

	if deporte.IDDeporte == 0 {
		t.Errorf("el deporte creado no recibió un ID")
	}

	horaApertura := time.Date(
		2000,
		time.January,
		1,
		8,
		0,
		0,
		0,
		time.Local,
	)

	horaCierre := time.Date(
		2000,
		time.January,
		1,
		23,
		0,
		0,
		0,
		time.Local,
	)

	canchaCreada, err := queries.CreateCancha(
		ctx,
		sqlc.CreateCanchaParams{
			Nombre:               "Cancha de prueba",
			Precio:               "15000.00",
			DuracionTurnoMinutos: 60,
			HoraApertura:         horaApertura.Format("15:04:05"),
			HoraCierre:           horaCierre.Format("15:04:05"),
			IDInstitucion:        institucion.IDInstitucion,
			IDDeporte:            deporte.IDDeporte,
		},
	)
	if err != nil {
		t.Fatalf("error al crear la cancha: %v", err)
	}

	if canchaCreada.IDCancha == 0 {
		t.Errorf("la cancha creada no recibió un ID")
	}

	canchaObtenida, err := queries.GetCancha(
		ctx,
		canchaCreada.IDCancha,
	)
	if err != nil {
		t.Fatalf("error al obtener la cancha: %v", err)
	}

	if canchaObtenida.IDCancha != canchaCreada.IDCancha {
		t.Errorf(
			"ID incorrecto: se esperaba %d y se obtuvo %d",
			canchaCreada.IDCancha,
			canchaObtenida.IDCancha,
		)
	}

	if canchaObtenida.Nombre != canchaCreada.Nombre {
		t.Errorf(
			"nombre incorrecto: se esperaba %q y se obtuvo %q",
			canchaCreada.Nombre,
			canchaObtenida.Nombre,
		)
	}

	if canchaObtenida.Precio != canchaCreada.Precio {
		t.Errorf(
			"precio incorrecto: se esperaba %s y se obtuvo %s",
			canchaCreada.Precio,
			canchaObtenida.Precio,
		)
	}

	if canchaObtenida.IDInstitucion != institucion.IDInstitucion {
		t.Errorf("la institución recuperada no coincide")
	}

	if canchaObtenida.IDDeporte != deporte.IDDeporte {
		t.Errorf("el deporte recuperado no coincide")
	}

	filasActualizadas, err := queries.UpdatePrecioCancha(
		ctx,
		sqlc.UpdatePrecioCanchaParams{
			IDCancha: canchaCreada.IDCancha,
			Precio:   "18000.00",
		},
	)
	if err != nil {
		t.Fatalf("error al actualizar el precio: %v", err)
	}

	if filasActualizadas != 1 {
		t.Errorf(
			"se esperaba actualizar una fila, pero se actualizaron %d",
			filasActualizadas,
		)
	}

	canchaActualizada, err := queries.GetCancha(
		ctx,
		canchaCreada.IDCancha,
	)
	if err != nil {
		t.Fatalf("error al obtener la cancha actualizada: %v", err)
	}

	if canchaActualizada.Precio != "18000.00" {
		t.Errorf(
			"el precio no se actualizó: se obtuvo %s",
			canchaActualizada.Precio,
		)
	}

	canchas, err := queries.ListCanchas(ctx)
	if err != nil {
		t.Fatalf("error al listar las canchas: %v", err)
	}

	encontrada := false

	for _, cancha := range canchas {
		if cancha.IDCancha == canchaCreada.IDCancha {
			encontrada = true

			if cancha.Precio != "18000.00" {
				t.Errorf(
					"el precio de la cancha listada no coincide",
				)
			}
		}
	}

	if !encontrada {
		t.Errorf("la cancha creada no aparece en la lista")
	}

	filasEliminadas, err := queries.DeleteCancha(
		ctx,
		canchaCreada.IDCancha,
	)
	if err != nil {
		t.Fatalf("error al eliminar la cancha: %v", err)
	}

	if filasEliminadas != 1 {
		t.Errorf(
			"se esperaba eliminar una fila, pero se eliminaron %d",
			filasEliminadas,
		)
	}

	_, err = queries.GetCancha(
		ctx,
		canchaCreada.IDCancha,
	)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf(
			"la cancha no se eliminó correctamente; error obtenido: %v",
			err,
		)
	}
}
