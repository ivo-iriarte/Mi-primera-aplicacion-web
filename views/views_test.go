package views

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	sqlc "canchapp/db/sqlc"
	"canchapp/logic"
)

func TestRenderizadoVistasSinJS(t *testing.T) {
	ctx := context.Background()

	t.Run("DeportesPage", func(t *testing.T) {
		var buf bytes.Buffer
		deportes := []sqlc.Deporte{
			{IDDeporte: 1, Nombre: "Fútbol", Activo: true},
			{IDDeporte: 2, Nombre: "Tenis", Activo: true},
		}
		if err := DeportesPage(deportes).Render(ctx, &buf); err != nil {
			t.Fatalf("error al renderizar DeportesPage: %v", err)
		}
		html := buf.String()
		if !strings.Contains(html, "Fútbol") || !strings.Contains(html, "/app/deportes/1") {
			t.Errorf("DeportesPage no contiene los enlaces esperados")
		}
		if strings.Contains(html, "hx-on") || strings.Contains(html, "onclick") {
			t.Errorf("se detectó JavaScript inline no permitido en DeportesPage")
		}
	})

	t.Run("CanchaDetallePage_TurnosDisponiblesYOcupados", func(t *testing.T) {
		var buf bytes.Buffer
		fecha := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)
		vm := DisponibilidadView{
			Cancha: CanchaCardView{
				IDCancha:             3,
				Nombre:               "Cancha Sintético 1",
				Precio:               "18000.00",
				Activa:               true,
				DuracionTurnoMinutos: 60,
				HoraApertura:         "08:00:00",
				HoraCierre:           "11:00:00",
				IDInstitucion:        1,
				InstitucionNombre:    "Club Náutico Necochea",
				IDDeporte:            1,
				DeporteNombre:        "Fútbol",
				Servicios:            []sqlc.Servicio{{IDServicio: 1, Servicio: "Iluminación LED", Activo: true}},
			},
			FechaSeleccionada:    "2026-12-01",
			FechaLegible:         FormatearFechaLegible(fecha),
			Dias:                 ConstruirDiasSelector(fecha, fecha),
			HayTurnosDisponibles: true,
			Turnos: []logic.Turno{
				{HoraInicio: "08:00", HoraFin: "09:00", Disponible: true},
				{HoraInicio: "09:00", HoraFin: "10:00", Disponible: false},
			},
			Usuarios:              []sqlc.Usuario{{IDUsuario: 1, Nombre: "Martín", Apellido: "Santos", Email: "martin@canchapp.com"}},
			IDUsuarioSeleccionado: 1,
		}

		if err := CanchaDetallePage(vm).Render(ctx, &buf); err != nil {
			t.Fatalf("error al renderizar CanchaDetallePage: %v", err)
		}
		html := buf.String()

		if !strings.Contains(html, `value="08:00"`) {
			t.Errorf("el turno disponible 08:00 debería renderizar un radio seleccionable")
		}
		if strings.Contains(html, `value="09:00"`) || !strings.Contains(html, "Reservado") {
			t.Errorf("el turno ocupado 09:00 debería figurar bloqueado como Reservado y sin input seleccionable")
		}
		if !strings.Contains(html, "Iluminación LED") {
			t.Errorf("no se renderizaron los servicios de la cancha")
		}
		if strings.Contains(html, "hx-on") || strings.Contains(html, "onclick") {
			t.Errorf("se detectó JavaScript inline no permitido en CanchaDetallePage")
		}
	})
}
