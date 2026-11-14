package paquete

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const ArchivoEliminados = "ELIMINADOS.txt"

func Nombre(rama, sufijo string) string {
	if i := strings.Index(rama, "/"); i >= 0 {
		rama = rama[i+1:]
	}
	return strings.ReplaceAll(rama, "/", "-") + sufijo
}

func CopiarArchivo(origen, destino string) error {
	entrada, err := os.Open(origen)
	if err != nil {
		return err
	}
	defer entrada.Close()
	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		return err
	}
	salida, err := os.Create(destino)
	if err != nil {
		return err
	}
	if _, err := io.Copy(salida, entrada); err != nil {
		salida.Close()
		return err
	}
	return salida.Close()
}

func CopiarCarpeta(origen, destino string) (int, error) {
	total := 0
	err := filepath.WalkDir(origen, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		relativa, err := filepath.Rel(origen, ruta)
		if err != nil {
			return err
		}
		total++
		return CopiarArchivo(ruta, filepath.Join(destino, relativa))
	})
	return total, err
}

func EscribirEliminados(carpeta string, rutas []string) error {
	if len(rutas) == 0 {
		return nil
	}
	contenido := strings.Join(rutas, "\r\n") + "\r\n"
	return os.WriteFile(filepath.Join(carpeta, ArchivoEliminados), []byte(contenido), 0o644)
}
