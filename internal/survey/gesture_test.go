package survey

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
)

func gestureFixture(choice *string) *gestureModel {
	sel := huh.NewSelect[string]().Title("t").Options(huh.NewOption("A", "a"), huh.NewOption("B", "b")).Value(choice)
	m := &gestureModel{form: huh.NewForm(huh.NewGroup(sel))}
	m.Init()
	return m
}

func TestGestureModel_DireitaFixaAOpcaoSobOCursor(t *testing.T) {
	choice := "a"
	m := gestureFixture(&choice)
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, GestureRight, m.gesture)
	assert.Equal(t, "b", choice)
	assert.NotNil(t, cmd)
}

func TestGestureModel_EsquerdaEEnterSemGesto(t *testing.T) {
	choice := "a"
	m := gestureFixture(&choice)
	m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, GestureLeft, m.gesture)

	choice2 := "a"
	m2 := gestureFixture(&choice2)
	m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, GestureNone, m2.gesture)
	assert.Equal(t, "a", choice2)
}
