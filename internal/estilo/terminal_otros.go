//go:build !windows

package estilo

import "os"

func habilitarTerminal(salida *os.File) bool {
	info, err := salida.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
