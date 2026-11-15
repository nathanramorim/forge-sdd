# Forge-SDD 3.0 — Brag Video Plan (`/brag`)

## 1. Resumo do Produto e Objetivo Educativo
**Forge-SDD 3.0** (`https://forge-sdd.org` · `@nathanramorim/forge-sdd`) é um framework open source (licença MIT) de desenvolvimento guiado por especificações (SDD) que substitui o modelo caótico de "um único agente tentando fazer tudo na mesma conversa" por uma **esteira autônoma de 5 estações especializadas** com **Governança em 3 Camadas** e **motor CLI determinístico em Go**.

O vídeo tem um objetivo duplo:
1. **Explicar o novo fluxo de 5 estações de forma educativa e visual** — mostrando exatamente o que cada especialista recebe (Entrada), faz e entrega (Saída), como o bastão passa de `01 → 02 → 03 → 04 → 05 ↺` e por que a Governança Somente-Leitura impede que o agente aprove o próprio erro.
2. **Ensinar como instalar e utilizar na prática** — mostrando a instalação via `npm` / `npx` e a sequência real dos 4 comandos do dia a dia (`/discovery`, `/split-features`, `/proxima-feature`, `/revisar`) junto ao motor CLI (`forge-sdd queue`, `forge-sdd gate`, `forge-sdd skills`).

---

## 2. The 9-Question Planning Rubric (`/brag`)

1. **Qual é a história de uma frase?**
   > Em vez de um único agente perdendo contexto e aprovando o próprio código, o Forge-SDD 3.0 coloca 5 especialistas em esteira — do Discovery com aprovação humana até Builders em paralelo e Governança somente-leitura.

2. **Quem está assistindo e onde?**
   > Desenvolvedores, Tech Leads e Engenheiros de Software assistindo na Landing Page (`forge-sdd.org`), LinkedIn, X/Twitter, GitHub README ou YouTube, muitas vezes no mudo no primeiro segundo e buscando entender rapidamente "como funciona" e "como eu uso no meu repositório".

3. **Qual é o momento herói (Hero Visual)?**
   > A **Visão Processual da Esteira (`01 → 02 → 03 → 04 → 05 ↺`)** combinada com o **Emulador Passo a Passo** e o **Simulador de Governança Read-Only**: ver os 5 mascotes especialistas passando o bastão, você aprovando a Spec no Portão Humano (`FLX-01`), os Builders consumindo `sdd/.queue/queue.json` em paralelo e o Auditor (`Estação 05`) barrando uma chave exposta (`SEG-01 · P-13`), devolvendo `05 ↺ 04` para correção e aprovando no `forge-sdd gate` com geração de **Auto-Skill (`DUP-05`)**.

4. **O que fica de fora?**
   > Nada de blocos gigantes de markdown ilegíveis, nada de jargão burocrático solto sem explicação visual e zero referências externas fora do ecossistema Forge-SDD. Cada tela ensina um conceito claro por vez.

5. **Qual é o ritmo e a duração?**
   > **24.0 segundos** exatos (**720 frames a 30 fps**, `1920x1080` landscape 16:9). Dividido em **5 cenas educativas** (~4.0s a 5.5s cada) com transições *staggered dip-through-background* (sem crossfades borrados), permitindo leitura confortável em português.

6. **Qual é a identidade visual?**
   > Extraída diretamente do Design System oficial do Forge-SDD 3.0 (`docs/showcase-v3-lp.html`):
   - **Superfícies:** `#0d0f14` (bg principal com grid 44x44px em `rgba(79,142,247,0.05)`), `#12151d` (cards), `#181c26` (elevated), `#1e2432` (hover/pill).
   - **Tipografia:** Geométrica estilo **Montserrat** (`Avenir Next Bold/DemiBold/Medium` no CoreText) + **JetBrains Mono / Menlo** para comandos de terminal, arquivos e badges técnicos.
   - **Paleta Semântica e dos 5 Especialistas:**
     - `01. O Líder` (Orquestrador): Laranja `#f2a65a` · Chest Mark `MARK_RING`
     - `02. PM & Arquiteto` (Concepção): Índigo `#8aa4ff` · Chest Mark `MARK_DIAMOND`
     - `03. O Planejador` (Fatiador): Violeta `#c9a7ff` · Chest Mark `MARK_BAR`
     - `04. Builders` (Execução): Turquesa `#4fd1b5` · Chest Mark `MARK_WAVE`
     - `05. Governança` (Auditor Read-Only): Coral `#ff8a5c` · Chest Mark `MARK_TRI`
   - **Estados Semânticos:** `#4f8ef7` (Working), `#9b7bff` (Thinking), `#38c6e8` (Searching), `#f5b04a` (Approval / Portão Humano), `#3ecf8e` (Finished / Gate PASS), `#f2677a` (Error / Violação bloqueada).

