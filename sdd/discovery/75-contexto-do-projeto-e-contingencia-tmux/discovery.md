# Discovery 75 — Contexto do Projeto na Discovery e Contingência do tmux

## Problema
1. **A Discovery esquece o projeto.** Em cada sessão a pessoa precisa lembrar à Discovery das decisões já tomadas: Constitution, arquitetura, módulos, fluxos, segurança, ADRs e rules. Causa: a leitura mínima carrega só os IDs que a tarefa cita; o prompt da sessão pede ADRs e rules só "antes de quebrar tarefas"; `/constitution` é opcional e a Constitution recém-gerada é placeholder; `modules.md`, `flows.md` e `decisions.md` existem mas nada manda lê-los.
2. **Sem tmux, a abertura falha ou degrada sem aviso.** Em `auto`, sem tmux o Forge abre janelas soltas sem dizer por quê; com `--terminal tmux` dá erro. No Windows nativo não há tmux e o Forge abre 3 janelas separadas, sem painéis lado a lado.

## Para quem
Quem usa `forge` num projeto novo ou existente e conversa com a Discovery: precisa que ela já nasça sabendo o que foi decidido, e que a tela de 3 painéis funcione em qualquer sistema.

## Jornada alvo
1. `forge-sdd init` (ou o menu `forge`) mostra como **primeiro passo** `/constitution`.
2. `/constitution` entrevista, varre o repositório e preenche Constitution, módulos, fluxos, decisões e segurança. Ao terminar, `forge-sdd brief --check` diz `pronta`.
3. Toda sessão Discovery abre rodando `forge-sdd brief` e já responde com base nas decisões, sem a pessoa repetir nada. Cada spec cita as decisões do projeto que aplica; conflito com uma decisão exige ADR.
4. `forge`/`up` sem tmux: o Forge mostra o comando de instalação do sistema, pergunta e instala ao confirmar. Se a pessoa recusar ou o sistema não suportar, a abertura continua por uma contingência (Windows Terminal com painéis; janelas separadas).

## Regras de negócio
- O `brief` é **somente leitura**, determinístico e curto (orçamento de bytes); não substitui a leitura do arquivo-fonte, aponta para ele.
- Constitution em rascunho **avisa, não bloqueia** (a pessoa pode estar começando).
- Instalação do tmux só com confirmação explícita; o Forge nunca guarda nem pede senha (o `sudo` pede no terminal da pessoa). Sem terminal interativo, só imprime o comando.
- Recusar a instalação pode ser lembrado (`terminal: native` em `forge.json`), para não perguntar toda vez.
- Em `auto`, falha do tmux cai para a contingência com aviso; `--terminal tmux` explícito continua dando erro com instrução.

## Escopo
**Dentro:** `forge-sdd brief` (+ `--check`); protocolo de abertura da Discovery (prompt, skills, templates dos agentes); `/constitution` como primeiro passo e povoando módulos, fluxos, decisões, segurança; avisos no `init`, no menu, no `up` e no `doctor`; detecção e instalação do tmux por gerenciador de pacotes; painéis do Windows Terminal; queda do tmux para a contingência; docs, ADR e release notes.
**Fora:** instalar o WSL; instalar gerenciadores de pacotes (brew, winget); bloquear o pipeline por Constitution em rascunho; reescrever Constitutions existentes; mudar o conteúdo das ADRs.

## Riscos
| Risco | Mitigação |
|---|---|
| Brief grande gasta contexto | Orçamento de 6 KB com truncamento explícito e ordem de prioridade (IDs e ADRs vigentes primeiro) |
| Modelo ignora a instrução de ler o brief | Prompt de sessão + skill + `CLAUDE.md`; o `brief --check` e o aviso no `up`/`doctor` tornam a omissão visível |
| Instalar pacote do sistema é invasivo | Sempre perguntar, mostrar o comando exato, argv sem shell, nunca silencioso |
| Windows Terminal com painéis não testado em Windows real | Teste do argv gerado; marcado "a confirmar" até alguém rodar em Windows |
| `winget`/`apt` mudam de comportamento | Comando é só sugestão confirmada; falha do instalador cai na contingência |
