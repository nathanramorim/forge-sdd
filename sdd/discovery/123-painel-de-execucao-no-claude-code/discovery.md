# Discovery 123 — Painel de execução no Claude Code (mods)

**Origem:** Nathan, 24/11/2026, no chat: "analisar o que pode se tornar um mod do Claude na estrutura do Forge SDD; eu quero acompanhar os gastos de perto e a execução dos meus agentes". Depois: "vou deixar tudo no Forge [...] só constrói a arquitetura para que isso possa acontecer lá na frente", com mods compatíveis com o Forge SDD.

**Mod** é uma extensão do Claude Code que muda a interface ou o comportamento dele: status line, painel lateral, alerta (toast), faixa acima do prompt, slash command e ganchos nas chamadas de ferramenta e de subagente.

## Problema
1. **Consumo invisível durante a execução.** No plano por assinatura, o que limita o trabalho são as janelas de uso (5 horas e 7 dias). Hoje só se descobre que a janela acabou quando o agente para.
2. **Telemetria estimada e manual.** O passo CLOSE pede `forge session record` ao fim de cada sessão; quando o agente esquece ou não sabe os tokens, o CLI estima por bytes. A 110 contou 97 sessões, 88 sem tokens.
3. **Agentes em caixa preta.** Com vários subagentes (e, com a 122, um Builder por task), não há uma visão ao vivo de quem está rodando, quanto já consumiu e quem terminou.
4. **Sem lugar para extensões de interface.** A ADR-0025 define onde moram skills e plugins (`.agents/plugins/<slug>/`), mas não há contrato para um mod do Claude Code nem um mod de referência.

## Para quem
Quem usa o Forge SDD com o Claude Code no terminal ou no app desktop e quer ver, sem perguntar, quanto da janela já foi e o que cada agente está fazendo.

## Jornada alvo
1. A pessoa instala o mod pelo forge: `forge skills instalar plugin forge-painel --de <origem>` (ADR-0025), só no projeto aberto.
2. Ao abrir o Claude Code, a status line mostra iniciativa, estação, janela de 5h, janela de 7d e contexto.
3. Ao passar de 70% e de 90% de uma janela, um alerta avisa e diz quando ela renova.
4. Cada subagente que termina grava sua telemetria em `sdd/.metrics` com `medicao: real`, sem o passo manual.
5. Um painel lateral mostra a esteira inteira aberta pelo `forge up`: jobs de Builder e Reviewer com estado, task, tokens e tempo (lidos do barramento e dos logs), mais os subagentes da própria sessão.
6. Antes de abrir um subagente com a janela acima do limite, o mod avisa (não bloqueia na primeira versão).
7. `/gastos` resume a iniciativa por estação, lendo `sdd/.metrics`.

## Regras de negócio
- **RN-1** Só leitura do que o Claude Code já mede: `session.measure`, `$.session.usage()` simples e `$.agent.list()`. Nenhuma chamada de modelo pelo mod.
- **RN-2** Sem rede. Escrita só em `sdd/.metrics/` e `.forge/runtime/painel/`.
- **RN-3** O registro gravado segue o schema `forge-sdd/metrics/1.0` existente; nenhum campo novo sem ADR (`DUP-04`). O painel de métricas da 111 lê esses registros sem mudança.
- **RN-4** O valor em dólar é sempre rotulado "custo de tabela" (estimativa do Claude Code), nunca fatura.
- **RN-5** Guarda de orçamento começa só avisando; bloquear exige decisão do dono depois de medir.
- **RN-6** O mod é opcional e exclusivo do Claude Code. Os outros agentes seguem com `forge session record`.
- **RN-7** Todo mod do Forge SDD segue o mesmo contrato: manifesto, permissões declaradas (rede, escrita, ganchos), versão mínima do Claude Code e versão mínima do forge. O `forge-painel` é o mod de referência desse contrato.
- **RN-9** Convivência com `forge up`: o mod roda na sessão interativa (forge-discovery), que vira o painel da esteira inteira. Os workers (forge-builder e forge-reviewer, com `claude -p`) continuam como estão, em janelas do sistema ou tmux (ADR-0026).
- **RN-10** O painel enxerga os workers lendo o que o Go já grava: estados do barramento em `.forge/runtime/`, logs dos jobs e `sdd/.metrics`. O mod nunca escreve no barramento nem em arquivo de outro processo; o Go segue como escritor único.
- **RN-11** Um dono por registro de telemetria. Job de worker é gravado pelo worker (`recordJobTelemetry`); o mod grava só a sessão interativa e os subagentes dela. Se o mod carregar dentro de um worker (variável `FORGE_WORKER` presente), ele não grava nada e não abre painel.
- **RN-12** O gancho `UserPromptSubmit` do inbox em `.claude/settings.json` continua; o mod não repete o aviso enquanto esse gancho existir.
- **RN-8** Dado sensível nunca entra na telemetria nem na tela (`SEG-05`): só ids, contagens e nomes de papel.

## Escopo
**Dentro:** contrato de mod (manifesto e permissões) com ADR; mod de referência `forge-painel` com status line, alertas, telemetria automática, painel de agentes, aviso de orçamento e `/gastos`; instalação pelo `forge skills instalar plugin`; testes do mod com `claude plugin test`.

**Fora:** painel HTML de métricas (é da 111); estimativa de tokens antes da execução (110); distribuição fora do projeto (`/plugin install` está rejeitado pela ADR-0025); mods para outros agentes.

## Riscos
| Risco | Mitigação |
|---|---|
| A API de mods do Claude Code é nova e muda por versão | Versão mínima no manifesto; teste do mod no CI a cada release |
| Carregar plugin do projeto sem `/plugin install` | Spike na onda 1 (configuração do projeto ou `--plugin-dir`); se não houver caminho, ADR registra a exceção |
| Workers headless (`claude -p`) podem não carregar o mod | Spike na onda 1; sem carga, o worker segue gravando por `usage.go` (122) |
| Primeira linguagem não Go no repositório (TypeScript) | Fica fora do binário (`ARQ-04`), em `.agents/plugins/forge-painel/`, sem dependência de runtime |
| Telemetria em dobro (worker e mod gravando o mesmo job) | RN-11: dono único por registro, mod mudo com `FORGE_WORKER` |
| Painel e janelas mostrando coisas diferentes | Os dois leem a mesma fonte (barramento e logs); o mod não tem estado próprio além de cache |
| Percentual das janelas só existe em assinatura | Sem janelas, a status line mostra só contexto e custo de tabela |

## Esforço estimado
Entre 2 e 3 dias, incluindo o spike de carregamento.
