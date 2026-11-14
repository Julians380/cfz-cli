package git

import (
	"fmt"
	"strings"
)

type Cambio struct {
	Estado       string
	Ruta         string
	RutaAnterior string
}

func ParsearNameStatus(salida string) ([]Cambio, error) {
	salida = strings.TrimRight(salida, "\x00")
	if salida == "" {
		return nil, nil
	}
	partes := strings.Split(salida, "\x00")
	var cambios []Cambio
	for i := 0; i < len(partes); {
		estado := partes[i][:1]
		switch estado {
		case "A", "M", "D", "T":
			if i+1 >= len(partes) {
				return nil, fmt.Errorf("salida de git incompleta tras %q", partes[i])
			}
			if estado == "T" {
				estado = "M"
			}
			cambios = append(cambios, Cambio{Estado: estado, Ruta: partes[i+1]})
			i += 2
		case "R":
			if i+2 >= len(partes) {
				return nil, fmt.Errorf("salida de git incompleta tras %q", partes[i])
			}
			cambios = append(cambios, Cambio{Estado: "R", RutaAnterior: partes[i+1], Ruta: partes[i+2]})
			i += 3
		default:
			return nil, fmt.Errorf("estado de git desconocido: %q", partes[i])
		}
	}
	return cambios, nil
}
