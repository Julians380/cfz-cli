package estilo

import (
	"fmt"
	"io"
	"os"
)

var (
	terminal bool
	activo   bool
)

func Activar(salida *os.File) bool {
	terminal = habilitarTerminal(salida)
	activo = terminal && os.Getenv("NO_COLOR") == ""
	return activo
}

func Limpiar(salida io.Writer) {
	if terminal {
		fmt.Fprint(salida, "\x1b[2J\x1b[3J\x1b[H")
	}
}

func aplicar(codigo, texto string) string {
	if !activo {
		return texto
	}
	return "\x1b[" + codigo + "m" + texto + "\x1b[0m"
}

func Negrita(texto string) string  { return aplicar("1", texto) }
func Rojo(texto string) string     { return aplicar("31", texto) }
func Verde(texto string) string    { return aplicar("32", texto) }
func Amarillo(texto string) string { return aplicar("33", texto) }
func Azul(texto string) string     { return aplicar("34", texto) }
func Gris(texto string) string     { return aplicar("90", texto) }
