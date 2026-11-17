# Status — Iniciativa 93

**Branch:** `feat/93-coleta-de-contexto` (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** discovery, critérios e plano prontos; aguardando o Gate 1.
**Próximo passo:** aprovar e rodar `/split-features` (7 tasks em 2 ondas). Só a 93-07 depende da 91-02.

## A confirmar (premissas assumidas sem perguntar)
- A hora de coletar é decidida por gancho e Go, não por uma skill.
- A coleta automática compacta (resume); nunca limpa tudo sozinha.
- O Forge não digita `/compact` no seu painel: no teto o Claude Code compacta; na meta, ociosidade ou aviso.
- Valores iniciais: meta de 100 mil e teto de 150 mil tokens.
- Nasce desligada e vale só para o `claude` nesta onda.
- Nome do comando: `forge session gc`.
- No build: quais ajustes de compactação do Claude Code são públicos.
- Numeração: ADR-0018 supõe a 0015, a 0016 (91) e a 0017 (92) antes.
