# feat-95-01 — Visualizador de artefatos da spec após conclusão

**Iniciativa:** 95-visualizador-discovery-apos-spec
**Branch:** feat/95-visualizador-discovery-apos-spec
**Status:** spec_pending
**Depende de:** fix-88-03 (deve ser mesclado em main antes desta branch — ambos editam `sdd/memory/clarify.md`, seu template e golden, gerando conflito se aplicados na ordem errada)

## Sintoma / Contexto

Ao concluir um `/discovery`, `/nova-feature`, `/novo-fix` ou `/split-features`, o agente PM+Arquiteto encerra sem oferecer ao usuário uma forma rápida de revisar os artefatos gerados. O usuário precisa navegar manualmente até `sdd/discovery/<id>-<nome>/` para abrir os arquivos, o que interrompe o fluxo — especialmente quando os artefatos contêm diagramas Mermaid ou imagens que só renderizam em um visualizador adequado.

A demanda é adicionar, imediatamente após a criação dos arquivos finais da spec e antes da submissão ao Gate 1, um `AskUserQuestion` com as opções: abrir a pasta no Finder, abrir no VS Code (que renderiza Mermaid com extensão), ver a lista dos arquivos existentes na pasta da iniciativa, ou pular.

**Interpretação de "naquele ask":** a demanda usa a expressão "naquele ask" para indicar o momento de interação ao final do fluxo — não um `AskUserQuestion` já existente, pois hoje nenhum existe no encerramento desses fluxos. O ask novo é inserido antes da submissão ao Gate 1 (ou antes da fila, em `/split-features`).

## Feature correlacionada

Iniciativa nova e isolada. A dependência em fix-88-03 é de merge-order: ambos editam `sdd/memory/clarify.md`, o template `clarify.md.tmpl` e o golden correspondente, gerando conflito se mesclados na ordem errada. Não é dependência funcional — o passo de visualização funciona independentemente da seção "Como perguntar" do fix-88-03.

## Modelo do módulo

- **Modelo:** enxuto.
- **Justificativa:** a entrega é exclusivamente em arquivos de skill (`.agents/skills/sdd-concepcao/references/`, `.agents/skills/sdd-planejamento/references/`) e nos seus templates de scaffold (`internal/scaffold/templates/`). Nenhum pacote Go novo é criado; nenhum binário é alterado. Não há domínio, fila, ACP nem configuração envolvidos (ARQ-11).

## Decisões do projeto aplicadas

- **ARQ-01:** sem toque em `cmd/`; mudança é só em arquivos de skill e templates embed (ARQ-03).
- **ARQ-03:** os templates `.tmpl` são a fonte canônica; o golden de scaffold deve espelhar (TST-02).
- **ARQ-04:** sem nova dependência de runtime Go; `open` e `code` são ferramentas de OS invocadas pelo agente, não pelo binário.
- **ARQ-06:** sem novo subcomando CLI — `forge open` não é criado; isso exigiria ADR por mudança de contrato (FLX-05 / GOV-01). Fora do escopo.
- **DUP-06:** a descrição de comportamento (opções, fallback, macOS-only) fica em `sdd/memory/clarify.md` uma única vez; os quatro arquivos de referência de comando apenas apontam para ela.
- **TST-02:** mudança em templates de scaffold atualiza goldens no mesmo PR.
- **FLX-01:** Spec aprovada antes de qualquer implementação.
- **FLX-02:** branch `feat/95-visualizador-discovery-apos-spec` criada a partir de `main`.
- **COD-05:** tudo em pt-BR.
- Sem ADR nova nem ADR depreciada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Pergunta pós-spec ao usuário | `.agents/skills/sdd-concepcao/references/discovery.md` (passos 10-11), `nova-feature.md` (passos 6-7), `novo-fix.md` (passos 6-7), `split-features.md` (passos 7-8) | Nenhum ask de visualização; os fluxos encerram direto no gate ou na telemetria | criar | Comportamento novo para todos os fluxos. |
| Regra AskUserQuestion nos skills | `.agents/skills/sdd-concepcao/SKILL.md` (live e template) e `sdd/memory/clarify.md` | Presente no branch `fix/88-05` (fix-88-03), ausente em `main` | criar | Passo de visualização é novo; a seção "Como perguntar" do fix-88-03 é independente. A dependência é de merge-order: ambos editam `clarify.md` e seu golden. |
| Abertura de pasta/arquivo por agente | Todos os skills em `.agents/skills/` | Nenhum skill abre pasta ou arquivo externo hoje | criar | Comportamento novo para todos os fluxos de spec. |
| Abertura de pasta no macOS | `open <dir>` (comando macOS) | Disponível como chamada de shell | reusar | Não requer nova dependência; o agente emite o comando. |
| Abertura no VS Code | `code <dir>` (VS Code CLI) | Disponível se VS Code estiver instalado e `code` no PATH | reusar com cautela | Verificar disponibilidade com `command -v code`; fallback para Finder se ausente. |
| Template de referências da concepcao | `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/` | Existem `discovery.md.tmpl`, `nova-feature.md.tmpl`, `novo-fix.md.tmpl`, sem passo de visualização | estender | As mudanças nos arquivos live devem espelhar nos `.tmpl`. |
| Template de referências do planejamento | `internal/scaffold/templates/.agents/skills/sdd-planejamento/references/split-features.md.tmpl` | Existe, sem passo de visualização | estender | Idem. |
| Template de clarify.md | `internal/scaffold/templates/sdd/memory/clarify.md.tmpl` | Existe (2334 bytes), sem seção "Visualização pós-spec" | estender | A seção canônica deve aparecer no template para ser scaffoldada. |
| Goldens de scaffold | `internal/scaffold/testdata/golden/` (sdd-concepcao references, sdd-planejamento references, `sdd/memory/clarify.md`) | Existem e espelham os templates atuais | estender | Regenerar com `-update` após mudar templates (TST-02). |

