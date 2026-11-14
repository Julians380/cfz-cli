package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const NombreArchivo = "cfz.yaml"

type Config struct {
	RamaBase     string   `yaml:"ramaBase"`
	Bin          string   `yaml:"bin"`
	Compilar     string   `yaml:"compilar"`
	DllPrincipal string   `yaml:"dllPrincipal"`
	Fuentes      []string `yaml:"fuentes"`
	Salida       string   `yaml:"salida"`
	Sufijo       string   `yaml:"sufijo"`
	Excluir      []string `yaml:"excluir"`
}

const plantilla = `# Rama contra la que se compara con --branch (main u origin/main)
ramaBase: main

# Carpeta donde se crean el paquete y el zip. Si es relativa, se toma desde la raíz del repo
salida: ../pases

# Texto agregado al nombre del paquete
# sufijo: _V1

# Carpeta que se copia completa dentro del paquete
# bin: bin

# Comando que compila el proyecto; se ejecuta desde la raíz del repo con cfz pack --build
# compilar: msbuild App.sln /p:Configuration=Release

# Archivo compilado que se compara contra "fuentes" para avisar si falta recompilar
# dllPrincipal: bin/App.dll
# fuentes:
#   - "**/*.cs"
#   - "**/*.csproj"

# Patrones de rutas que nunca entran al paquete (aceptan ** y no distinguen mayúsculas)
excluir:
  - "cfz.yaml"
`

// Crear escribe un cfz.yaml inicial. Si ya existe, falla salvo que forzar sea true.
func Crear(ruta string, forzar bool) error {
	banderas := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if forzar {
		banderas = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	archivo, err := os.OpenFile(ruta, banderas, 0o644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("ya existe %s (usa --force para reemplazarlo)", ruta)
	}
	if err != nil {
		return err
	}
	if _, err := archivo.WriteString(plantilla); err != nil {
		archivo.Close()
		return err
	}
	return archivo.Close()
}

func Leer(ruta string) (Config, error) {
	var cfg Config
	contenido, err := os.ReadFile(ruta)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, fmt.Errorf("no se encontró %s", ruta)
	}
	if err != nil {
		return cfg, err
	}
	decodificador := yaml.NewDecoder(bytes.NewReader(contenido))
	decodificador.KnownFields(true)
	if err := decodificador.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("%s inválido: %w", ruta, err)
	}
	if cfg.RamaBase == "" {
		return cfg, fmt.Errorf("%s: falta el campo ramaBase", ruta)
	}
	if cfg.Salida == "" {
		return cfg, fmt.Errorf("%s: falta el campo salida", ruta)
	}
	return cfg, nil
}
