# ADR-0013: Reviewer Revisa a Spec no Modo Avançado

- **Data:** 2026-11-16
- **Status:** `proposto`
- **Regras afetadas (IDs da constitution):** `FLX-01`, `FLX-05`, `FLX-06`, `GOV-01`, `SEG-06` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (iniciativa `88-modo-pipeline-avancado`, `feat-88-03`)

## Contexto

Hoje o Gate 1 (aprovação da spec) é humano (`FLX-01`): a Discovery termina, a pessoa aprova e só então o Builder recebe a fila. Em um ambiente isolado e descartável, quem quer rodar o pipeline sem acompanhar cada passo fica preso a essa pausa. O Reviewer já existe como sessão própria, com contexto limpo e somente leitura (`SEG-06`, ADR-0012), e já revisa o PR com um checklist.

## Opções Consideradas

1. **Manter o humano como único revisor da spec:** seguro, mas impede a esteira autônoma.
2. **A própria Discovery se autoaprova:** elimina a pausa, mas quem escreve a spec não pode ser quem a aprova.
3. **O Reviewer revisa a spec no modo `advanced`:** sessão independente, mesmo checklist do PR, decisão aplicada pelo Go.

## Decisão

- **D-1:** com `autonomy: advanced` em `forge.json`, o Reviewer substitui a aprovação humana do Gate 1. No modo `normal` (padrão) o humano continua obrigatório.
- **D-2:** a Camada A (`forge gate`, `FLX-06`) é obrigatória e roda antes: só um gate sem bloqueio cria o job `spec_review`.
- **D-3:** o Reviewer devolve o parecer em `queue.Verdict` (`conforme` = aprovado, `violacao` = ajustar, `duvida` = escalar). Quem aplica e grava `review-spec.md` é o Go; o Reviewer continua somente leitura (`SEG-06`).
- **D-4:** `aprovado` libera a fila (job `feature_build`); `ajustar` devolve o feedback à Discovery até `reviewer.max_rework_attempts`; `escalar` ou o limite marcam as tasks como `blocked`, avisam pelo inbox e deixam as tasks independentes seguirem.
- **D-5:** o modo exige `permissions: full` e a confirmação de ambiente isolado; há tetos de tokens e tempo por job, teto diário de jobs e o arquivo `.forge/runtime/STOP`.
- **D-6 (`FLX-01`, `FLX-05`, `GOV-01`):** esta ADR altera quem pode ser o revisor da spec apenas quando o modo avançado está ligado; o texto de `FLX-01` não muda.

## Consequências

- **Positivas:** esteira sem pausa humana em ambiente isolado; revisão com contexto limpo e checklist único para spec e entrega.
- **Negativas / Trade-offs:** risco de o mesmo modelo escrever e revisar a spec; mitigação recomendada: usar agentes diferentes na Discovery e no Reviewer. O modo é experimental.
- **Ações de acompanhamento:** `forge doctor` avisa quando `advanced` está ativo sem ambiente isolado confirmado; promover a ADR a `aprovado` após uso real.