7. **Qual é a direção de áudio?**
   > Trilha sonora original sintetizada em PCM estéreo 48 kHz / 16-bit (`work/generate-audio.mjs`):
   - Progressão harmônica quente em Fá Lídio / Dó Maior com sintetizador estilo Rhodes/Marimba e pulso rítmico suave a 120 BPM.
   - **Efeitos sonoros sincronizados ao frame:**
     - `t = 0.25s`: alerta sutil do "agente solitário sobrecarregado" seguido do acorde de abertura (`t = 1.15s`) quando as 5 estações entram em cena.
     - `t = 4.00s`: transição *whoosh* para a Cena 2 (Visão Processual).
     - `t = 6.20s`: clique tátil + chime duplo de aprovação humana no Portão `FLX-01` (`Aprovar Spec`) + passagem de bastão `02 → 03 → 04`.
     - `t = 9.00s`: transição *whoosh* para a Cena 3 (Como Instalar e Usar na Prática).
     - `t = 9.60s, 10.80s, 12.00s, 13.20s`: ticks de digitação e confirmação a cada passo do fluxo (`npx @nathanramorim/forge-sdd@latest init`, `/discovery`, `/split-features`, `/proxima-feature`).
     - `t = 14.50s`: transição *whoosh* para a Cena 4 (Governança em 3 Camadas).
     - `t = 15.60s`: tom grave de bloqueio quando a Estação 05 intercepta a violação `SEG-01 · P-13` e aciona o loop `05 ↺ 04`.
     - `t = 17.50s`: arpejo cristalino quando o Builder corrige, o `forge-sdd gate` passa (`PASS`) e a **Auto-Skill (`DUP-05`)** é gerada.
     - `t = 20.00s`: transição *whoosh* + acorde final luminoso na Cena 5 (Outro / Call to Action).

8. **Qual é o texto de compartilhamento (Share Copy)?**
   > Curto (2–3 frases), direto e pronto para postar no LinkedIn, X/Twitter ou GitHub Release (salvo em `brag-output/share-copy.txt`).

9. **Qual frame é o Poster (`brag.jpg` e Frame 0 de `brag.mp4`)?**
   > O frame consolidado da **Visão Processual + Esteira de 5 Especialistas (`t = 7.40s`)**, onde todas as 5 estações, o fluxo `01 → 02 → 03 → 04 → 05 ↺`, o contrato de Entrada/Saída e o selo do Portão Humano `FLX-01` estão 100% nítidos e legíveis. Esse mesmo frame é gravado no Frame 0 do `brag.mp4` para garantir miniatura perfeita em qualquer plataforma.

---

## 3. Storyboard Detalhado Cena a Cena (24.0s / 720 frames @ 30fps)

### Cena 1 (0.0s – 4.0s | Frames 0–120): O Problema vs. A Nova Esteira de 5 Especialistas
- **Objetivo Educativo:** Mostrar por que trabalhar com um único agente falha quando o projeto cresce e apresentar os 5 especialistas do Forge-SDD 3.0.
- **Visual:**
  - Topo: Tag `FORGE-SDD 3.0 · ESTEIRA AUTÔNOMA COM GOVERNANÇA` + Título forte: `"De 1 agente sobrecarregado para 5 especialistas em esteira."`
  - Metade superior/esquerda (comparativo rápido): Card "Antes: 1 Agente Fazendo Tudo" com mascote em `error` (`Mistura spec e código · Esquece regras · Aprova o próprio erro`).
  - Centro/Base (Hero Grid): Os **5 cards das Estações** se acendem em sequência com seus mascotes cápsula, marcas peitorais e papéis claros:
    1. `01. O Líder` (`#f2a65a`) — Orquestra o time e fala com você
    2. `02. PM & Arquiteto` (`#8aa4ff`) — Faz Reuse Check e escreve a Spec
    3. `03. O Planejador` (`#c9a7ff`) — Fatia em subtarefas P/M/G na fila
    4. `04. Builders` (`#4fd1b5`) — Codificam em paralelo sem conflito
    5. `05. Governança` (`#ff8a5c`) — Audita em modo Somente-Leitura

### Cena 2 (4.0s – 9.0s | Frames 120–270): Visão Processual — Como o Bastão Passa (`01 → 02 → 03 → 04 → 05 ↺`)
- **Objetivo Educativo:** Explicar visualmente o fluxo processual entre os agentes (Entrada → Ação → Saída) e o **Portão Humano Obrigatório (`FLX-01`)**.
- **Visual:**
  - Trilho processual conectado no topo: nós `01 O Líder` ──► `02 PM & Arquiteto` ──► `03 O Planejador` ──► `04 Builders` ──► `05 Governança` com pacotes de dados animados viajando pelos conectores e seta tracejada de retorno `↺ Loop de Autocorreção (05 ↺ 04)`.
  - Painel de Contrato da Estação em Foco:
    - Entre `4.0s` e `6.2s`: Foco na **Estação 02 (PM & Arquiteto)** em estado `approval` (`Aguarda você`). Mostra os 3 blocos didáticos:
      - **1. RECEBE (ENTRADA):** `Intenção repassada pelo Líder + Código existente`
      - **2. PORTÃO HUMANO (FLX-01):** Cursor se move e clica em `[ Aprovar Spec · De Acordo ]` em `t = 6.20s`.
      - **3. ENTREGA (SAÍDA):** `Spec aprovada (discovery/criteria/plan) → Passa bastão para Estação 03`
    - Após o clique (`6.2s` a `9.0s`): O Portão `FLX-01` fica verde (`APROVADO PELO USUÁRIO`), o pacote desliza para `03 O Planejador` (`Fatia em subtarefas P/M/G e alimenta sdd/.queue/queue.json`) e aciona `04 Builders`.

