# Status — Iniciativa 92

**Branch:** `feat/92-paralelismo-automatico-builder-reviewer` (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** discovery, critérios e plano prontos; aguardando o Gate 1.
**Próximo passo:** aprovar e rodar `/split-features` (9 tasks em 3 ondas). Depende da 91-07.

## A confirmar (premissas assumidas sem perguntar)
- Quem paraleliza é o Go (um worker com vagas), não subagentes do modelo.
- Dependência conta como pronta quando integrada, não quando aprovada.
- "Assim como o Reviewer": uma revisão por task, várias ao mesmo tempo.
- A integração acontece no seu checkout, que precisa estar na branch da iniciativa.
- Nasce desligado; liga com "um job por task" e vagas maiores que 1.
- Numeração: ADR-0017 supõe a 0015 (alinhamento) e a 0016 (91) antes.
