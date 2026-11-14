package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Repo struct {
	Dir string
}

func (r Repo) ejecutar(args ...string) (string, error) {
	comando := exec.Command("git", append([]string{"-C", r.Dir}, args...)...)
	var salida, errores bytes.Buffer
	comando.Stdout = &salida
	comando.Stderr = &errores
	if err := comando.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(errores.String()))
	}
	return salida.String(), nil
}

func (r Repo) Raiz() (string, error) {
	salida, err := r.ejecutar("rev-parse", "--show-toplevel")
	return strings.TrimSpace(salida), err
}

func (r Repo) RamaActual() (string, error) {
	salida, err := r.ejecutar("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	rama := strings.TrimSpace(salida)
	if rama == "HEAD" {
		return "", errors.New("no estás en ninguna rama (HEAD desacoplado)")
	}
	return rama, nil
}

func (r Repo) ArchivosSinCommitear() ([]string, error) {
	salida, err := r.ejecutar("status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return nil, err
	}
	var archivos []string
	for _, linea := range strings.Split(salida, "\n") {
		if linea = strings.TrimSpace(linea); linea != "" {
			archivos = append(archivos, linea)
		}
	}
	return archivos, nil
}

func (r Repo) CambiosLocales() ([]Cambio, error) {
	salida, err := r.ejecutar("diff", "--name-status", "-z", "-M", "HEAD")
	if err != nil {
		return nil, err
	}
	cambios, err := ParsearNameStatus(salida)
	if err != nil {
		return nil, err
	}
	nuevos, err := r.ejecutar("ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	for _, ruta := range strings.Split(strings.TrimRight(nuevos, "\x00"), "\x00") {
		if ruta != "" {
			cambios = append(cambios, Cambio{Estado: "A", Ruta: ruta})
		}
	}
	return cambios, nil
}

func (r Repo) Diferencias(base, rama string) ([]Cambio, error) {
	salida, err := r.ejecutar("diff", "--name-status", "-z", "-M", base+"..."+rama)
	if err != nil {
		return nil, err
	}
	return ParsearNameStatus(salida)
}
