package main

import (
	"time"
)

type Polozka struct {
	Nazov    string
	Mnozstvo int
	Cena     float64
}

type Faktura struct {
	Cislo           string
	DatumVystavenia time.Time
	DatumSplatnosti time.Time
	Odberatel       string
	Polozky         []Polozka
}

func (f *Faktura) PridajPolozku(nazov string, mnozstvo int, cena float64) {
	f.Polozky = append(f.Polozky, Polozka{Nazov: nazov, Mnozstvo: mnozstvo, Cena: cena})
}

func (f *Faktura) CelkovaSuma() float64 {
	suma := 0.0
	for _, p := range f.Polozky {
		suma += float64(p.Mnozstvo) * p.Cena
	}
	return suma
}
