package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	sqlc "tp2web/db/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func abrirDB() (*sql.DB, error) {
	db, err := sql.Open("pgx", "user=CanchAppADMIN password=canchappweb2026 dbname=CanchApp host=localhost port=5432 sslmode=disable")

	if err != nil {
		return nil, fmt.Errorf("error al iniciar la base de datos : %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error en la conexion con la BD: %w", err)
	}

	db.SetMaxOpenConns(25)
	return db, nil
}

func main() {

	database, err := abrirDB()
	if err != nil {
		log.Fatalf("error al abrir la base de datos: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	queries := sqlc.New(database)

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

	createdCancha, err := queries.CreateCancha(
		ctx,
		sqlc.CreateCanchaParams{
			Nombre:               "Cancha principal",
			Precio:               "15000.00",
			DuracionTurnoMinutos: 60,
			HoraApertura:         horaApertura,
			HoraCierre:           horaCierre,
			IDInstitucion:        1,
			IDDeporte:            1,
		},
	)
	if err != nil {
		log.Fatalf("error al crear la cancha: %v", err)
	}

	fmt.Printf("Cancha creada: %+v\n", createdCancha)

	cancha, err := queries.GetCancha(
		ctx,
		createdCancha.IDCancha,
	)
	if err != nil {
		log.Fatalf("error al obtener la cancha: %v", err)
	}

	fmt.Printf("Cancha obtenida: %+v\n", cancha)

	canchas, err := queries.ListCanchas(ctx)
	if err != nil {
		log.Fatalf("error al listar las canchas: %v", err)
	}

	fmt.Printf("Todas las canchas: %+v\n", canchas)

	filasActualizadas, err := queries.UpdatePrecioCancha(
		ctx,
		sqlc.UpdatePrecioCanchaParams{
			IDCancha: createdCancha.IDCancha,
			Precio:   "18000.00",
		},
	)
	if err != nil {
		log.Fatalf(
			"error al actualizar el precio de la cancha: %v",
			err,
		)
	}

	if filasActualizadas == 0 {
		log.Fatal(
			"no se encontró la cancha que se quería actualizar",
		)
	}

	fmt.Println("Precio de la cancha actualizado correctamente")

	updatedCancha, err := queries.GetCancha(
		ctx,
		createdCancha.IDCancha,
	)
	if err != nil {
		log.Fatalf(
			"error al obtener la cancha actualizada: %v",
			err,
		)
	}

	fmt.Printf("Cancha actualizada: %+v\n", updatedCancha)

	filasEliminadas, err := queries.DeleteCancha(
		ctx,
		createdCancha.IDCancha,
	)
	if err != nil {
		log.Fatalf("error al eliminar la cancha: %v", err)
	}

	if filasEliminadas == 0 {
		log.Fatal(
			"no se encontró la cancha que se quería eliminar",
		)
	}

	fmt.Println("Cancha eliminada correctamente")

	_, err = queries.GetCancha(
		ctx,
		createdCancha.IDCancha,
	)

	if err == sql.ErrNoRows {
		fmt.Println(
			"La cancha ya no existe después de eliminarla",
		)
	} else if err != nil {
		log.Fatalf(
			"error al comprobar la eliminación de la cancha: %v",
			err,
		)
	} else {
		log.Fatal(
			"la cancha todavía existe después de eliminarla",
		)
	}
}
