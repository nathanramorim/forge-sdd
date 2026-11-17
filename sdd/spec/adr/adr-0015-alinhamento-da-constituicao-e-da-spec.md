# ADR-0015: Alinhamento da Constituição e da Especificação Arquitetural com o Código

- **Data:** 2026-11-17
- **Status:** `proposto`
- **Regras afetadas (IDs da constitution):** `ARQ-06`, `SEG-03`, `DUP-01`, `GOV-01` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (comando `/constitution`)

## Contexto

Depois das iniciativas 88 a 90, a Constituição e `sdd/spec/` ficaram atrás do código:
- `ARQ-06` é o contrato dos subcomandos da CLI (Regra 10), mas não listava `archive`, `brief`, `ps` e `version`, que existem e são usados (`brief` é o primeiro passo de todo comando do Líder).
- `SEG-03` nomeava `git` e `go`, mas desde `forge pipeline pr` (ADR-0014 D-2) o Go também executa `gh`, com a lista de comandos permitidos em `RunGitOrGh` (`internal/pipeline/pr.go`).
- `sdd/spec/` estava na 4.7.7-beta: sem modo avançado, `/toca-ai`, PR aberto pelo Go, base da iniciativa e uso real por job; `decisions.md` mantinha D7 ("nenhum subcomando de runtime", falso desde a ADR-0004), D2 em cobra v1.8 (o `go.mod` tem v1.9.1) e o índice parava na ADR-0012; `overview.md` listava `review` como comando de raiz (é `forge bus review`); `modules.md` descrevia os checks `P-xx` do gate com ids trocados.
- O `forge brief` monta o contexto de todas as sessões a partir desses arquivos, então a defasagem chega a todo agente.

## Opções Consideradas

1. **Atualizar só a documentação:** deixa `ARQ-06` e `SEG-03` divergentes do código.
2. **Atualizar a Constituição e a especificação juntas, com esta ADR no mesmo PR (`GOV-01`).**
3. **Gerar `ARQ-06` a partir do `forge --help`:** elimina a deriva, mas pede um check novo no gate; fica como acompanhamento.

## Decisão

Opção 2.
- **D-1 (`ARQ-06`):** passa a listar também `archive`, `brief`, `ps` e `version`. Ficam de fora `help` e `completion`, gerados pelo cobra. A lista confere com `forge --help`.
- **D-2 (`SEG-03`):** os subprocessos passam a ser `git`, `gh` e `go`, sempre com argumentos explícitos.
- **D-3 (ajuste editorial, sem mudar regra):** o parágrafo "Roteamento para `.agents/rules/`" ficou mais curto e cita `forge context` (nome único do comando, ADR-0010), para a Constituição caber no limite de 9.216 bytes (ADR-0003, `P-24`); ela tem 9.183.
- **D-4 (`DUP-01`):** `sdd/spec/` volta a refletir o código até a iniciativa 90 (`feat/88`): módulos, fluxos, conceitos (cada conceito com seu dono), decisões D18 a D21, índice de ADRs completo, D7 marcada como substituída, D2 corrigida e D12 resolvida (ldflags e `config.BinaryVersion`).
- **D-5:** as ADRs 0009 a 0014 continuam `proposto`: promover é decisão do dono do produto (Camada C), depois de uso real.

## Consequências

- **Positivas:** o `forge brief` e as specs voltam a dizer o que o código faz; `ARQ-06` pode ser conferido contra `forge --help`.
- **Negativas / Trade-offs:** `ARQ-06` continua uma lista mantida à mão e volta a defasar a cada subcomando novo; a Constituição tem só 33 bytes livres no limite.
- **Ações de acompanhamento:** avaliar um check do gate que compare `ARQ-06` com os subcomandos reais; promover as ADRs 0009 a 0015 depois de uso real; atualizar o `.sddrc` (hoje em 4.5.0-beta), que `forge update --upgrade` reescreve junto das skills e por isso fica fora desta entrega.
