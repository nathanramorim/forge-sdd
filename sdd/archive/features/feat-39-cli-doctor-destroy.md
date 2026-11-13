# Especificação Técnica — Fase 39: Comandos doctor e destroy no CLI

Esta especificação define o design e o comportamento dos novos comandos `doctor` e `destroy` a serem incorporados ao CLI `forge-sdd`.

## 1. Comando `destroy`
O comando `destroy` é responsável por desinstalar e limpar completamente a estrutura do Forge-SDD do repositório/diretório local do usuário.

### Comportamento
- **Flags:**
  - `--yes`, `-y` (bool): Pula a confirmação interativa de segurança.
  - `--dry-run` (bool): Lista os arquivos que seriam removidos sem deletá-los de fato.
- **Fluxo:**
  1. Verifica se existe o arquivo `sdd/.sddrc`. Se não existir, reporta que nenhuma estrutura foi encontrada no diretório e finaliza com sucesso.
  2. Caso exista a estrutura, exibe um aviso de alerta informando que esta ação apagará permanentemente a pasta `sdd/` (especificações e progresso) e os arquivos/pastas de agentes.
  3. Solicita confirmação interativa com prompt (`huh.Confirm`).
  4. Se confirmado (ou se a flag `--yes` estiver ativa):
     - Deleta recursivamente os caminhos estruturais (se existirem):
       - `sdd/`
       - `CLAUDE.md` e `.claude/`
       - `GEMINI.md` e `.gemini/`
       - `.github/copilot-instructions.md`
       - `.vscode/mcp.json`
     - Imprime uma mensagem de confirmação com a lista dos arquivos e diretórios removidos.

---

## 2. Comando `doctor`
O comando `doctor` analisa a pasta local e diagnostica a saúde da instalação do Forge-SDD, reportando inconsistências ou o progresso geral.

### Comportamento
- **Fluxo de Análise:**
  1. **Detecção do SDD:**
     - Checa a presença de `sdd/.sddrc` e `sdd/.sdd-version`.
     - Se ausentes, reporta status `Não iniciado` e instrui o usuário a rodar `forge-sdd init`.
  2. **Validação de Agentes:**
     - Lê os agentes configurados no `.sddrc`.
     - Checa se os respectivos arquivos de cada agente configurado existem fisicamente (ex: pasta `.gemini` e `GEMINI.md`).
     - Alerta se houver agentes no `.sddrc` ausentes no disco ou vice-versa.
  3. **Verificação de Estrutura:**
     - Checa se `sdd/memory/progress.md` e `sdd/memory/constitution.md` existem.
  4. **Varredura de Features:**
     - Conta o número de especificações de feature no diretório `sdd/features/`.
     - Analisa o status do progresso lendo as features ativas.
  5. **Saída Formatada:**
     - Mostra um relatório amigável no terminal utilizando emojis e cores (se suportado pelo TTY) indicando a saúde de cada componente da metodologia.
