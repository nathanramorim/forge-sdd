# Guia Técnico: Piloto Automático e Autonomia no Forge-SDD

Este documento consolida as especificações e o guia prático para configurar o Forge-SDD em modo **Autônomo (Auto-Pilot / Hardness Engineer)**, permitindo que a IA execute e revise tarefas em lote por horas, sem necessidade de intervenção humana turno a turno.

---

## 1. Visão Geral do Fluxo Autônomo

O ciclo de desenvolvimento interativo padrão do Forge-SDD foi estruturado para manter o desenvolvedor no controle a cada etapa. Contudo, em cenários de desenvolvimento paralelo ou execução noturna (*overnight building*), o fluxo de parada pode ser opcionalmente desabilitado.

A arquitetura autônoma divide-se em duas camadas:
1. **No Editor / IDE (Antigravity):** Utilização do comando nativo `/goal` do motor da IA para loops contínuos de turnos de construção.
2. **Na CLI / Terminal (Claude / Copilot / Gemini):** Utilização de uma flag física de bypass (`.sdd-auto-pilot`) em conjunto com um script daemon de controle (`sdd-daemon.sh`).

---

## 2. A Flag de Bypass: `.sdd-auto-pilot`

Para que o agente Orquestrador saiba que não deve interromper o fluxo na fase `PLAN` aguardando confirmação (`PLAN: Reporte status + próximas tasks -> aguarde confirmação`), ele verifica a presença do arquivo `.sdd-auto-pilot` na raiz do projeto.

### Regra de Transição no Prompt:
> **Piloto Automático:** Se o arquivo `sdd/.sdd-auto-pilot` ou `.sdd-auto-pilot` estiver presente na raiz, o Orquestrador pula a validação do `PLAN` e assume aprovação tácita, iniciando imediatamente a fase `ACT` (Builder). Ao fechar a sessão, ele atualiza o `progress.md`, comita as alterações e cria o PR (`gh pr create --fill`) sem interrupção.

## 3. O Script Daemon: sdd-daemon.sh

Esta lógica de controle e loop de sessão é gerada automaticamente pelo CLI dentro da pasta `sdd/scripts/sdd-daemon.sh` de cada projeto configurado, permitindo que você personalize as chamadas de agentes e controle de orçamento localmente.

---

## 4. Autonomia com Antigravity (`/goal`)

No Google Antigravity, a autonomia contínua de longa duração é integrada diretamente ao motor de execução através do comando `/goal`.

*   **Uso:** `/goal [descrição macro do objetivo]`
*   **Comportamento:** O motor AGY assume a fila de subtarefas criadas a partir do objetivo e executa sucessivas chamadas de ferramentas de leitura, escrita e execução de terminal sem perguntar ou interromper a sessão, contanto que o progresso não encontre bloqueios insolúveis.

---

## 5. Boas Práticas e Custos de Token

> [!WARNING]
> **Monitoramento de Custos:** Loops contínuos automáticos consomem uma quantidade significativa de tokens em turnos de rework. É recomendado definir um orçamento de limites de chamadas de API de segurança no daemon para evitar surpresas no faturamento de provedores de IA.

> [!IMPORTANT]
> **Budget de Rework:** Limite o rework automático de código do Builder a no máximo **3 turnos** em caso de falha de compilação ou testes. Se falhar repetidamente, marque a feature como `blocked` no progress e notifique o desenvolvedor.
