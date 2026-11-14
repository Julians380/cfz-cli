package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"cli-cfz/internal/config"
	"cli-cfz/internal/estilo"
	"cli-cfz/internal/filtro"
	"cli-cfz/internal/git"
	"cli-cfz/internal/paquete"
)

const ramaAyuda = "compara la rama contra ramaBase en vez de leer los cambios sin commitear"

const uso = `comandos:
  init [--force]                     crea un cfz.yaml inicial en la raíz del repo
  list [--branch <rama>]             muestra qué entraría al paquete
  pack [--branch <rama>] [--force] [--build]
                                     arma la carpeta y el zip
  help                               muestra esta ayuda
  exit                               cierra el modo consola

sin --branch se leen los cambios sin commitear (incluye archivos nuevos);
con --branch se compara esa rama contra ramaBase`

func main() {
	estilo.Activar(os.Stdout)
	if len(os.Args) == 1 {
		consola(os.Stdin, os.Stdout)
		return
	}
	os.Exit(ejecutar(os.Args[1:]))
}

func ejecutar(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, uso)
		return 2
	}
	var err error
	switch args[0] {
	case "init":
		err = iniciar(args[1:])
	case "list":
		err = listar(args[1:])
	case "pack":
		err = empaquetar(args[1:])
	case "help":
		fmt.Println(uso)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "%s %s\n\n%s\n", estilo.Rojo("comando desconocido:"), args[0], uso)
		return 2
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, estilo.Rojo("error:"), err)
		return 1
	}
	return 0
}

type contexto struct {
	cfg       config.Config
	repo      git.Repo
	rama      string
	modoRama  bool
	incluidos []git.Cambio
	excluidos []git.Cambio
}

func preparar(rama string) (contexto, error) {
	var c contexto
	directorio, err := os.Getwd()
	if err != nil {
		return c, err
	}
	raiz, err := git.Repo{Dir: directorio}.Raiz()
	if err != nil {
		return c, fmt.Errorf("%s no es un repositorio git", directorio)
	}
	c.repo = git.Repo{Dir: filepath.Clean(raiz)}
	if c.cfg, err = config.Leer(filepath.Join(c.repo.Dir, config.NombreArchivo)); err != nil {
		return c, err
	}
	if !filepath.IsAbs(c.cfg.Salida) {
		c.cfg.Salida = filepath.Join(c.repo.Dir, c.cfg.Salida)
	}
	if err := filtro.Validar(slices.Concat(c.cfg.Excluir, c.cfg.Fuentes)); err != nil {
		return c, err
	}
	c.rama = rama
	c.modoRama = rama != ""
	var cambios []git.Cambio
	if c.modoRama {
		cambios, err = c.repo.Diferencias(c.cfg.RamaBase, c.rama)
	} else {
		if c.rama, err = c.repo.RamaActual(); err != nil {
			c.rama = "cambios"
		}
		cambios, err = c.repo.CambiosLocales()
	}
	if err != nil {
		return c, err
	}
	c.incluidos, c.excluidos = filtro.Separar(cambios, c.cfg.Excluir)
	return c, nil
}

func describir(cambio git.Cambio) string {
	if cambio.Estado == "R" {
		return fmt.Sprintf("R  %s  (antes %s)", cambio.Ruta, cambio.RutaAnterior)
	}
	return cambio.Estado + "  " + cambio.Ruta
}

func colorDeEstado(estado string) func(string) string {
	switch estado {
	case "A":
		return estilo.Verde
	case "D":
		return estilo.Rojo
	case "R":
		return estilo.Azul
	default:
		return estilo.Amarillo
	}
}

func iniciar(args []string) error {
	banderas := flag.NewFlagSet("init", flag.ContinueOnError)
	forzar := banderas.Bool("force", false, "reemplaza un cfz.yaml existente")
	if err := banderas.Parse(args); err != nil {
		return err
	}
	directorio, err := os.Getwd()
	if err != nil {
		return err
	}
	raiz, err := git.Repo{Dir: directorio}.Raiz()
	if err != nil {
		return fmt.Errorf("%s no es un repositorio git", directorio)
	}
	ruta := filepath.Join(filepath.Clean(raiz), config.NombreArchivo)
	if err := config.Crear(ruta, *forzar); err != nil {
		return err
	}
	fmt.Printf("Creado %s\nRevisa ramaBase y salida antes de usar cfz pack.\n", ruta)
	return nil
}

