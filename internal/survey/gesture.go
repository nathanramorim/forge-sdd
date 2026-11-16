package survey

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// Gestos devolvidos por SelectGesture quando a pessoa aperta → ou ← sobre uma opção.
const (
	GestureNone  = ""
	GestureRight = "right"
	GestureLeft  = "left"
)

// gestureModel envolve o form do huh e captura →/← sem atrapalhar as demais teclas.
type gestureModel struct {
	form    *huh.Form
	gesture string
}

func (m *gestureModel) Init() tea.Cmd { return m.form.Init() }

func (m *gestureModel) View() string { return m.form.View() }

func (m *gestureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.Type {
		case tea.KeyRight, tea.KeyLeft:
			m.gesture = GestureRight
			if k.Type == tea.KeyLeft {
				m.gesture = GestureLeft
			}
			// O huh só grava o valor ao confirmar; um Enter sintético fixa a opção sob o cursor.
			model, _ := m.form.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if f, ok := model.(*huh.Form); ok {
				m.form = f
			}
			return m, tea.Quit
		}
	}
	model, cmd := m.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State != huh.StateNormal {
		return m, tea.Quit
	}
	return m, cmd
}

// SelectGesture é o SelectOne que também captura → e ←: devolve o valor sob o cursor e o gesto.
// Enter devolve o valor sem gesto; Esc devolve ErrBack.
func SelectGesture(title string, opts []Option, def string) (choice, gesture string, err error) {
	choice = def
	hopts := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		hopts[i] = huh.NewOption(o.Label, o.Value)
	}
	sel := huh.NewSelect[string]().Title(title).Options(hopts...).Value(&choice)
	if len(opts) > 0 {
		sel = sel.Height(min(len(opts)+2, 16))
	}
	form := huh.NewForm(huh.NewGroup(sel)).WithKeyMap(menuKeyMap())
	final, err := tea.NewProgram(&gestureModel{form: form}).Run()
	if err != nil {
		return "", GestureNone, err
	}
	m := final.(*gestureModel)
	if m.gesture == GestureNone && m.form.State == huh.StateAborted {
		return "", GestureNone, ErrBack
	}
	if m.gesture == GestureNone && m.form.State != huh.StateCompleted {
		return "", GestureNone, errors.New("seleção interrompida")
	}
	return choice, m.gesture, nil
}
