package survey

import (
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

// ErrBack indica que a pessoa pressionou Esc (ou Ctrl+C) para voltar um nível no menu.
var ErrBack = errors.New("voltar")

// Option é uma linha de um seletor: o texto mostrado e o valor devolvido.
type Option struct {
	Label string
	Value string
}

func menuKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "voltar"))
	return km
}

// SelectOne mostra uma lista navegável (setas e Enter) e devolve o valor escolhido. def marca a
// opção inicial. Esc devolve ErrBack.
func SelectOne(title string, opts []Option, def string) (string, error) {
	choice := def
	hopts := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		hopts[i] = huh.NewOption(o.Label, o.Value)
	}
	sel := huh.NewSelect[string]().Title(title).Options(hopts...).Value(&choice)
	if len(opts) > 0 {
		sel = sel.Height(min(len(opts)+2, 16))
	}
	err := huh.NewForm(huh.NewGroup(sel)).WithKeyMap(menuKeyMap()).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", ErrBack
	}
	return choice, err
}

// ConfirmYesNo pergunta Sim/Não; Esc devolve ErrBack.
func ConfirmYesNo(title string, def bool) (bool, error) {
	v := def
	err := huh.NewForm(huh.NewGroup(huh.NewConfirm().Title(title).Affirmative("Sim").Negative("Não").Value(&v))).
		WithKeyMap(menuKeyMap()).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return false, ErrBack
	}
	return v, err
}

// Pause espera Enter antes de voltar ao menu, para a saída de uma ação não sumir da tela.
func Pause(msg string) error {
	ok := true
	err := huh.NewForm(huh.NewGroup(huh.NewConfirm().Title(msg).Affirmative("Voltar ao menu").Negative("Voltar ao menu").Value(&ok))).
		WithKeyMap(menuKeyMap()).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return nil
	}
	return err
}
