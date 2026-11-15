---
description: Como cumprir as diretrizes de arquitetura, separação de pacotes, templates embutidos e declaração de modelo de módulo.
globs: ["cmd/**", "internal/**", "sdd/spec/**"]
regras: [ARQ-01, ARQ-02, ARQ-03, ARQ-04, ARQ-05, ARQ-06, ARQ-11]
---

# Guia Prático — Arquitetura e Fronteiras de Módulos

Instruções operacionais para aplicar os IDs `ARQ-01` a `ARQ-06` e `ARQ-11` definidos em `sdd/memory/constitution.md` (sem reproduzir o texto normativo da Constituição, conforme `GOV-04`).

## 1. Organização entre `cmd/` e `internal/` (`ARQ-01`, `ARQ-02`, `ARQ-06`)

- **Camada de entrada (`cmd/forge-sdd/`):**
  - Deve conter apenas a configuração de comandos/flags Cobra, leitura de argumentos, chamada aos pacotes de `internal/` e formatação da saída para o terminal.
  - Se um handler em `cmd/` precisar de mais de uma validação condicional de negócio ou manipular estruturas de estado (`queue.json`, `.sddrc`, métricas), extraia essa lógica para uma função ou serviço dentro de `internal/<pacote>/`.
- **Direção de dependências (`ARQ-02`):**
  - Pacotes em `internal/` podem importar pacotes utilitários/configuração de nível inferior (como `internal/config`), mas nunca `cmd/` nem criar dependências circulares entre si.
  - Ao expor um novo subcomando CLI (`ARQ-06`), documente suas flags no `--help` do comando e mantenha compatibilidade com os fluxos existentes de `init` e `update`.

## 2. Recursos Embutidos, Runtime Enxuto e `--dry-run` (`ARQ-03`, `ARQ-04`, `ARQ-05`)

- **Templates via `embed.FS` (`ARQ-03`):**
  - Todo arquivo que o scaffold materializa no projeto do usuário deve residir em `internal/scaffold/templates/` com sufixo `.tmpl`.
  - Nunca utilize `os.ReadFile` apontando para caminhos relativos do repositório do próprio `forge-sdd` para buscar templates em tempo de execução; leia sempre da instância `embed.FS`.
- **Preferência por `stdlib` (`ARQ-04`):**
  - Resolva parsing de JSON, manipulação de caminhos (`filepath`), expressões regulares e execução de processos com pacotes nativos do Go antes de cogitar bibliotecas externas.
- **Execução segura em `--dry-run` (`ARQ-05`):**
  - Antes de qualquer `os.MkdirAll`, `os.WriteFile`, `os.Rename` ou `os.Remove`, verifique a flag `DryRun`. Quando ativa, apenas registre ou imprima a operação planejada e retorne sem tocar no disco.

## 3. Declaração de Modelo na Spec (`ARQ-11`)

- Ao redigir `criteria-ID-*.md` ou `feat-XX-*.md`, inclua a seção `## Modelo do módulo`:
  - Use **`enxuto`** quando a entrega alterar apenas templates estáticos, prompts/rules Markdown ou ajustes pontuais sem novo estado de domínio.
  - Use **`completo`** quando a entrega introduzir ou modificar regras de negócio, máquinas de estado (ex.: fila, portão, sincronizador de skills) ou persistência estruturada em `internal/`.