func listar(args []string) error {
	banderas := flag.NewFlagSet("list", flag.ContinueOnError)
	rama := banderas.String("branch", "", ramaAyuda)
	if err := banderas.Parse(args); err != nil {
		return err
	}
	c, err := preparar(*rama)
	if err != nil {
		return err
	}
	if c.modoRama {
		fmt.Printf("%s contra %s\n\n", c.rama, c.cfg.RamaBase)
	} else {
		fmt.Printf("cambios sin commitear en %s\n\n", c.rama)
	}
	fmt.Printf("%s\n", estilo.Negrita(fmt.Sprintf("Incluidos (%d):", len(c.incluidos))))
	for _, cambio := range c.incluidos {
		fmt.Println("  " + colorDeEstado(cambio.Estado)(describir(cambio)))
	}
	fmt.Printf("\n%s\n", estilo.Gris(fmt.Sprintf("Excluidos (%d):", len(c.excluidos))))
	for _, cambio := range c.excluidos {
		fmt.Println("  " + estilo.Gris(describir(cambio)))
	}
	return nil
}

func revisarBin(c contexto) string {
	if c.cfg.DllPrincipal == "" || len(c.cfg.Fuentes) == 0 {
		return ""
	}
	dll, err := os.Stat(filepath.Join(c.repo.Dir, filepath.FromSlash(c.cfg.DllPrincipal)))
	if err != nil {
		return fmt.Sprintf("no se encontró %s, ¿compilaste?", c.cfg.DllPrincipal)
	}
	for _, cambio := range slices.Concat(c.incluidos, c.excluidos) {
		if cambio.Estado == "D" || !filtro.Coincide(cambio.Ruta, c.cfg.Fuentes) {
			continue
		}
		fuente, err := os.Stat(filepath.Join(c.repo.Dir, filepath.FromSlash(cambio.Ruta)))
		if err == nil && fuente.ModTime().After(dll.ModTime()) {
			return fmt.Sprintf("%s es más reciente que %s, vuelve a compilar", cambio.Ruta, c.cfg.DllPrincipal)
		}
	}
	return ""
}

func compilarProyecto(c contexto) error {
	var comando *exec.Cmd
	if runtime.GOOS == "windows" {
		comando = exec.Command("cmd", "/C", c.cfg.Compilar)
	} else {
		comando = exec.Command("sh", "-c", c.cfg.Compilar)
	}
	comando.Dir = c.repo.Dir
	comando.Stdout = os.Stdout
	comando.Stderr = os.Stderr
	fmt.Println(estilo.Negrita("Compilando:"), c.cfg.Compilar)
	if err := comando.Run(); err != nil {
		return fmt.Errorf("la compilación falló, no se creó el paquete: %w", err)
	}
	return nil
}

func empaquetar(args []string) error {
	banderas := flag.NewFlagSet("pack", flag.ContinueOnError)
	rama := banderas.String("branch", "", ramaAyuda)
	forzar := banderas.Bool("force", false, "reemplaza un paquete existente")
	compilar := banderas.Bool("build", false, "compila el proyecto (comando 'compilar' del cfz.yaml) antes de empaquetar")
	if err := banderas.Parse(args); err != nil {
		return err
	}
	c, err := preparar(*rama)
	if err != nil {
		return err
	}
	if *compilar && c.cfg.Compilar == "" {
		return errors.New("--build necesita el campo compilar en cfz.yaml (por ejemplo: msbuild App.sln /p:Configuration=Release)")
	}
	if c.modoRama {
		actual, err := c.repo.RamaActual()
		if err != nil {
			return err
		}
		if actual != c.rama {
			return fmt.Errorf("estás en %s y pediste empaquetar %s; haz checkout de %s primero", actual, c.rama, c.rama)
		}
		sucios, err := c.repo.ArchivosSinCommitear()
		if err != nil {
			return err
		}
		if len(sucios) > 0 {
			return fmt.Errorf("hay cambios sin commitear:\n  %s", strings.Join(sucios, "\n  "))
		}
	}
	if len(c.incluidos) == 0 {
		return errors.New("no hay archivos para empaquetar")
	}
	if *compilar {
		if err := compilarProyecto(c); err != nil {
			return err
		}
	}
	var archivos, eliminados []string
	for _, cambio := range c.incluidos {
		switch cambio.Estado {
		case "D":
			eliminados = append(eliminados, cambio.Ruta)
		case "R":
			archivos = append(archivos, cambio.Ruta)
			eliminados = append(eliminados, cambio.RutaAnterior)
		default:
			archivos = append(archivos, cambio.Ruta)
		}
	}
	if aviso := revisarBin(c); aviso != "" {
		fmt.Println(estilo.Amarillo("ADVERTENCIA:"), aviso)
	}
	r, err := paquete.Armar(paquete.Entrada{
		RaizRepo:   c.repo.Dir,
		Archivos:   archivos,
		Eliminados: eliminados,
		Bin:        c.cfg.Bin,
		Salida:     c.cfg.Salida,
		Nombre:     paquete.Nombre(c.rama, c.cfg.Sufijo),
		Forzar:     *forzar,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Archivos copiados: %d\nArchivos del bin:  %d\nExcluidos:         %d\nEliminados:        %d\n\nCarpeta: %s\nZip:     %s\n",
		r.Archivos, r.ArchivosBin, len(c.excluidos), len(eliminados), r.Carpeta, r.Zip)
	return nil
}
