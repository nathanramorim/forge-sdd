# Discovery 120 — Fatiamento vertical e Strangler Fig nos planos do Forge

**Origem:** Nathan, 2026-11-23: "ver onde encaixa isso para melhoria", sobre o texto "Por que não entregar tudo de uma vez? Teoria de Sistemas aplicada às entregas e ao Strangler Fig".

## Princípio vale para os planos, não para o forge-sdd
O princípio vale para os **planos que o Forge gera para os projetos dos usuários** (`plan.md`, tasks, veredito do Reviewer). Não é refatoração do código do forge-sdd. Só mudam textos de skills, seus templates embutidos e o golden.

## Problema
Hoje o fatiamento manda quebrar "de baixo para cima" (infraestrutura, domínio, aplicação, interface) e pede só "testável sozinha". Isso gera fatias horizontais: nenhuma entrega funciona ou valida até a última camada subir (Big Bang em parcelas). O risco se acumula, a causa de uma falha fica difícil de achar, e migração de legado não tem orientação (risco de "construir o novo inteiro e virar a chave").

## Para quem
Pessoa que roda `/discovery` e `/split-features` num projeto gerado pelo Forge; Builder que recebe as tasks; Reviewer que dá o veredito.

## Fundamento (resumo do texto)
Quanto maior a complexidade, menor deve ser cada parte entregue. Fatiar se apoia em: (a) fatias verticais colocadas no ar separadamente; (b) ciclos curtos de feedback; (c) baixo acoplamento de implantação (se 10 módulos sobem juntos para um funcionar, ainda há acoplamento). Strangler Fig real: fatia pequena e de baixo risco (um fluxo ou endpoint, não o módulo), desvio só daquele tráfego por roteador, gateway ou feature flag, produção e validação, repetir até matar o legado. Construir o novo em paralelo e "virar a chave" é Big Bang disfarçado. Referências: Fowler (Strangler Fig Application); Humble/Kim et al. (Continuous Delivery, Manual de DevOps); Kniberg (Making Sense of MVP); Newman (Monolith to Microservices: Canary Release, Parallel Run, Branch by Abstraction).

## Regras de negócio
- RN-1: cada task é uma fatia vertical: atravessa as camadas necessárias para entregar um comportamento observável, valida sozinha e entra na main sem outra task para funcionar.
- RN-2: o `plan.md` ganha a seção `## Fatias e validação`: fatia piloto de baixo risco primeiro; como cada fatia é validada; como é revertida.
- RN-3: se a iniciativa é migração ou substituição de legado, a seção traz a variante Strangler: desvio por roteador ou flag, fatia pequena, repetir até aposentar o legado.
- RN-4: o Reviewer checa acoplamento de implantação (S-08) no veredito.
- RN-5: a ordem "de baixo para cima" continua válida **dentro** de uma fatia; deixa de ser o critério de quebra entre tasks.

## Escopo
**Dentro (encaixes 1, 2 e 4 decididos pelo Nathan):**
1. `split-features.md` passo 4 (critérios de quebra).
2. `plan.md`: passo 7 do `discovery.md` e `sdd-planejamento/SKILL.md`: seção `Fatias e validação`.
4. Governança: verificação S-08 em `sdd-governanca/SKILL.md` e `pr-review.md`.
Cada um no arquivo do repositório, no template embutido (`internal/scaffold/templates/...tmpl`) e no golden.

**Fora:** encaixe 3 e demais; Constituição; ADR; código Go; o código do forge-sdd em si; `forge dispatch`/fila.

## Decisões em aberto para o Gate 1
- DA-1 (**Constituição/ADR**): S-08 não tem ID da Constituição próprio. Sem criar regra, o achado cita `regra` de um ID existente (proposta: `FLX-03`, aderência à spec) e nasce como `duvida` quando não houver evidência literal. Criar ID novo exigiria Constituição/ADR, fora do escopo; decisão do Nathan.
- DA-2 (**contrato do Parecer**): o contrato aceita `verificacao` como string livre (`internal/harness/parecer.go`, sem validação de valores), então S-08 **não exige mudança de código Go nem ADR**. Campo estruturado novo (ex.: `acoplamento`) exigiria Go e ADR; não proposto.
- DA-3 (**deriva do template**): `sdd-governanca/SKILL.md` do repositório já difere do template (`regra_descricao` e texto de AskUserQuestion só no repositório). Proposta: alinhar o template ao repositório na task 3, ampliando o golden além do S-08. Confirmar.
- DA-4: `plan.md` do projeto não tem template próprio no `forge init`; a seção vive só nas instruções (discovery.md e SKILL). Aceitável? Template de plan seria escopo novo.

## Riscos
- R-1: golden e cópia do repositório divergirem do template (mitigado: `go test ./internal/scaffold/ -update` e depois sem a flag).
- R-2: fatia vertical em projeto de camada única parecer artificial (mitigado: o texto admite exceção justificada no `plan.md`).
- R-3: falso positivo de S-08 (mitigado: `duvida` sem evidência literal, `FLX-08`).
