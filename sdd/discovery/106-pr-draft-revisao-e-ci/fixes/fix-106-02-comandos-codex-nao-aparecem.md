# fix-106-02 — Ações Forge SDD não aparecem no seletor do Codex

**Iniciativa:** 106-pr-draft-revisao-e-ci  
**Branch:** `fix/106-02-codex-slash-commands`  
**Origem:** relato do usuário e auditoria do scaffold Codex (21/11/2026)  
**GitHub:** [issue #200](https://github.com/nathanramorim/forge-sdd/issues/200)  
**Tamanho:** `M`  
**Status:** `queued` (backlog; Gate 1 aprovado pelo pedido explícito do usuário em 21/11/2026)

## Descrição de negócio (release notes)

Quem usa Forge SDD no Codex poderá encontrar e acionar os fluxos de concepção e operação pelo mecanismo que o Codex realmente suporta, sem receber instruções para comandos que não aparecem no seletor.

## Problema

O scaffold do Forge SDD anuncia `/discovery`, `/implementar`, `/revisar`, `/status` e outros comandos no `AGENTS.md`, mas essa lista não registra comandos no Codex. Na instalação auditada, `.openai/prompts/` só contém `constitution.prompt.md` e `.agents/skills/` contém skills por estação, sem entradas específicas para todos esses fluxos. Por isso, digitar `/` no Codex não mostra os comandos do Forge.

O mecanismo antigo de custom prompts do Codex é local ao `CODEX_HOME`, não é compartilhado pelo repositório e está oficialmente obsoleto. A instalação não pode presumir que arquivos em `.openai/prompts/` serão descobertos pelo Codex nem prometer que skills criem aliases slash personalizados.

## Feature Correlacionada

- Origem: `sdd/discovery/106-pr-draft-revisao-e-ci/features/feat-106-26a-comandos-de-concepcao.md` e o scaffold/instruções Codex derivados de `internal/scaffold/`.
- Causa-raiz: a lista de slash commands foi tratada como instrução de prompt, sem artefato registrado na superfície de comandos do agente. Os arquivos `.openai/prompts/` não correspondem ao diretório local `CODEX_HOME/prompts` que o recurso legado lia; as skills existentes têm nomes de estação e não registram `/discovery` como alias.
- Evidência: `AGENTS.md` lista os aliases; `.openai/prompts/` não contém esses comandos; o Codex 0.171.0 instalado não os mostra ao digitar `/`. A documentação atual recomenda Skills e marca custom prompts como obsoletos.
- Sequenciamento: revisar a solução com a `feat-106-26a` antes de implementar; não editar os mesmos templates em paralelo.

## Modelo do módulo

**Enxuto** — correção de integração e artefatos gerados para Codex, sem novo serviço ou estado persistente (`ARQ-11`).

## Decisões do projeto aplicadas

| Decisão | ID | Aplicação |
|---|---|---|
| Adaptadores e contratos por agente | ADR-0004, ADR-0005 | A instrução e a superfície de ativação devem corresponder ao Codex real, preservando os comandos dos demais agentes. |
| Skills como fluxo reutilizável e fonte única | ADR-0003, ADR-0012 | Preferir mecanismo suportado e compartilhável; não duplicar o corpo dos comandos em prompts locais obsoletos. |
| Compatibilidade documentada | ARQ-06, FLX-01 | Não anunciar um comando como disponível sem prova de descoberta e invocação na interface alvo. |

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Comandos de concepção | `feat-106-26a`, `.agents/skills/sdd-concepcao/` | Skill canônica e referências por comando; comandos listados em arquivos de instrução | Estender | Reusar o fluxo canônico sem copiar conteúdo. |
| Integração Codex | `.openai/`, `.agents/skills/`, templates em `internal/scaffold/` | `system_instructions.md`, `constitution.prompt.md` e skills por estação; sem registro dos aliases da lista | Corrigir | A localização atual não registra os comandos no Codex. |
| Descoberta de prompts e skills | Documentação oficial atual do Codex | Skills são recomendadas; custom prompts são obsoletos, locais ao `CODEX_HOME` | Adotar mecanismo suportado | Evitar depender de uma superfície local e depreciada. |

## Dependências novas

Nenhuma dependência de runtime. Consultar documentação oficial/context7 na implementação para confirmar a superfície suportada na versão mínima do Codex (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-06`, `DUP-10`, `FLX-01`, `TST-02`, `TST-03`, `GOV-04`, `SEG-07`.

## Critérios de Aceitação Executáveis

1. O scaffold gerado para Codex não anuncia `/discovery` ou outros aliases como comandos disponíveis sem que a instalação registre esses comandos numa superfície suportada pelo Codex.
2. Os fluxos Forge SDD necessários ficam descobríveis e invocáveis pelo caminho suportado do Codex (por exemplo, skill via seletor de skills ou `$nome-da-skill`), apontando para as referências canônicas sem duplicar o procedimento.
3. O README/instruções gerados mostram o caminho exato para localizar e acionar cada fluxo no Codex; a nomenclatura não promete um alias `/` que o Codex não oferece.
4. A verificação de scaffold cobre o alvo Codex e confirma que os caminhos emitidos correspondem ao mecanismo suportado; os adaptadores Claude Code, Gemini/agy e Copilot mantêm os comandos que já registram.
5. A descoberta e a invocação são confirmadas numa sessão limpa do Codex da versão mínima declarada. Se a UI não oferecer alias customizado, a entrega documenta a alternativa suportada e não marca `/discovery` como disponível.

## Dependências e backlog

- Depende de: `feat-106-26a` (comandos de concepção), para sequenciar mudanças nos mesmos templates.
- Estado: `queued`; não iniciar Builder até a dependência e a ordem de colisão serem resolvidas.
