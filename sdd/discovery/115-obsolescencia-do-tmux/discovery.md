# Discovery — 115 Pipeline contínuo e seleção de modelos (tmux opcional)

## Prioridade atualizada pelo dono

O principal deve continuar coordenando os agentes. PM-Architect classifica complexidade e consulta um Advisor de raciocínio apenas quando há incerteza arquitetural relevante. O split marca o tier do Builder por task. Tmux passa a ser opcional e não bloqueia esta entrega.

## Problema

O projeto permite fixar modelos por papel, mas o job do Builder não carrega modelo por task. A sessão de Discovery não orienta quando uma demanda bem definida pode seguir com um modelo simples, nem quando usar um modelo de raciocínio. Assim, o Builder recebe apenas a configuração estática do papel.

## Jornada desejada

1. O agente principal aciona PM-Architect com brief e contexto.
2. PM-Architect classifica cada task: simples e clara usa tier rápido; maior complexidade usa tier equilibrado; alta incerteza arquitetural pode chamar Advisor no tier de raciocínio.
3. O split registra complexidade, justificativa e tier recomendado em cada task. Para Claude, Haiku é o padrão das tasks bem definidas, Sonnet atende tasks mais complexas e Opus fica como Advisor.
4. Depois de spec/plano aprovados e gates satisfeitos, o job do Builder recebe o tier escolhido e o worker resolve o modelo conforme a CLI configurada.
5. O Builder implementa e valida o código. O Reviewer só recebe o trabalho após Camada A e registro do PR.
6. Cada worker atende somente a iniciativa para a qual foi iniciado, inclusive o ciclo de review/rework, e encerra quando ela drena. Jobs ativos concluem normalmente.

**Interpretação de “enviar com o código pronto e validado”:** o handoff após código validado é para o Reviewer/PR. O Builder continua recebendo a spec aprovada antes de escrever o código; o fluxo atual já executa Camada A antes de enfileirar o Reviewer.

## Escopo

**Dentro:** rubrica de complexidade; chamada condicional ao Advisor; tiers neutros de CLI; tier registrado na task e no job; resolução do modelo no worker; workers escopados por iniciativa; handoff validado ao Reviewer.

**Opcional:** retirar tmux. A remoção pode seguir em outra entrega e não bloqueia o pipeline ou o roteamento de modelos.

**Fora:** instalar CLIs/modelos, alterar contas dos provedores, buscar catálogos pela rede ou aplicar modelo de raciocínio a todo trabalho.

## Regras de negócio

- Tasks novas registram complexidade, justificativa e tier do Builder.
- Tasks claras usam o tier rápido; tier de raciocínio é reservado ao Advisor quando houver alta incerteza.
- O job carrega o tier; a CLI resolve seu próprio modelo. Nunca usar silenciosamente o nome de modelo de outro provedor.
- Jobs antigos sem tier continuam usando o modelo fixo por papel ou inherit.
- O Advisor não escreve código nem substitui Gate 1.
- O pipeline em execução não é interrompido para aplicar a política nova.

## Métricas de sucesso

Jobs novos registram tier e modelo executado; tasks simples não usam tier de raciocínio sem justificativa; jobs de outra iniciativa permanecem intactos; Reviewer só recebe código após Gate A.

## A confirmar

A interpretação do handoff após código pronto está registrada na jornada e aguarda Gate 1.