## Dependências novas

- **Nenhuma Go** (`ARQ-04`): `open` e `code` são ferramentas externas invocadas pelo agente em linguagem natural ou via shell do ambiente. O binário `forge-sdd` não as invoca.
- **VS Code (opcional, usuário):** para renderização Mermaid é necessária a extensão `bierner.markdown-mermaid` ("Markdown Preview Mermaid Support"). Não é dependência do projeto Go.
- **macOS-only:** `open` é nativo no macOS. Projetos que rodam em Linux ou Windows devem ignorar ou substituir a opção "Abrir no Finder" (ver Fora do escopo).

## Rules e IDs da Constituição Aplicáveis

- `ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-06`
- `DUP-02`, `DUP-06`
- `FLX-01`, `FLX-02`, `FLX-05`
- `GOV-01`, `GOV-04`
- `TST-02`
- `COD-05`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-fluxo-e-portao.md`, `.agents/rules/sdd-codigo-e-testes.md`

## Escopo

**Orçamento de leitura do /discovery:** 18.355 bytes usados de 22.528 bytes (limite). Headroom disponível: ~4.173 bytes. As mudanças em `discovery.md` e `clarify.md` adicionam ~200 bytes cada, dentro do orçamento.

### O que fazer, passo a passo

1. **`sdd/memory/clarify.md` — adicionar seção "Visualização pós-spec":**
   - Definir as 4 opções canônicas do `AskUserQuestion` de visualização:
     - `"Abrir pasta no Finder"` → o agente emite `open <diretório da iniciativa>`
     - `"Abrir no VS Code"` → verificar `command -v code`; se disponível, emitir `code <diretório da iniciativa>`; verificar `code --list-extensions | grep -qi mermaid` e informar se a extensão `bierner.markdown-mermaid` está ausente (sugerir `code --install-extension bierner.markdown-mermaid`) e indicar que Cmd+Shift+V abre a preview; se `code` não estiver no PATH, informar e sugerir o Finder
     - `"Ver lista de arquivos"` → listar todos os arquivos (`.md`, imagens etc.) existentes na pasta da iniciativa (`sdd/discovery/<id>-<nome>/`) na resposta do agente, incluindo subpastas `features/` e `fixes/`
     - `"Pular"` → continuar sem ação
   - Fallback quando `AskUserQuestion` não está disponível: listar os caminhos dos arquivos em texto plano e continuar.
   - Especificar que `open` é macOS-only; em ambientes não-macOS o agente omite a opção Finder e indica `xdg-open` como alternativa manual.

2. **`.agents/skills/sdd-concepcao/references/discovery.md` — adicionar passo 10a:**
   - Entre os passos 10 (telemetria) e 11 (Gate 1), inserir:
     > **10a. Visualização:** `AskUserQuestion` com as opções de `sdd/memory/clarify.md` (seção "Visualização pós-spec"). Execute a ação escolhida antes de continuar para o Gate 1.

3. **`.agents/skills/sdd-concepcao/references/nova-feature.md` — adicionar passo 6a:**
   - Após o passo 6 (Conteúdo, onde o arquivo da spec é criado) e antes do passo 7 (Gate e fila), inserir:
     > **6a. Visualização:** `AskUserQuestion` com as opções de `sdd/memory/clarify.md` (seção "Visualização pós-spec"). Execute a ação antes de submeter ao Gate.

4. **`.agents/skills/sdd-concepcao/references/novo-fix.md` — adicionar passo 6a:**
   - Após o passo 6 (Conteúdo, onde o arquivo do fix é criado) e antes do passo 7 (Telemetria), inserir:
     > **6a. Visualização:** `AskUserQuestion` com as opções de `sdd/memory/clarify.md` (seção "Visualização pós-spec"). Execute a ação antes de gravar telemetria.

5. **`.agents/skills/sdd-planejamento/references/split-features.md` — adicionar passo 7a:**
   - Após o passo 7 (Estado: `status.md` atualizado, onde as specs estão criadas) e antes do passo 8 (Fila), inserir:
     > **7a. Visualização:** `AskUserQuestion` com as opções de `sdd/memory/clarify.md` (seção "Visualização pós-spec"). Execute a ação antes de enfileirar as tasks.

6. **Templates de scaffold — espelhar as mesmas mudanças (passos 1–5):**
   - `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/discovery.md.tmpl`
   - `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/nova-feature.md.tmpl`
   - `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/novo-fix.md.tmpl`
   - `internal/scaffold/templates/.agents/skills/sdd-planejamento/references/split-features.md.tmpl`
   - `internal/scaffold/templates/sdd/memory/clarify.md.tmpl`

7. **Goldens de scaffold — atualizar para refletir os templates alterados:**
   - Executar `go test ./internal/scaffold/ -run TestGoldenInit -update` para regenerar; revisar o diff antes de commitar.
   - Arquivos afetados:
     - `internal/scaffold/testdata/golden/.agents/skills/sdd-concepcao/references/discovery.md`
     - `internal/scaffold/testdata/golden/.agents/skills/sdd-concepcao/references/nova-feature.md`
     - `internal/scaffold/testdata/golden/.agents/skills/sdd-concepcao/references/novo-fix.md`
     - `internal/scaffold/testdata/golden/.agents/skills/sdd-planejamento/references/split-features.md`
     - `internal/scaffold/testdata/golden/sdd/memory/clarify.md`

## Fora do escopo

- Criar subcomando `forge open` ou qualquer novo comando CLI (exigiria ADR por FLX-05 / GOV-01).
- Suporte a Linux e Windows (`open` não existe nessas plataformas; `xdg-open` ou equivalente é uma feature separada).
- Instalar extensões do VS Code automaticamente (somente detectar e informar).
- Modificar o `SKILL.md` principal da concepcao ou de qualquer estação (a regra de AskUserQuestion é de responsabilidade do fix-88-03).

## Critérios de Aceitação Executáveis

```bash
# 1. Os templates de scaffold contêm o novo passo de visualização
grep -c "Visualização" \
  internal/scaffold/templates/.agents/skills/sdd-concepcao/references/discovery.md.tmpl \
  internal/scaffold/templates/.agents/skills/sdd-concepcao/references/nova-feature.md.tmpl \
  internal/scaffold/templates/.agents/skills/sdd-concepcao/references/novo-fix.md.tmpl \
  internal/scaffold/templates/.agents/skills/sdd-planejamento/references/split-features.md.tmpl
# Esperado: cada arquivo retorna 1

# 2. Os arquivos live da skill contêm o novo passo
grep -c "Visualização" \
  .agents/skills/sdd-concepcao/references/discovery.md \
  .agents/skills/sdd-concepcao/references/nova-feature.md \
  .agents/skills/sdd-concepcao/references/novo-fix.md \
  .agents/skills/sdd-planejamento/references/split-features.md
# Esperado: cada arquivo retorna 1

# 3. clarify.md (live e template) contêm a seção canônica com as 4 opções
grep -c "Visualização pós-spec" sdd/memory/clarify.md
# Esperado: 1
grep -c "Visualização pós-spec" internal/scaffold/templates/sdd/memory/clarify.md.tmpl
# Esperado: 1
grep -c "bierner.markdown-mermaid" sdd/memory/clarify.md
# Esperado: 1

# 4. Goldens e orçamento de leitura passam sem erro
go test ./internal/scaffold/...
# Esperado: exit code 0 (TestGoldenInit, TestContextBudget e demais passam)

# 5. go vet limpo
go vet ./...
# Esperado: sem saída (exit code 0)
```
