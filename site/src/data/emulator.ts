import type { MascotState } from './stations';

export interface EmuFileItem {
  name: string;
  badge: string;
  color: string;
}

export interface EmuActionBtn {
  label: string;
  nextStep: number;
  style?: string;
}

export interface EmuStepData {
  stationLabel: string;
  agentName: string;
  agentHex: number;
  agentState: MascotState;
  chipColor: string;
  chipText: string;
  speech: string;
  logHtml: string;
  files: [string, string, string][];
  actionBtn: EmuActionBtn;
}

export const EMU_STEPS: EmuStepData[] = [
  {
    stationLabel: 'Menu forge · Início',
    agentName: 'Você + o menu do forge',
    agentHex: 0xf2a65a,
    agentState: 'finished',
    chipColor: 'var(--ok)',
    chipText: '3 sessões abertas',
    speech: 'Escolhi o agente de cada sessão no menu, conferi a checagem e abri tudo numa tela só. A última escolha fica lembrada.',
    logHtml: `
      <div class="cmd-input">$ forge</div>
      <div class="accent">Forge SDD 4.4.0-beta — o que você quer fazer?  ▸ Pipeline multi-sessão</div>
      <div class="dim">→ Iniciar as 3 sessões: forge-discovery claude · forge-builder codex · forge-reviewer gemini</div>
      <div class="dim">Checagem antes de abrir: ✓ claude  ✓ codex  ✓ gemini  ✓ arquivos de instrução</div>
      <div class="ok">✓ Sessão tmux forge-checkout-api-1a2b3c aberta: 3 painéis nomeados</div>
    `,
    files: [
      ['.forge/config/forge.json', 'agentes lembrados', 'var(--ok)'],
      ['.forge/runtime/ (fila e logs)', 'fora do git', 'var(--accent)'],
      ['sdd/discovery/<id>-<nome>/', 'uma pasta por iniciativa', 'var(--accent)']
    ],
    actionBtn: { label: 'Conversar na Discovery (/discovery) →', nextStep: 1 }
  },
  {
    stationLabel: 'Discovery · Concepção (você)',
    agentName: 'Produto & Arquiteto',
    agentHex: 0x8aa4ff,
    agentState: 'approval',
    chipColor: 'var(--warn)',
    chipText: 'Portão FLX-01 (Aguarda Você)',
    speech: 'Mapeei o que já existe para não duplicar código e gerei a Spec. Preciso do seu "De Acordo" antes de fatiar em tarefas.',
    logHtml: `
      <div class="cmd-input">&gt; /discovery adicionar pagamento PIX com webhook idempotente</div>
      <div class="accent">[Discovery] Carregando ADRs e rules antes de propor (nada pode violá-las)…</div>
      <div class="dim">  • Reaproveitado: internal/http/middleware/hmac.go e internal/db/tx.go</div>
      <div class="ok">✓ Gerado: sdd/discovery/12-pagamento-pix/discovery.md e criteria.md (CA-01..CA-05)</div>
      <div class="ok">✓ Gerado: plan.md e status.md da iniciativa</div>
      <div class="warn">! PORTÃO HUMANO FLX-01: revise a Spec antes de fatiar.</div>
    `,
    files: [
      ['sdd/discovery/12-pagamento-pix/discovery.md', 'Reuse Check OK', 'var(--ok)'],
      ['sdd/discovery/12-pagamento-pix/criteria.md', 'aguarda seu OK', 'var(--warn)'],
      ['sdd/discovery/12-pagamento-pix/status.md', '≤ 1 KB', 'var(--accent)']
    ],
    actionBtn: { label: 'Aprovar Spec (FLX-01) e fatiar →', nextStep: 2, style: 'warn' }
  },
  {
    stationLabel: 'Discovery → Builder · Entrega',
    agentName: 'Planejador → fila do Builder',
    agentHex: 0xc9a7ff,
    agentState: 'thinking',
    chipColor: 'var(--accent)',
    chipText: 'Entregue ao Builder',
    speech: 'Fatiei em 3 tarefas e já entreguei ao Builder. Esta sessão não constrói: ela só acompanha pelo inbox.',
    logHtml: `
      <div class="cmd-input">&gt; /split-features</div>
      <div class="dim">  • feat-12-01 Domínio PIX [ARQ-01, TST-01]  → Onda 1 (paralelo)</div>
      <div class="dim">  • feat-12-02 Cliente do banco [SEG-01]     → Onda 1 (paralelo)</div>
      <div class="dim">  • feat-12-03 Webhook idempotente [SEG-02]  → Onda 2 (depende de 01 e 02)</div>
      <div class="cmd-input">$ forge-sdd dispatch pagamento-pix</div>
      <div class="ok">✓ Gate pré-build: spec, ADRs e rules conferidas</div>
      <div class="ok">✓ Job enfileirado em builder/pending. Entregue ao Builder.</div>
    `,
    files: [
      ['sdd/discovery/12-pagamento-pix/features/ (3 specs)', 'geradas', 'var(--ok)'],
      ['sdd/discovery/12-pagamento-pix/queue.json', '3 tasks na fila', 'var(--accent)'],
      ['.forge/runtime/queues/builder/pending/', '1 job', 'var(--warn)']
    ],
    actionBtn: { label: 'Ver o painel do Builder →', nextStep: 3 }
  },
  {
    stationLabel: 'forge-builder · Execução',
    agentName: 'forge-builder (worker)',
    agentHex: 0x4fd1b5,
    agentState: 'working',
    chipColor: 'var(--ok)',
    chipText: 'Construindo em paralelo',
    speech: 'Peguei o job da fila e estou construindo. O painel mostra o que faço, em linhas curtas; o log completo fica à parte.',
    logHtml: `
      <div class="dim">[forge-builder] agente: codex · permissões: PROJETO</div>
      <div class="accent">▶ job pagamento-pix · feature_build · tentativa 1 · codex · permissões project</div>
      <div class="ok">  ✎ internal/pix/domain.go</div>
      <div class="ok">  ✎ internal/pix/client.go</div>
      <div class="dim">  $ go test ./...</div>
      <div class="ok">  ✓ comando terminou (0)</div>
      <div class="dim">  … trabalhando há 2m10s · 14 arquivos alterados</div>
      <div class="ok">■ concluído em 4m12s · PR aberto e enfileirado para o Reviewer</div>
    `,
    files: [
      ['builder/processing → done', 'job concluído', 'var(--ok)'],
      ['PR #12 (feat/12-pagamento-pix)', 'aberto', 'var(--ok)'],
      ['.forge/runtime/queues/reviewer/pending/', '1 job', 'var(--warn)']
    ],
    actionBtn: { label: 'Ver o painel do Reviewer →', nextStep: 4 }
  },
  {
    stationLabel: 'forge-reviewer → você · Revisão',
    agentName: 'forge-reviewer → você',
    agentHex: 0x3ecf8e,
    agentState: 'finished',
    chipColor: 'var(--ok)',
    chipText: 'Conforme · aguarda sua aprovação',
    speech: 'Reprovei uma vez por violar uma ADR, o Builder corrigiu e agora está conforme. Quem aprova e mescla o PR é você.',
    logHtml: `
      <div class="accent">▶ job pagamento-pix · pr_review · tentativa 1 · gemini</div>
      <div class="warn">  ✗ ADR-0003 violada em internal/pix/client.go:42</div>
      <div class="dim">↺ devolvido ao Builder (rework, prioridade alta, tentativa 2 de 3)</div>
      <div class="accent">▶ job pagamento-pix · pr_review · tentativa 2 · gemini</div>
      <div class="ok">■ conforme: ADRs e rules respeitadas</div>
      <div class="ok">✓ Inbox da forge-discovery: PR revisado, sem inconformidades, pronto para a sua aprovação</div>
    `,
    files: [
      ['PR #12', 'revisado, sem inconformidades', 'var(--ok)'],
      ['Inbox da Discovery (/inbox)', 'aviso com o link do PR', 'var(--ok)'],
      ['Aprovar e mesclar', 'só você', 'var(--warn)']
    ],
    actionBtn: { label: 'Reiniciar demonstração ↺', nextStep: 0, style: 'ghost' }
  }
];
