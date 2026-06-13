# feat/unified-onboarding-docs

**Branch:** `feat/unified-onboarding-docs`
**Fase:** 17
**Depende de:** `feat-14-discovery-command`, `feat-16-constitution-command`
**Status:** `todo`

## Objetivo

Unificar e simplificar a documentação de "onboarding" (Getting Started) em todos os READMEs do projeto. O objetivo é fornecer uma sequência lógica de comandos para dois cenários:
1. **Projeto do Zero:** Caminho desde o `init` até a primeira `feature`.
2. **Projeto Existente:** Como adotar o SDD em um codebase já consolidado.

## Critério de conclusão

- [ ] README.md principal contém as seções "Guia de Início Rápido" (Novos vs Existentes).
- [ ] npm/README.md reflete as mesmas instruções simplificadas.
- [ ] O template `sdd/README.md.tmpl` instrui o usuário sobre o ciclo de vida dos comandos.
- [ ] O arquivo `sdd/HOWTO.md` (template) é detalhado com exemplos reais.

## Tarefas

- [ ] **17-1** Definir a "Trilha de Sucesso" para Novos Projetos: `init` -> `/constitution` -> `/discovery` -> `/nova-feature` -> `/proxima-feature`.
- [ ] **17-2** Definir a "Trilha de Adoção" para Projetos Existentes: `init` -> `/constitution` (scan) -> `/status`.
- [ ] **17-3** Atualizar o `README.md` da raiz com fluxogramas ou listas numeradas dessas trilhas.
- [ ] **17-4** Sincronizar as mudanças no `npm/README.md`.
- [ ] **17-5** Atualizar o template `internal/scaffold/templates/sdd/README.md.tmpl`.
- [ ] **17-6** Criar/Atualizar o template `internal/scaffold/templates/sdd/HOWTO.md.tmpl` com o guia passo-a-passo.

## Fluxo Lógico (Proposto)

### 🆕 Novos Projetos
1. `forge-sdd init` -> Prepara o solo.
2. `/constitution` -> Define as regras e arquitetura base.
3. `/discovery "ideia"` -> Estrutura o produto e requisitos técnicos.
4. `/nova-feature` -> Cria as tarefas no roadmap.
5. `/proxima-feature` -> Inicia a execução.

### 🏗️ Projetos Existentes
1. `forge-sdd init` -> Adiciona a estrutura SDD.
2. `/constitution` -> O agente lê o codebase e "aprende" as regras atuais.
3. `/status` -> Verifica o progresso (inicialmente vazio).
4. `/nova-feature` -> Mapeia a próxima dívida técnica ou feature.

## Arquivos gerados/modificados

```
README.md
npm/README.md
internal/scaffold/templates/sdd/README.md.tmpl
internal/scaffold/templates/sdd/HOWTO.md.tmpl
```
