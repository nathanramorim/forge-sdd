---
name: pr-novo-quando-force-push-e-negado
description: Rebase de branch com PR aberto exige push forçado, que o classificador nega: abrir PR novo em vez de contornar
status: proposed
origin: auto-governance
---

# Skill: pr-novo-quando-force-push-e-negado

1. Antes de criar a branch: git fetch e partir de origin/main, para não precisar de rebase.
2. Se um rebase for inevitável e o push forçado for negado, não tente outra forma de forçar.
3. Crie uma branch nova com o mesmo conteúdo, empurre sem force, abra PR novo em draft apontando o antigo, feche o antigo com comentário e apague a branch remota antiga.
4. Se o conteúdo mudou depois do parecer, peça nova revisão do Reviewer (o guard exige parecer do commit atual).
