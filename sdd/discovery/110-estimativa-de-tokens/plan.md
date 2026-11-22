# Plano — 110 Régua de estimativa de tokens

Tasks do **módulo** têm prefixo `E-`; mudanças no **core** têm prefixo `K-` e só entram se o dono aprovar (decisões 2 e 3). Até 7 tasks por onda.

## Onda 1 — primeira entrega pequena e útil: `forge estimar` só leitura (não depende da 109)
| Task | O quê | Depende | Rules |
|---|---|---|---|
| E-00 | Bootstrap do repositório (abaixo; só após decisão do dono) | decisão 5 | FLX-01 |
| E-01 | Leitor da amostra: `sdd/.metrics`, logs com `result`, tamanho das specs; classifica real, estimado e zero | E-00 | SEG-01, SEG-05, DUP-02 |
| E-02 | Estatística: mediana, p25–p75, n, rótulos "sem estimativa" e "amostra pequena" | E-01 | TST-01 |
| E-03 | `forge-estimar relatorio`: tabela por papel com a definição de token e o aviso | E-02 | COD-05 |
| E-04 | `precos.json` opcional com `atualizado_em`; sem arquivo, sem dinheiro | E-03 | SEG-02 |
| E-05 | Fixtures anonimizadas e testes sem rede | E-01 | TST-03, TST-04 |
| E-06 | README (readme-studio) e diagramas | E-03 | — |

## Onda 2 — estimativa por feature
| Task | O quê | Depende |
|---|---|---|
| E-07 | Ler `features/*.md` e `plan.md` (bytes, critérios, tipo) | onda 1 |
| E-08 | `estimativa.md`: linha por feature e total da iniciativa | E-07 |
| E-09 | Recomendação descritiva de agente, modelo, ferramenta (só com n suficiente) | E-08 |
| E-10 | Adaptador da transcrição do Claude (`usage` por mensagem), isolado | onda 1 |

## Onda 3 — contrato e calibração (depende da 109 e do dono)
K-01 e K-02 são as donas, no core, do evento `sessao.registrada` e do campo `tool`; a 111 (task 111-C1) apenas consome (decisão do dono em 2026-11-21: ambas na onda 5 da 109, depois de 109-06 e 109-07).
| Task | O quê | Depende |
|---|---|---|
| K-01 | Evento aditivo `spec.criada` e `sessao.registrada` no contrato v1 (onda 5 da 109) | 109-06 |
| K-02 | Campo opcional `tool` no JSON da sessão gerado pelo pipeline (onda 5 da 109) | decisão 3, 107; 109-07 |
| E-11 | Reagir aos eventos; histórico estimado versus real | K-01 |
| E-12 | `forge estimar calibrar` e meta de erro | E-11 |
| E-13 | Skill fina (opção C) | E-08 |

## P-00 / E-00 Bootstrap do repositório (descrito, não executado)
1. **forge-sdd estável:** conferir `npm view @nathanramorim/forge-sdd dist-tags` (em 20/11/2026 `latest` = 5.0.0, tag `v5.0.0`), instalar a versão **exata**, conferir `forge-sdd version`; divergência com a release `Latest` do GitHub: parar e perguntar.
2. **Configurar:** `forge-sdd init --yes --agent claude --name forge-sdd-estimar --stack go --lang pt-BR`, Constituição com Go e stdlib e `contrato.min = 1`.
3. **Diagramas** com o diagram-studio (assinatura Forge SDD, sem nome pessoal nem LinkedIn), cada PNG aberto e conferido.
4. **README** com o readme-studio (skill `readme-comercial`); o plano é mostrado ao dono antes de gravar.
5. Licença e visibilidade conforme o dono; o repositório só é criado depois.

## Ordem
Onda 1 inteira roda sem 109, 106 ou 107. A onda 3 espera o contrato da 109 e, para o estado por feature, a 106/107.
