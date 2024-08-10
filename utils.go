package main

import (
	"strconv"
	"strings"
	"time"
)

func parseInput(content string) Faktura {
	lines := strings.Split(content, "\n")
	faktura := Faktura{
		DatumVystavenia: time.Now(),
		DatumSplatnosti: time.Now().AddDate(0, 0, 14),
	}

	for i, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.TrimSpace(parts[1])

		switch i {
		case 0:
			faktura.Cislo = value
		case 1:
			faktura.Odberatel = value
		default:
			polozka := strings.Split(value, ",")
			if len(polozka) == 3 {
				nazov := strings.TrimSpace(polozka[0])
				mnozstvo, _ := strconv.Atoi(strings.TrimSpace(polozka[1]))
				cena, _ := strconv.ParseFloat(strings.TrimSpace(polozka[2]), 64)
				faktura.PridajPolozku(nazov, mnozstvo, cena)
			}
		}
	}

	return faktura
}
