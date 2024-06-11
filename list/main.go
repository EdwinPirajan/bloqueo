package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/process"
)

func main() {
	if runtime.GOOS != "windows" {
		fmt.Println("Este código solo funciona en Windows.")
		return
	}

	procs, err := process.Processes()
	if err != nil {
		fmt.Println("Error al obtener procesos:", err)
		return
	}

	var userProcs []string
	excludeNames := map[string]bool{
		"svchost.exe": true,
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Filtrar procesos del sistema y procesos específicos
		if p.Pid < 1000 || excludeNames[name] {
			continue
		}

		title, err := p.CmdlineSlice()
		if err != nil {
			continue
		}

		userProcs = append(userProcs, fmt.Sprintf("Nombre: %s\nTítulo: %s\n\n", name, strings.Join(title, " ")))
	}

	// Ordenar los procesos alfabéticamente por nombre
	sort.Slice(userProcs, func(i, j int) bool {
		return strings.ToLower(userProcs[i]) < strings.ToLower(userProcs[j])
	})

	fmt.Println("Procesos activos en Windows (excepto del sistema y procesos específicos) ordenados alfabéticamente:")
	for _, proc := range userProcs {
		fmt.Println(proc)
	}
}
