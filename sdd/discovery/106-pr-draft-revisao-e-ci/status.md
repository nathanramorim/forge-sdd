# Status — 106-pr-draft-revisao-e-ci

**Status:** Gate 1 aprovado em 2026-11-20 (completo). Spec realinhada à 109 (módulos externos) em 2026-11-20; aguarda novo Gate 1 desta revisão. Nada implementado.
**Modo:** com-humano (padrão). **Atestado:** local (padrão); `vcs` adiado.
**Decisões mantidas:** (1) PR único da Constituição na task 0.8 (ADR-0022), campo `Módulos permitidos` no lugar de `Quadro de acompanhamento`, `modulo` no `ARQ-06`, 9.173 de 9.216 bytes; (2) fallback só `github` e `nenhum`; (3) `vcs` e Azure adiados.
**Próximo passo:** `/split-features` (`## Preparo para o /split-features` do `plan.md`). Podem começar já: 1.1, 0.2, 0.7, depois 1.3. Branch por feature exige `constitution-exception` (`GOV-02`) até a 0.8.

## A confirmar
`## Ordem das ondas do core` na 109 ainda não existe; a ordem entre ondas fica "definida na 109, a conferir".

## Colisões
107-12 depende da 0.5 e da 0.8. 105: templates, goldens, `hooks.go`, `doctor`, menus.
