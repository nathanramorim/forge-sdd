# feat/56-lp-uso-forge-sdd

**Branch:** `feat/56-lp-uso-forge-sdd`
**Fase:** 56
**Depende de:** repositório externo `lp-forge-sdd` (fora deste monorepo — implementação real acontece lá, via sessão `lp-forge`)
**Status:** `ready for handoff`

## Objetivo

Ilustrar de forma prática, didática e de fácil entendimento — numa nova
seção da landing page (lp-forge-sdd) — como é usar o forge-sdd no dia a
dia, aproveitando as 5 imagens já geradas nesta sessão.

## Contexto

As imagens mostram uso real do CLI/comandos forge-sdd (não o histórico
de PRs do próprio forge-sdd, que foi descartado por não ilustrar uso).
Foram capturadas num projeto de demonstração isolado
(`notifica-app`, fora deste repo), com um fluxo real: `forge-sdd init`
→ `/discovery` → `/nova-feature` → `/status`, todos com output genuíno.

## Assets (entregues via SendUserFile nesta sessão)

| # | Arquivo | Conteúdo |
|---|---------|----------|
| 1 | `u0-init.png` | `forge-sdd init` criando a estrutura `sdd/` + `.claude/commands/` |
| 2 | `u1-discovery.png` | `/discovery` gerando `discovery-01` + `criteria-01` a partir de uma demanda em linguagem natural |
| 3 | `u2-criteria-file.png` | Conteúdo real do `criteria-01` gerado (inclui diagrama C4/Mermaid) |
| 4 | `u3-nova-feature.png` | `/nova-feature` quebrando o discovery em `feat-01` acionável |
| 5 | `u4-status.png` | `/status` mostrando o `progress.md` atualizado com a feature pronta pro Builder |

## Escopo (para a sessão lp-forge implementar)

- [ ] Nova seção na LP com narrativa em 4-5 passos: `init → discovery → nova-feature → status`, uma imagem por passo, em ordem.
- [ ] Texto curto por passo (1-2 linhas) explicando o que o comando faz, em linguagem acessível a quem nunca usou o forge-sdd.
- [ ] Layout responsivo — mobile precisa ler a sequência sem cortar as imagens.
- [ ] Alt-text descritivo em cada imagem (acessibilidade).

## Fora de escopo

- Novas capturas de tela (as 5 já entregues são suficientes para a narrativa).
- Qualquer alteração de código no repositório forge-sdd (este repo não contém a LP).

## Critérios de Aceitação

- Seção publicada exibe as 5 imagens na ordem correta com legendas.
- Fluxo é compreensível para alguém que nunca usou forge-sdd (sem jargão não explicado).
- Nenhum código deste repositório (forge-sdd) foi alterado para viabilizar a seção — o trabalho é 100% no repositório `lp-forge-sdd`.

## Handoff

Repassar este spec + os caminhos/uuids das 5 imagens para a sessão `lp-forge`
implementar diretamente no repositório `lp-forge-sdd`, em
`assets/images/forge/` (conforme já indicado por ela).
