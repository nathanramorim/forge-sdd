# ADR-0007: Contexto do Projeto na Discovery, Briefing e Contingência de Terminal

- **Data:** 2026-11-14
- **Status:** `aprovado`
- **Regras afetadas (IDs da constitution):** `DUP-01`, `DUP-02`, `DUP-05`, `FLX-01`, `FLX-05`, `GOV-01`, `ARQ-06`
- **Autor / Responsável:** Líder (decisão do dono do produto no discovery 75)

## Contexto

A experiência inicial com o pipeline multi-sessão e o scaffold de novos projetos revelou duas lacunas críticas:
1. **Falta de contexto na Discovery:** A sessão de concepção iniciava sem ler missão, módulos, fluxos e decisões tomadas, gerando propostas conflitantes com decisões arquiteturais prévias ou reinventando componentes existentes. Além disso, em projetos novos recém-scaffoldados, a constituição nascia como rascunho sem que o usuário fosse orientado a preenchê-la antes de abrir novas features.
2. **Fragilidade na abertura de terminais:** Usuários sem `tmux` instalado encontravam falha ou eram jogados em janelas soltas sem qualquer orientação de como obter a visualização em painéis. No Windows, apesar da disponibilidade do Windows Terminal (`wt.exe`) com suporte a split panes nativos, o pipeline abria janelas separadas.

Referência: `sdd/discovery/75-contexto-do-projeto-e-contingencia-tmux/` e `feat-75-01` a `feat-75-06`.

## Decisão

- **D-1 `forge-sdd brief` e `--check` (`internal/gate/brief.go`).** Criamos um comando determinístico sem efeitos colaterais no disco (`brief`) que lê `sdd/memory/constitution.md`, `sdd/spec/modules.md`, `sdd/spec/flows.md` e `sdd/spec/decisions.md`. Ele compõe um resumo estruturado com orçamento de bytes configurável e avalia a prontidão com `--check` (código de saída 0 para `pronta`, 3 para `rascunho`). Toda sessão de Discovery agora inicia executando `forge-sdd brief` e precisa citar explicitamente as decisões resolvidas na especificação sob `## Decisões do projeto aplicadas`.
- **D-2 Constituição como primeiro passo (`/constitution`).** Em projetos novos ou com constituição em rascunho, o menu interativo do `forge` e o `forge-sdd doctor` destacam `/constitution` como o passo número 1 para alinhar missão, stack, módulos, fluxos e decisões antes da primeira feature, emitindo avisos sem bloquear o desenvolvedor.
- **D-3 Instalação assistida do tmux e queda suave (`internal/pipeline/tmuxinstall.go`).** Em modo `auto`, se o `tmux` não estiver no PATH, o Forge identifica o gerenciador de pacotes do sistema (Homebrew no macOS; apt-get, dnf, pacman, apk ou zypper no Linux) e propõe a instalação interativa com confirmação e suporte a `sudo`. Se o usuário recusar ou a instalação falhar, o pipeline executa contingência suave para janelas nativas do sistema e oferece salvar `"terminal": "native"` em `forge.json` para não incomodar novamente. Em caso de falha no attach após a criação da sessão, `TmuxAttachError` evita duplicação de processos dos workers.
- **D-4 Painéis no Windows Terminal (`internal/pipeline/launcher.go`).** No Windows (`GOOS=windows`) com `wt.exe` disponível, `LaunchPlanFor` constrói um comando único encadeando `new-tab` para a Discovery e dois `split-pane` para Builder e Reviewer separados por `;`, oferecendo a experiência de painéis unificados sem necessidade de tmux. No WSL, o tmux opera normalmente dentro do Linux.

## Consequências

- **Positivas:** Todas as sessões de Discovery operam sob as mesmas decisões e premissas arquiteturais; a experiência do usuário com novos projetos direciona para a definição da Constituição antes do código; a abertura das sessões não quebra quando o tmux não está instalado, oferecendo caminho de instalação ou contingência elegante.
- **Negativas / Trade-offs:** A instalação do tmux depende de gerenciadores de pacotes suportados e permissões adequadas de sudo; em ambientes Windows sem Windows Terminal moderno (`wt.exe`), a contingência recorre a janelas separadas de `cmd.exe`.
