package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bico/internal/dto"
)

func BuscarCoordPorCep(cep string) (float64, float64, error) {
	cep = strings.NewReplacer("-", "", ".", "", " ", "").Replace(cep)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://brasilapi.com.br/api/cep/v2/" + cep)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("cep não encontrado (status %d)", resp.StatusCode)
	}

	var res struct {
		Location struct {
			Coordinates struct {
				Latitude  string `json:"latitude"`
				Longitude string `json:"longitude"`
			} `json:"coordinates"`
		} `json:"location"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, 0, err
	}
	if res.Location.Coordinates.Latitude == "" || res.Location.Coordinates.Longitude == "" {
		return 0, 0, fmt.Errorf("cep sem coordenadas disponíveis")
	}
	lat, err := strconv.ParseFloat(res.Location.Coordinates.Latitude, 64)
	if err != nil {
		return 0, 0, err
	}
	lng, err := strconv.ParseFloat(res.Location.Coordinates.Longitude, 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lng, nil
}

func CalcDistanciaPorCoord(lat1, lng1, lat2, lng2 float64) float64 {
	const raioTerra = 6371.0

	rad := func(graus float64) float64 {
		return graus * math.Pi / 180
	}

	dLat := rad(lat2 - lat1)
	dLng := rad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(rad(lat1))*math.Cos(rad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return raioTerra * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func OrdPorDistancia(prestadores []dto.PrestadorPublicoDTO) []dto.PrestadorPublicoDTO {
	if len(prestadores) <= 1 {
		return prestadores
	}

	n := prestadores[len(prestadores)/2].DistanciaKm
	var menores, iguais, maiores []dto.PrestadorPublicoDTO

	for _, p := range prestadores {
		switch {
		case p.DistanciaKm < n:
			menores = append(menores, p)
		case p.DistanciaKm == n:
			iguais = append(iguais, p)
		default:
			maiores = append(maiores, p)
		}
	}

	resultado := append(OrdPorDistancia(menores), iguais...)
	return append(resultado, OrdPorDistancia(maiores)...)
}
