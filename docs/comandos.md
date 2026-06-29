# Referência de Comandos

O ecossistema **forge-sdd** fornece um conjunto de ferramentas divididas entre **Comandos de Terminal (CLI)** e **Comandos de Chat (Agentes de IA)** para guiar o ciclo de desenvolvimento.

---

## 💻 1. Comandos de Terminal (CLI / NPX)
Executados diretamente no terminal do desenvolvedor, normalmente utilizando o wrapper NPX.

### `npx @nathanramorim/forge-sdd init [diretório]`
- **O que faz:** Inicializa toda a estrutura de pastas do Forge-SDD (diretórios `sdd/`, `.vscode/`, `.gemini/`, `.github/`, etc.) adaptada para os agentes de IA escolhidos.
- **Como usar:**
  - **Interativo (Recomendado):**
    ```bash
    npx @nathanramorim/forge-sdd init
    ```
  - **Automático (Não-Interativo via flags):**
    ```bash
    npx @nathanramorim/forge-sdd init --yes --name "Meu Projeto" --stack "go" --db "postgres" --agent "gemini,claude"
    ```
  - **Simulação (Dry-run):**
    ```bash
    npx @nathanramorim/forge-sdd init --dry-run
    ```

### `npx @nathanramorim/forge-sdd update [diretório]`
- **O que faz:** Atualiza as regras estruturais de agentes de um projeto existente ou adiciona novos agentes de IA (ex: adicionar Claude a um projeto que só usava Gemini), preservando estritamente os dados de domínio (features, memory, etc.).
- **Segurança (Preservação de Domínio):** Para evitar perda de dados, o comando atualiza apenas arquivos estruturais e de agentes (como `.gemini/`, `.claude/`, `CLAUDE.md`, `GEMINI.md`, `sdd/.sddrc` e `sdd/.sdd-version`). Todo o domínio do projeto sob a pasta `sdd/` (incluindo suas especificações de features, histórico de releases e arquivos de memória como `progress.md`) é estritamente preservado.
- **Como usar:**
  - **Interativo:**
    ```bash
    npx @nathanramorim/forge-sdd update
    ```
  - **Automático (Não-Interativo via flags):**
    ```bash
    npx @nathanramorim/forge-sdd update --yes --upgrade --agent "gemini"
    ```
    - **Flags Disponíveis:**
      - `--yes`: Pula os prompts e executa a ação imediatamente.
      - `--agent <lista,de,agentes>`: Adiciona novos agentes (ex: `claude,gemini,openai`).
      - `--upgrade`: Atualiza a estrutura para a versão mais recente do CLI.
      - `--version <versao>`: Atualiza a estrutura para uma versão específica (ex: `1.6.0`).

### `npx @nathanramorim/forge-sdd version`
- **O que faz:** Exibe a versão instalada do CLI do forge-sdd.
- **Como usar:**
  ```bash
  npx @nathanramorim/forge-sdd version
  ```

---

## 🤖 2. Comandos de Chat (Agentes de IA / Chatmode)
Digitados na interface de chat do agente de IA configurado no projeto (ex: Gemini, Claude Code, GitHub Copilot Chat). Eles disparam prompts padronizados sob a pasta `.gemini/prompts/` (ou correspondente ao agente configurado).

### `/status` ou `"qual o progresso?"`
- **Quando usar:** No início de qualquer sessão de trabalho para que o agente entenda o contexto atual (comando obrigatório de inicialização).
- **O que faz:** Lê o arquivo `sdd/memory/progress.md` e resume o status das features ativas, próximos passos, bloqueios e percentual de conclusão do projeto.
- **Exemplo de uso:** 
  > `/status`

### `/discovery <ideia>` ou `"fazer discovery de..."`
- **Quando usar:** Quando você tem uma ideia ou demanda vaga e precisa debater arquitetura, especificações técnicas e viabilidade.
- **O que faz:** Inicia uma sessão de descoberta com duas personas virtuais da IA (PM + Engenheiro Sênior) e gera especificações preliminares em `sdd/discovery/`.
- **Exemplo de uso:** 
  > `/discovery Criar um sistema de notificações por e-mail`

### `/nova-feature <nome>` ou `"criar nova feature..."`
- **Quando usar:** Para registrar uma evolução específica ou requisito novo já mapeado e definido.
- **O que faz:** Cria a branch local correspondente e gera o arquivo de especificação inicial em `sdd/features/feat-XX-nome.md`.
- **Exemplo de uso:** 
  > `/nova-feature login-oauth`

### `/split-features`
- **Quando usar:** Se um plano de discovery ou tarefa se tornou complexo demais (geralmente contendo mais de 7 sub-tarefas).
- **O que faz:** Quebra um arquivo de plano em múltiplas features menores e independentes no backlog.
- **Exemplo de uso:** 
  > `/split-features`

### `/proxima-feature` ou `"iniciar próxima tarefa"`
- **Quando usar:** Quando estiver pronto para começar a programar a próxima tarefa pendente.
- **O que faz:** Analisa o backlog, seleciona a próxima feature pendente (`status: todo`), faz o checkout na branch correta e prepara o ambiente para o agente codificar.
- **Exemplo de uso:** 
  > `/proxima-feature`

### `/revisar` ou `"validar feature"`
- **Quando usar:** Antes de abrir um Pull Request ou fazer o merge de uma feature.
- **O que faz:** Executa os testes de integração, verificações de lint e auditoria geral para garantir a conformidade técnica da implementação atual.
- **Exemplo de uso:** 
  > `/revisar`

### `/constitution` ou `"alinhar arquitetura"`
- **Quando usar:** No início do projeto ou após mudanças drásticas de arquitetura/dependências.
- **O que faz:** Faz uma varredura completa do repositório para mapear as dependências e tecnologias usadas, alinhando a constituição (`sdd/memory/constitution.md`) do projeto.
- **Exemplo de uso:** 
  > `/constitution`

### `/c4-architecture` ou `"desenhar arquitetura"`
- **Quando usar:** Para documentar visualmente a topologia e fluxos do projeto.
- **O que faz:** Desenha diagramas de sistema de níveis 1 e 2 (Contexto e Container) nos padrões C4 Model utilizando Mermaid.
- **Exemplo de uso:** 
  > `/c4-architecture`

### `/doctor` ou `"check-up do projeto"`
- **Quando usar:** Se notar falhas de carregamento, conflito de MCP ou arquivos corrompidos.
- **O que faz:** Verifica a integridade estrutural e a compatibilidade dos arquivos e pastas do Forge-SDD no repositório.
- **Exemplo de uso:** 
  > `/doctor`

### `/archive` ou `"limpar progresso"`
- **Quando usar:** Quando o arquivo `sdd/memory/progress.md` exceder o limite recomendado de 1 KB.
- **O que faz:** Limpa e move as features finalizadas de `progress.md` para o histórico `progress-log.md`, mantendo o progresso leve.
- **Exemplo de uso:** 
  > `/archive`

### `/install-skill`
- **Quando usar:** Para estender as habilidades da IA com novos scripts ou padrões.
- **O que faz:** Importa uma pasta de Skill diretamente de uma URL externa (ex: repositório GitHub) para `.gemini/skills/`.
- **Exemplo de uso:** 
  > `/install-skill`

### `/upgrade-sdd <versao>` ou `"atualizar sdd para vX"`
- **Quando usar:** Quando uma nova versão do framework forge-sdd for lançada.
- **O que faz:** Inicia a rotina do **Migrator** para atualizar a metodologia no projeto para a versão indicada.
- **Exemplo de uso:** 
  > `/upgrade-sdd 1.6.0`
