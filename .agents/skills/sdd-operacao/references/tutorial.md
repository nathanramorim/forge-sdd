# Comando: tutorial

**Uso:** `/tutorial` guia um ciclo fictício do método SDD sem tocar nos dados reais.

Guie um ciclo fictício do método SDD sem tocar nos dados reais:
1. Crie `sdd/discovery/_tutorial/` com `discovery.md`, `criteria.md` e `plan.md` para uma demanda simples (por exemplo, um endpoint de health-check).
2. Quebre em 1 ou 2 features em `sdd/discovery/_tutorial/features/`, explicando cada seção.
3. Simule `PLAN`, `ACT` e `CLOSE` de uma feature, narrando Líder, Builder e Governança.
4. Explique a analogia com o ciclo real e que tudo fica isolado em `_tutorial/`.

**Guardrail:** nunca escreva em `sdd/.metrics/`, na fila real nem no `status.md` de iniciativas reais.

Ao fim, sugira `/discovery` com uma demanda real.
