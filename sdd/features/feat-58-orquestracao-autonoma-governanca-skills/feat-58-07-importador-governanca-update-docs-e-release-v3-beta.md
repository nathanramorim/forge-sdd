# feat-58-07 — Importador de Pacotes de Governança (`manifest.yaml`), Migração (`update`), Documentação e Release `v3.0.0-beta`

**Branch:** `feat/58-orquestracao-autonoma-governanca-skills`
**Fase:** 58-07
**Depende de:** `feat-58-01`, `feat-58-02`, `feat-58-03`, `feat-58-04`, `feat-58-05`, `feat-58-06`
**Status:** `done`

## Objetivo

Concluir o pacote do Discovery 58 adicionando o suporte à importação idempotente de pacotes externos de governança baseados em `manifest.yaml` (padrão `planning-brain-governanca`), atualizando `forge-sdd update` para migrar projetos existentes sem perda de domínio, sincronizando toda a documentação do repositório (`sdd/FLOW.md`, `README.md`, `sdd/releases/history.md`) e publicando a release **`v3.0.0-beta`** diretamente na branch da feature com o PR aberto no GitHub (Regras 11 e 12).

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `completo` (migração no `internal/scaffold`, importador de manifesto de governança, documentação e publicação de release beta).
- **Justificativa:** Coordena a atualização não-destrutiva de arquivos de governança (`bloco`, `substituir`, `criar_se_ausente`, `mesclar`, `criar_uma_vez`) e os artefatos de release `v3.0.0-beta`.

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Migração em `forge-sdd update` | `internal/scaffold/` (`feat-22`, `feat-31`, `feat-02-04`, `feat-03-02`) | Preserva arquivos de domínio do usuário e migra estruturas antigas | `estender` | Garantir que `forge-sdd update` adicione `.agents/roles/`, `.agents/rules/sdd-*.md` e `.agents/skills/` preservando rules e blocos customizados do usuário |
| Manifesto de importação | `busize-solutions/planning-brain-governanca` (`manifest.yaml` e `IMPORTAR.md`) | Modos `bloco` (entre marcadores na constitution), `substituir`, `criar_se_ausente`, `mesclar`, `criar_uma_vez` + `docs/governanca/VERSAO` | `reusar` | Suportar aplicação determinística de blocos demarcados na `constitution.md` e comando/prompt de importação de governança |
| Documentação do fluxo e releases | `sdd/FLOW.md`, `README.md`, `npm/README.md`, `sdd/releases/history.md` | Documentação atual na `v2.4.0` | `estender` | Atualizar com a arquitetura de 5 estações, Portão A/B/C, Fila, Auto-Skills e notas da `v3.0.0-beta` (Regra 12) |

## Dependências novas (`SEG-07`)
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-fluxo-e-portao.md`
- **IDs:** `GOV-01`, `GOV-04`, `FLX-06..08`, Regra 11, Regra 12, Regra 14.

## Critérios de Aceitação Executáveis

1. `forge-sdd update` migra projetos existentes adicionando `.agents/roles/`, `.agents/rules/sdd-*.md`, `.agents/skills/` e os novos adaptadores de agentes, preservando intactos quaisquer arquivos de regra customizados pelo usuário em `.agents/rules/` e qualquer bloco `<!-- *-governanca:inicio --> ... <!-- *-governanca:fim -->` já presente em `sdd/memory/constitution.md`.
2. `sdd/discovery/discovery-53-multi-sessao-esteira-feature.md` e `sdd/discovery/discovery-56-jev-despachante-e-operacao-por-sessao.md` marcados no topo com nota de status `Superseded by discovery-58`.
3. `sdd/FLOW.md`, `README.md` e `npm/README.md` atualizados descrevendo as 5 estações (Líder, PM/Arquiteto, Planejador, Fila do Builder, Analista de Governança), o Portão A/B/C, a Constituição por IDs e os novos subcomandos CLI (`forge-sdd queue`, `forge-sdd gate`, `forge-sdd skills`).
4. `sdd/releases/history.md` atualizado com a seção `### Versão 3.0.0-beta` contendo bullets reais das melhorias, prévias dos comandos CLI e identificação explícita dos agentes nos comandos de chat (Regra 12).
5. `go test ./...`, `go vet ./...`, `forge-sdd gate` e `forge-sdd doctor` passam sem erros.
6. Pull Request da branch `feat/58-orquestracao-autonoma-governanca-skills` aberto no GitHub (mantido aberto sem mergear na `main`, conforme Regra 11) e tag `v3.0.0-beta` criada diretamente na branch da feature.

## Handoff
Fecha o pacote `feat-58-orquestracao-autonoma-governanca-skills` e publica a `v3.0.0-beta`.
