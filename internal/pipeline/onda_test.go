package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContinuarAgrupadoPorOnda(t *testing.T) {
	plano := [][]Task{
		{{ID: "feat-1-01"}, {ID: "feat-1-02"}},
		{{ID: "feat-1-03", DependsOn: []string{"feat-1-01"}}},
	}

	// Primeira onda mesclada: a segunda pode começar, porque a predecessora está na principal.
	estados := map[string]string{
		"feat-1-01": TaskMesclada,
		"feat-1-02": TaskMesclada,
	}
	res, fechada, err := ContinuarPorOnda(plano, estados)
	require.NoError(t, err)
	assert.False(t, fechada, "a onda 2 ainda não fechou")
	assert.ElementsMatch(t, []string{"feat-1-01", "feat-1-02"}, res.Mescladas)
	assert.Equal(t, []string{"feat-1-03"}, res.PodeComecar)

	// O texto agrupa por onda, não por task.
	texto := TextoDoContinuar(res, 2)
	assert.Contains(t, texto, "Onda 2")
	assert.Contains(t, texto, "Podem começar")
	assert.Contains(t, texto, "[Continuar]")
	assert.Contains(t, texto, "[Ajustar]")

	// Com tudo mesclado, a onda fecha.
	todos := map[string]string{
		"feat-1-01": TaskMesclada, "feat-1-02": TaskMesclada, "feat-1-03": TaskMesclada,
	}
	_, fechada, err = ContinuarPorOnda(plano, todos)
	require.NoError(t, err)
	assert.True(t, fechada)
}

func TestDependenteNaoComecaAntesDaPredecessoraMesclada(t *testing.T) {
	plano := [][]Task{
		{{ID: "feat-1-01"}},
		{{ID: "feat-1-02", DependsOn: []string{"feat-1-01"}}},
	}

	// Predecessora em andamento (PR aberto, não mesclado): a dependente NÃO começa.
	for _, estado := range []string{OndaEmAndamento, TaskPendente, OndaBloqueada} {
		estados := map[string]string{"feat-1-01": estado}
		res, _, err := ContinuarPorOnda(plano, estados)
		require.NoError(t, err)
		assert.NotContains(t, res.PodeComecar, "feat-1-02",
			"com a predecessora em %q a dependente não pode começar", estado)
		assert.Contains(t, res.Aguardando, "feat-1-02")
	}

	// PodeComecar recusa diretamente, com o motivo e o estado da predecessora.
	tarefa := Task{ID: "feat-1-02", DependsOn: []string{"feat-1-01"}}
	ok, err := PodeComecar(tarefa, map[string]string{"feat-1-01": OndaEmAndamento})
	require.ErrorIs(t, err, ErrDependenteSemPredecessora)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "em-andamento")

	// Dependência fora do quadro da onda também recusa.
	_, err = PodeComecar(tarefa, map[string]string{})
	require.ErrorIs(t, err, ErrDependenteSemPredecessora)

	// Estado desconhecido recusa em vez de assumir que está tudo bem.
	_, err = PodeComecar(tarefa, map[string]string{"feat-1-01": "pronto-?"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "desconhecido")

	// Só mesclada libera.
	ok, err = PodeComecar(tarefa, map[string]string{"feat-1-01": TaskMesclada})
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestFimDaOnda(t *testing.T) {
	// Em andamento ou pendente seguram a onda.
	fim, err := FimDaOnda([]EstadoTaskNaOnda{{ID: "a", Estado: OndaEmAndamento}})
	require.NoError(t, err)
	assert.False(t, fim)

	fim, err = FimDaOnda([]EstadoTaskNaOnda{{ID: "a", Estado: TaskPendente}})
	require.NoError(t, err)
	assert.False(t, fim)

	// Mescladas e bloqueadas fecham a onda (bloqueada pede decisão, mas não trava a onda).
	fim, err = FimDaOnda([]EstadoTaskNaOnda{
		{ID: "a", Estado: TaskMesclada},
		{ID: "b", Estado: OndaBloqueada},
	})
	require.NoError(t, err)
	assert.True(t, fim)

	// Onda vazia é erro; estado inválido é erro.
	_, err = FimDaOnda(nil)
	require.Error(t, err)
	_, err = FimDaOnda([]EstadoTaskNaOnda{{ID: "a", Estado: "talvez"}})
	require.Error(t, err)
}

func TestEsperaEntreOndas(t *testing.T) {
	plano := [][]Task{
		{{ID: "feat-1-01"}},
		{{ID: "feat-1-02"}},
	}

	// A onda 1 ainda não fechou: a onda 2 não começa, mesmo sem dependência declarada.
	res, fechada, err := ContinuarPorOnda(plano, map[string]string{"feat-1-01": OndaEmAndamento})
	require.NoError(t, err)
	assert.False(t, fechada)
	assert.Empty(t, res.PodeComecar, "nada da próxima onda começa antes da anterior fechar")
	assert.Contains(t, res.Aguardando, "feat-1-02")
}

func TestTextoDoContinuarNomeiaBloqueadas(t *testing.T) {
	res := ResumoDaOnda{
		Mescladas:  []string{"feat-1-01"},
		Bloqueadas: []string{"feat-1-02"},
	}
	texto := TextoDoContinuar(res, 1)
	assert.Contains(t, texto, "BLOQUEADAS")
	assert.Contains(t, texto, "precisam de decisão sua",
		"bloqueada não pode passar por aprovada no texto do Continuar")
}
