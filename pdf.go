package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jung-kurt/gofpdf"
)

func generatePDF(f Faktura) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Nastavenie fontu s podporou diakritiky
	pdf.AddUTF8Font("DejaVu", "", "DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVu", "", 16)

	pdf.Cell(40, 10, "Faktúra")

	pdf.Ln(10)
	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Číslo faktúry: %s", f.Cislo))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Dátum vystavenia: %s", f.DatumVystavenia.Format("02.01.2006")))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Dátum splatnosti: %s", f.DatumSplatnosti.Format("02.01.2006")))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Odberateľ: %s", f.Odberatel))

	pdf.Ln(15)
	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(80, 10, "Položka")
	pdf.Cell(30, 10, "Množstvo")
	pdf.Cell(40, 10, "Cena za kus")
	pdf.Cell(40, 10, "Spolu")

	pdf.Ln(10)
	for _, p := range f.Polozky {
		pdf.Cell(80, 10, p.Nazov)
		pdf.Cell(30, 10, strconv.Itoa(p.Mnozstvo))
		pdf.Cell(40, 10, fmt.Sprintf("%.2f EUR", p.Cena))
		pdf.Cell(40, 10, fmt.Sprintf("%.2f EUR", float64(p.Mnozstvo)*p.Cena))
		pdf.Ln(8)
	}

	pdf.Ln(10)
	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(150, 10, "Celková suma:")
	pdf.Cell(40, 10, fmt.Sprintf("%.2f EUR", f.CelkovaSuma()))

	// Vytvorenie priečinka ~/Desktop/FAs, ak neexistuje
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("chyba pri získavaní domovského priečinka: %v", err)
	}

	faDir := filepath.Join(homeDir, "Desktop", "FAs")
	err = os.MkdirAll(faDir, 0755)
	if err != nil {
		return fmt.Errorf("chyba pri vytváraní priečinka: %v", err)
	}

	// Uloženie PDF súboru
	fileName := filepath.Join(faDir, fmt.Sprintf("%s.pdf", f.Cislo))
	err = pdf.OutputFileAndClose(fileName)
	if err != nil {
		return fmt.Errorf("chyba pri vytváraní PDF: %v", err)
	}

	return nil
}
