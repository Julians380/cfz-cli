package paquete

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Entrada struct {
	RaizRepo   string
	Archivos   []string
	Eliminados []string
	Bin        string
	Salida     string
	Nombre     string
	Forzar     bool
}

type Resultado struct {
	Carpeta     string
	Zip         string
	Archivos    int
	ArchivosBin int
}

func Comprimir(carpeta, destino string) error {
	archivo, err := os.Create(destino)
	if err != nil {
		return err
	}
	escritor := zip.NewWriter(archivo)
	if err := escritor.AddFS(os.DirFS(carpeta)); err != nil {
		escritor.Close()
		archivo.Close()
		return err
	}
	if err := escritor.Close(); err != nil {
		archivo.Close()
		return err
	}
	return archivo.Close()
}

func Armar(e Entrada) (Resultado, error) {
	var r Resultado
	if e.Nombre == "" || e.Nombre == "." || e.Nombre == ".." || strings.ContainsAny(e.Nombre, `/\`) {
		return r, fmt.Errorf("nombre de paquete inválido: %q", e.Nombre)
	}
	r.Carpeta = filepath.Join(e.Salida, e.Nombre)
	r.Zip = r.Carpeta + ".zip"
	for _, ruta := range []string{r.Carpeta, r.Zip} {
		if _, err := os.Stat(ruta); err != nil {
			continue
		}
		if !e.Forzar {
			return r, fmt.Errorf("ya existe %s (usa --force para reemplazarlo)", ruta)
		}
		if err := os.RemoveAll(ruta); err != nil {
			return r, err
		}
	}
	if err := os.MkdirAll(r.Carpeta, 0o755); err != nil {
		return r, err
	}
	for _, ruta := range e.Archivos {
		origen := filepath.Join(e.RaizRepo, filepath.FromSlash(ruta))
		destino := filepath.Join(r.Carpeta, filepath.FromSlash(ruta))
		if err := CopiarArchivo(origen, destino); err != nil {
			return r, fmt.Errorf("no se pudo copiar %s: %w", ruta, err)
		}
		r.Archivos++
	}
	if e.Bin != "" {
		origen := filepath.Join(e.RaizRepo, filepath.FromSlash(e.Bin))
		destino := filepath.Join(r.Carpeta, filepath.FromSlash(e.Bin))
		total, err := CopiarCarpeta(origen, destino)
		if err != nil {
			return r, fmt.Errorf("no se pudo copiar el bin: %w", err)
		}
		r.ArchivosBin = total
	}
	if err := EscribirEliminados(r.Carpeta, e.Eliminados); err != nil {
		return r, err
	}
	if err := Comprimir(r.Carpeta, r.Zip); err != nil {
		return r, err
	}
	return r, nil
}