### Cena 3 (9.0s – 14.5s | Frames 270–435): Como Instalar e Usar na Prática (Passo a Passo Real)
- **Objetivo Educativo:** Ensinar exatamente como instalar o Forge-SDD 3.0 e quais comandos rodar no dia a dia.
- **Visual (Split Screen Educativo):**
  - **Coluna Esquerda (Instalação + 4 Passos no Chat/Terminal):**
    - Bloco de instalação rápida no topo:
      `$ npm install -g @nathanramorim/forge-sdd   # ou: npx @nathanramorim/forge-sdd@latest init`
    - Linha do tempo interativa dos **4 Comandos do Dia a Dia** (acendendo passo a passo):
      - **Passo 1 (`9.4s`):** `/discovery [ideia]` → *Estação 02 faz o Reuse Check, cria a Spec e pede seu OK.*
      - **Passo 2 (`10.7s`):** `/split-features` → *Estação 03 fatia o plano em subtarefas P/M/G na fila.*
      - **Passo 3 (`12.0s`):** `/proxima-feature` → *Estação 04 (Builders) implementa em paralelo com testes.*
      - **Passo 4 (`13.2s`):** `/revisar` → *Estação 05 roda o `forge-sdd gate` + Auditoria Read-Only.*
  - **Coluna Direita (O que acontece nos bastidores em tempo real):**
    - Mostra o mascote da estação ativa trabalhando + visualização ao vivo da fila determinística `sdd/.queue/queue.json`:
      - `feat-58-01 · Constitution & Rules` (`COMPLETED` · Builder #1)
      - `feat-58-02 · Canonical Roles .agents/` (`CLAIMED` → `COMPLETED` · Builder #2 em paralelo, `0 arquivos sobrepostos`)
      - `feat-58-04 · CLI Go queue & gate` (`CLAIMED` → `COMPLETED` · Builder #3)

### Cena 4 (14.5s – 20.0s | Frames 435–600): Governança em 3 Camadas & Auto-Skills (Por que é Confiável)
- **Objetivo Educativo:** Mostrar por que a IA não aprova o próprio erro no Forge-SDD 3.0 — explicando as 3 camadas (`Camada A: forge-sdd gate`, `Camada B: Analista Read-Only`, `Camada C: Painel do Líder`) e o ciclo de autocorreção `05 ↺ 04`.
- **Visual:**
  - **Fase 1 (`14.5s – 17.2s`) — Violação Interceptada:**
    - A Estação 05 (`Governança · 0 arquivos com permissão de escrita`) examina o PR e bloqueia uma violação real:
      - `CAMADA A (forge-sdd gate): FAIL [SEG-01 / P-13] Segredo detectado em config.go:42`
      - `CAMADA B (Auditor Read-Only): REPROVADO — Devolvendo item para a fila do Builder (05 ↺ 04)`
    - Conector visual mostra o pacote voltando da **Estação 05 (`error`)** para a **Estação 04 (`working`)** com o plano de correção exato.
  - **Fase 2 (`17.2s – 20.0s`) — Correção, Gate Verde e Auto-Skill (`DUP-05`):**
    - O Builder remove a chave hardcoded (`os.Getenv("API_KEY")`) e reenvia.
    - Os 3 portões ficam verdes (`PASS`):
      - `Camada A · forge-sdd gate`: `PASS (0 segredos · testes verdes · spec vinculada)`
      - `Camada B · Auditoria Read-Only`: `APROVADO (Constitution & Rules 100% íntegras)`
      - `Camada C · Auto-Skill (DUP-05)`: `Nova skill criada e sincronizada em .agents/skills/`

### Cena 5 (20.0s – 24.0s | Frames 600–720): Outro — Comece Agora em `forge-sdd.org`
- **Objetivo Educativo:** Consolidar os 5 especialistas e entregar os comandos e links oficiais de forma memorável.
- **Visual:**
  - Os 5 mascotes especialistas alinhados comemorando em estado `finished` com estrelas cintilantes (`01. O Líder`, `02. PM & Arquiteto`, `03. O Planejador`, `04. Builders`, `05. Governança`).
  - Headline: `"Especificação clara. Execução em fila. Governança que não mente."`
  - Caixa de comando pronta para copiar:
    `$ npx @nathanramorim/forge-sdd@latest init`
  - Rodapé de ecossistema:
    `forge-sdd.org` · `github.com/nathanramorim/forge-sdd` · `Claude Code · Gemini CLI · GitHub Copilot` · `Open Source · MIT License`
