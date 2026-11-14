package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cli-cfz/internal/config"
	"cli-cfz/internal/estilo"
	"cli-cfz/internal/git"
)

const version = "0.1.0"

const logo = `
   ██████╗███████╗███████╗
  ██╔════╝██╔════╝╚══███╔╝
  ██║     █████╗    ███╔╝
  ██║     ██╔══╝   ███╔╝
  ╚██████╗██║     ███████╗
   ╚═════╝╚═╝     ╚══════╝
`

func bienvenida(salida io.Writer) {
	fmt.Fprint(salida, estilo.Verde(logo))
	fmt.Fprintf(salida, "\n cli-cfz v%s · empaquetador\n\n", version)
	directorio, err := os.Getwd()
	if err != nil {
		directorio = "."
	}
	raiz, err := git.Repo{Dir: directorio}.Raiz()
	if err != nil {
		fmt.Fprintf(salida, " Carpeta:     %s (no es un repositorio git)\n", directorio)
	} else {
		rama, err := git.Repo{Dir: raiz}.RamaActual()
		if err != nil {
			rama = "(sin rama)"
		}
		estadoConfig := config.NombreArchivo
		if _, err := os.Stat(filepath.Join(raiz, config.NombreArchivo)); err != nil {
			estadoConfig = "falta " + config.NombreArchivo
		}
		fmt.Fprintf(salida, " Repositorio: %s\n Rama:        %s\n Config:      %s\n", raiz, rama, estadoConfig)
	}
	fmt.Fprintln(salida, "\n Escribe 'help' para ver los comandos o 'exit' para terminar.")
}

func consola(entrada io.Reader, salida io.Writer) {
	estilo.Limpiar(salida)
	bienvenida(salida)
	lector := bufio.NewScanner(entrada)
	for {
		fmt.Fprint(salida, "\n"+estilo.Verde("cfz>")+" ")
		if !lector.Scan() {
			fmt.Fprintln(salida)
			return
		}
		args := strings.Fields(lector.Text())
		if len(args) > 0 && args[0] == "cfz" {
			args = args[1:]
		}
		if len(args) == 0 {
			continue
		}
		if args[0] == "exit" {
			return
		}
		ejecutar(args)
	}
}
