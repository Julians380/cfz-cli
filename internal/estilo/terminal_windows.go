//go:build windows

package estilo

import (
	"os"

	"golang.org/x/sys/windows"
)

func habilitarTerminal(salida *os.File) bool {
	consola := windows.Handle(salida.Fd())
	var modo uint32
	if err := windows.GetConsoleMode(consola, &modo); err != nil {
		return false
	}
	return windows.SetConsoleMode(consola, modo|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}
