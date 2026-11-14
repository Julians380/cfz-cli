package filtro

import (
	"fmt"
	"strings"

	"cli-cfz/internal/git"

	"github.com/bmatcuk/doublestar/v4"
)

func Validar(patrones []string) error {
	for _, patron := range patrones {
		if !doublestar.ValidatePattern(patron) {
			return fmt.Errorf("patrón inválido: %q", patron)
		}
	}
	return nil
}

func Coincide(ruta string, patrones []string) bool {
	ruta = strings.ToLower(ruta)
	for _, patron := range patrones {
		if ok, _ := doublestar.Match(strings.ToLower(patron), ruta); ok {
			return true
		}
	}
	return false
}

func Separar(cambios []git.Cambio, patrones []string) (incluidos, excluidos []git.Cambio) {
	for _, cambio := range cambios {
		if Coincide(cambio.Ruta, patrones) {
			excluidos = append(excluidos, cambio)
		} else {
			incluidos = append(incluidos, cambio)
		}
	}
	return incluidos, excluidos
}
