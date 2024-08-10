package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type invoiceCreatedMsg struct {
	path string
}

type model struct {
	inputs         []textinput.Model
	focused        int
	err            error
	quitting       bool
	invoiceCreated bool
	pdfPath        string
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.focused == len(m.inputs) {
				// Ak sme na tlačidle "Vytvoriť faktúru"
				return m, createInvoiceCmd(m.inputs)
			}
			m.nextInput()

		case "up", "shift+tab":
			m.prevInput()

		case "down", "tab":
			m.nextInput()

		case "ctrl+n":
			// Pridanie novej položky
			m.addNewItem()
			return m, nil
		}

	case invoiceCreatedMsg:
		m.invoiceCreated = true
		m.pdfPath = msg.path
		return m, tea.Sequence(
			func() tea.Msg {
				openPDF(m.pdfPath)
				return nil
			},
			tea.Quit,
		)

	case errMsg:
		m.err = msg
		return m, nil
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteRune('\n')
	}

	button := &blurredButton
	if m.focused == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n%s\n\n", *button)

	b.WriteString("\nStlačte Ctrl+N pre pridanie novej položky\n")

	if m.invoiceCreated {
		b.WriteString("\n")
		b.WriteString(successStyle.Render("Faktúra bola úspešne vytvorená!"))
		b.WriteString("\n")
		b.WriteString(successStyle.Render(fmt.Sprintf("PDF súbor bol uložený: %s", m.pdfPath)))
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	return b.String()
}

func (m *model) nextInput() {
	m.focused = (m.focused + 1) % (len(m.inputs) + 1)
	m.updateFocus()
}

func (m *model) prevInput() {
	m.focused = (m.focused - 1 + len(m.inputs) + 1) % (len(m.inputs) + 1)
	m.updateFocus()
}

func (m *model) updateFocus() {
	for i := range m.inputs {
		if i == m.focused {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = noStyle
			m.inputs[i].TextStyle = noStyle
		}
	}
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *model) addNewItem() {
	newInput := textinput.New()
	newInput.Placeholder = fmt.Sprintf("Položka %d (názov,množstvo,cena)", len(m.inputs)-1)
	newInput.CharLimit = 64
	m.inputs = append(m.inputs, newInput)
	m.focused = len(m.inputs) - 1 // Nastavíme focus na novo pridanú položku
	m.updateFocus()
}

func initialModel() model {
	var inputs []textinput.Model
	inputs = make([]textinput.Model, 3) // Začíname s 1 položkou + číslo faktúry a odberateľ

	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 64

		switch i {
		case 0:
			t.Placeholder = "Číslo faktúry"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Odberateľ"
		case 2:
			t.Placeholder = "Položka 1 (názov,množstvo,cena)"
		}

		inputs[i] = t
	}

	return model{
		inputs:  inputs,
		focused: 0,
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Chyba: %v", err)
		return
	}
}

func createInvoiceCmd(inputs []textinput.Model) tea.Cmd {
	return func() tea.Msg {
		cislo := inputs[0].Value()
		odberatel := inputs[1].Value()

		var polozky []Polozka
		for i := 2; i < len(inputs); i++ {
			polozka := parsePolozka(inputs[i].Value())
			if polozka.Nazov != "" {
				polozky = append(polozky, polozka)
			}
		}

		faktura := Faktura{
			Cislo:           cislo,
			DatumVystavenia: time.Now(),
			DatumSplatnosti: time.Now().AddDate(0, 0, 14),
			Odberatel:       odberatel,
			Polozky:         polozky,
		}

		err := generatePDF(faktura)
		if err != nil {
			return errMsg(err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return errMsg(fmt.Errorf("chyba pri získavaní domovského priečinka: %v", err))
		}

		pdfPath := filepath.Join(homeDir, "Desktop", "Fargo", "FAs", fmt.Sprintf("%s.pdf", faktura.Cislo))

		return invoiceCreatedMsg{path: pdfPath}
	}
}

func parsePolozka(input string) Polozka {
	parts := strings.Split(input, ",")
	if len(parts) != 3 {
		return Polozka{}
	}

	nazov := strings.TrimSpace(parts[0])
	mnozstvo, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	cena, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)

	return Polozka{
		Nazov:    nazov,
		Mnozstvo: mnozstvo,
		Cena:     cena,
	}
}

func openPDF(path string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // Linux a ostatné Unix-like systémy
		cmd = exec.Command("xdg-open", path)
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Chyba pri otváraní PDF: %v\n", err)
	}
}
