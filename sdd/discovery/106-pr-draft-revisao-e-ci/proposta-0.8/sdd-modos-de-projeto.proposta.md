---
description: Como ler e respeitar os campos de modo da tabela Ferramentas e Integrações da Constituição (VCS, atestado, revisão humana, WIP humano, módulos permitidos e início automático).
globs: ["sdd/memory/constitution.md", "internal/vcs/**", "internal/config/**", "scripts/atestado.py"]
regras: [ARQ-06, SEG-05, FLX-01, FLX-02, FLX-06, GOV-01, GOV-04]
---

# Guia Prático — Modos de projeto (campos da Constituição)

Os campos ficam na tabela `Ferramentas e Integrações` de `sdd/memory/constitution.md`. Todos são lidos da **branch base**: a branch da feature não os sobrepõe. O valor efetivo da sessão vai ao `status.md` e ao veredito. Mudar um campo ou um valor aceito exige ADR (`GOV-01`).

## 1. VCS / Work Item System (`FLX-02`)
- `github`: `gh pr create --draft`, `gh pr ready`, `gh pr merge` pelo adaptador `internal/vcs`.
- `azure-devops`: **não verificado** nesta entrega. O adaptador recusa a operação com a mensagem "não verificado" e o `forge doctor` informa. Antes, `az repos pr create` substituía `gh pr create`.
- `nenhum`: deixe a branch pronta e informe o usuário, sem comando de VCS; o merge é local.
- O status real dos MCPs fica em `sdd/memory/mcps.md` (`ativo` ou `indisponível`).

## 2. Modo do atestado (`FLX-06`, `SEG-05`)
- Valores: `desligado`, `local` (padrão) e `vcs`. Vale **só para PRs de feature**; o PR da spec é sempre `local`.
- `local`: veredito conforme com `head_sha` igual ao HEAD mais a aprovação do humano no chat. Sem prova externa: quem tem acesso ao repositório pode forjar.
- `desligado`: ninguém confere veredito nem aprovação; o PR pode sair sem conformidade.
- `vcs` (prova do VCS com aprovadores): **adiado**. É aceito como valor, mas recusado com a mensagem "modo adiado". `nenhum` só admite `desligado` ou `local`.
- Cadeia de fallback entregue: `local`, com tempo limite; sem resposta, pergunta ao humano (no modo sem humano, avisa e grava). Nunca aprova em silêncio.

## 3. Modo de revisão humana (`FLX-01`)
- `com-humano` (padrão) ou `sem-humano`. O segundo só vale se escolhido no início da sessão, depois do aviso dos riscos: nenhuma pessoa vê o código antes de o PR sair do draft; erro do revisor passa; o custo fica sem teto humano.

## 4. WIP humano (`FLX-01`)
- Inteiro positivo (padrão 3) ou `desligado`: limite de itens aguardando o humano. No limite o orquestrador para e pergunta em menu. Reduzir ou desligar exige o aviso dos riscos e o registro no `status.md`. Distinto de `builder.max_parallel_tasks`.

## 5. Módulos permitidos e Início automático do Builder (`ARQ-06`, `SEG-03`)
- `Módulos permitidos`: `nenhum` (padrão; campo ausente vale o mesmo) ou lista de nomes. Só roda módulo listado aqui, instalado na pasta do usuário, com hash conferido e autorizado pelo dono para aquela versão. `nenhum` não lê disco do usuário nem chama processo.
- A configuração de cada módulo fica em `sdd/modulos/<nome>.json`, opaca ao core. O quadro do GitHub Project é um módulo externo, opcional, e não um campo da Constituição.
- `Início automático do Builder`: `desligado` (padrão) ou `local`; é do laço da 107.
- Nota copiada do `discovery.md` da 109 (commit `b7c5745`, seção "Constituição").

## 6. Como cumprir
- Leia os campos por um leitor único (`internal/vcs`), com valor inválido recusado na borda (`SEG-01`); não copie os valores para `forge.json`.
- Texto de comando ou template que cite um modo aponta para esta rule, sem repetir o conteúdo (`GOV-04`).
