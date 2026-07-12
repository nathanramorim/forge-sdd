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

### `npx @nathanramorim/forge-sdd doctor [diretório]`
- **O que faz:** Diagnostica a saúde e a integridade da metodologia instalada no repositório. Verifica a existência de arquivos estruturais, valida se as pastas e integrações físicas dos agentes de IA estão de acordo com o `.sddrc`, lista as features em andamento e detecta deriva de convenção de nomenclatura (mistura de `feat-NN` sequencial com `feat-xxxx` por hash em `sdd/features/` e `sdd/discovery/`).
- **Como usar:**
  ```bash
  npx @nathanramorim/forge-sdd doctor
  ```

### `npx @nathanramorim/forge-sdd autopilot [diretório]`
- **O que faz:** Ativa o modo autopilot criando `sdd/.sdd-auto-pilot`, mas só libera a criação depois que o projeto acumular um número mínimo de ciclos completos (`outcome: approved`) registrados em `sdd/.metrics/` — preservando a função didática da metodologia para quem ainda está aprendendo. Idempotente se o arquivo já existir.
- **Como usar:**
  ```bash
  npx @nathanramorim/forge-sdd autopilot
  ```
  - **Flags Disponíveis:**
    - `--min-cycles <N>`: Número mínimo de ciclos completos exigidos (default: `3`).
    - `--skip-graduation`: Bypass consciente — ativa o autopilot mesmo sem os ciclos mínimos completos.

### `npx @nathanramorim/forge-sdd destroy [diretório]`
- **O que faz:** Desinstala e remove completamente toda a estrutura do Forge-SDD e integrações de agentes de IA locais (como a pasta `sdd/`, `.claude/`, `.gemini/`, `CLAUDE.md`, etc.). Requer confirmação interativa de segurança, a menos que a flag `--yes` esteja presente.
- **Como usar:**
  - **Confirmando interativamente:**
    ```bash
    npx @nathanramorim/forge-sdd destroy
    ```
  - **Remoção Direta:**
    ```bash
    npx @nathanramorim/forge-sdd destroy --yes
    ```
  - **Simulação (Dry-run):**
    ```bash
    npx @nathanramorim/forge-sdd destroy --dry-run
    ```

---

## 🤖 2. Comandos de Chat (Agentes de IA / Chatmode)
Digitados na interface de chat do agente de IA configurado no projeto (ex: Gemini, Claude Code, GitHub Copilot Chat). Eles disparam prompts padronizados sob a pasta `.gemini/prompts/` (ou correspondente ao agente configurado).

### `/status` ou `"qual o progresso?"`
- **Quando usar:** No início de qualquer sessão de trabalho para que o agente entenda o contexto atual (comando obrigatório de inicialização).
- **O que faz:** Lê o arquivo `sdd/memory/progress.md` e resume o status das features ativas, próximos passos, bloqueios e percentual de conclusão do projeto. Sempre encerra com a linha `Próximo comando sugerido: <comando>`, calculada a partir do estado real do projeto — elimina a necessidade de memorizar a ordem dos comandos.
- **Exemplo de uso:** 
  > `/status`

### `/tutorial` ou `"me ensine o ciclo SDD"`
- **Quando usar:** Na primeira vez que for usar a metodologia, antes de rodar `/discovery` com uma demanda real.
- **O que faz:** Guia um ciclo SDD completo e fictício (discovery → features → PLAN/ACT/CLOSE), isolado em `sdd/discovery/_tutorial/` e `sdd/features/_tutorial/`, sem tocar `sdd/features/index.md`, `sdd/memory/progress.md` ou telemetria.
- **Exemplo de uso:** 
  > `/tutorial`

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
- **O que faz:** Faz uma varredura completa do repositório para mapear as dependências e tecnologias usadas, alinhando a constituição (`sdd/memory/constitution.md`) do projeto. Também pergunta (opcionalmente) o idioma de interação e o nível de linguagem desejado — `padrão` (jargão técnico) ou `iniciante` (linguagem simplificada, com exemplos concretos) — persistindo a escolha para todos os comandos seguintes.
- **Exemplo de uso:** 
  > `/constitution`

### `/c4-architecture` ou `"desenhar arquitetura"`
- **Quando usar:** Para documentar visualmente a topologia e fluxos do projeto.
- **O que faz:** Desenha diagramas de sistema de níveis 1 e 2 (Contexto e Container) nos padrões C4 Model utilizando Mermaid.
- **Exemplo de uso:** 
  > `/c4-architecture`

### `/doctor` ou `"check-up do projeto"`
- **Quando usar:** Se notar falhas de carregamento, conflito de MCP ou arquivos corrompidos.
- **O que faz:** Verifica a integridade estrutural e a compatibilidade dos arquivos e pastas do Forge-SDD no repositório. Equivalente em chat do `forge-sdd doctor` do CLI, incluindo a detecção de deriva de convenção de nomenclatura.
- **Exemplo de uso:** 
  > `/doctor`

### `/archive` ou `"limpar progresso"`
- **Quando usar:** Quando o arquivo `sdd/memory/progress.md` exceder o limite recomendado de 1 KB.
- **O que faz:** Limpa e move as features finalizadas de `progress.md` para o histórico `progress-log.md`, mantendo o progresso leve.
- **Exemplo de uso:** 
  > `/archive`


### `/upgrade-sdd <versao>` ou `"atualizar sdd para vX"`
- **Quando usar:** Quando uma nova versão do framework forge-sdd for lançada.
- **O que faz:** Inicia a rotina do **Migrator** para atualizar a metodologia no projeto para a versão indicada.
- **Exemplo de uso:** 
  > `/upgrade-sdd 1.6.0`
