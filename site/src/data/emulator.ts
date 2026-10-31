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
    stationLabel: 'Bootstrap CLI Go',
    agentName: '01. O Líder (Orquestrador)',
    agentHex: 0xf2a65a,
    agentState: 'finished',
    chipColor: 'var(--ok)',
    chipText: 'Estrutura Pronta',
    speech: 'Instalei os papéis das 5 estações (.agents/roles/), as regras modulares (.agents/rules/) e a Constitution por IDs.',
    logHtml: `
      <div class="cmd-input">$ forge init . --agent claude,gemini --name checkout-api</div>
      <div class="dim">→ Gerando papéis canônicos em .agents/roles/{lider,pm-architect,planner,builder,governance}.md</div>
      <div class="dim">→ Distribuindo regras em .agents/rules/sdd-{arquitetura,seguranca,reuso-e-duplicacao,codigo-e-testes,fluxo-e-portao}.md</div>
      <div class="dim">→ Criando sdd/memory/constitution.md (IDs estáveis: ARQ-*, SEG-*, DUP-*, COD-*, TST-*, FLX-*, GOV-*)</div>
      <div class="ok">✓ Estrutura Forge-SDD v3.0.0-beta criada com sucesso!</div>
    `,
    files: [
      ['.agents/roles/ (5 estações)', 'criado', 'var(--ok)'],
      ['.agents/rules/sdd-*.md', 'criado', 'var(--ok)'],
      ['sdd/memory/constitution.md', 'criado', 'var(--ok)'],
      ['sdd/memory/progress.md', '240 bytes ≤ 1 KB', 'var(--accent)']
    ],
    actionBtn: { label: 'Rodar /discovery no chat →', nextStep: 1 }
  },
  {
    stationLabel: 'Estação 02 · Concepção',
    agentName: '02. Produto & Arquiteto',
    agentHex: 0x8aa4ff,
    agentState: 'approval',
    chipColor: 'var(--warn)',
    chipText: 'Portão FLX-01 (Aguarda Você)',
    speech: 'Mapeei o que já existe para não duplicar código e gerei a Spec. Preciso do seu "De Acordo" antes de acionar o Planejador.',
    logHtml: `
      <div class="cmd-input">&gt; /discovery adicionar pagamento PIX com webhook idempotente</div>
      <div class="accent">[Estação 02 · PM &amp; Arquiteto] Executando Reuse Check obrigatório (DUP-01 / DUP-02)…</div>
      <div class="dim">  • Reaproveitado: internal/http/middleware/hmac.go e internal/db/tx.go</div>
      <div class="ok">✓ Gerado: sdd/discovery/discovery-01-pagamento-pix.md</div>
      <div class="ok">✓ Gerado: sdd/discovery/criteria-01-pagamento-pix.md (CA-01..CA-05)</div>
      <div class="warn">! PORTÃO HUMANO FLX-01: Revise a Spec antes de prosseguir para o plano técnico.</div>
    `,
    files: [
      ['sdd/discovery/discovery-01-pagamento-pix.md', 'Reuse Check OK', 'var(--ok)'],
      ['sdd/discovery/criteria-01-pagamento-pix.md', 'aguarda seu OK', 'var(--warn)'],
      ['sdd/discovery/plan-01-*.md', 'bloqueado (FLX-01)', 'var(--faint)']
    ],
    actionBtn: { label: 'Aprovar Spec (FLX-01) e Planejar →', nextStep: 2, style: 'warn' }
  },
  {
    stationLabel: 'Estação 03 · Planejamento & Fila',
    agentName: '03. O Planejador',
    agentHex: 0xc9a7ff,
    agentState: 'thinking',
    chipColor: 'var(--accent)',
    chipText: 'Fila Calculada',
    speech: 'Dividi a Spec em 3 tarefas: 01 e 02 não tocam nos mesmos arquivos e vão rodar em paralelo; a 03 aguarda na fila.',
    logHtml: `
      <div class="cmd-input">&gt; /split-features</div>
      <div class="accent">[Estação 03 · Planejador] Vinculando IDs da Constitution e rules por subtarefa (GOV-04)…</div>
      <div class="dim">  • feat-01-01 (P): Domínio PIX [ARQ-01, TST-01] → sem conflito (Onda Paralela 1)</div>
      <div class="dim">  • feat-01-02 (M): Cliente Banco [SEG-01, TST-01] → sem conflito (Onda Paralela 1)</div>
      <div class="dim">  • feat-01-03 (M): Webhook Handler [SEG-02, DUP-01] → depends_on: [01, 02] (Onda 2)</div>
      <div class="cmd-input">$ forge-sdd queue enqueue sdd/features/feat-01-pagamento-pix/*.md</div>
      <div class="ok">✓ 3 tarefas gravadas em sdd/.queue/queue.json (0 colisões na Onda 1)</div>
    `,
    files: [
      ['sdd/discovery/plan-01-pagamento-pix.md', 'gerado', 'var(--ok)'],
      ['sdd/.queue/queue.json (feat-01-01 & 02)', 'queued (paralelo)', 'var(--accent)'],
      ['sdd/.queue/queue.json (feat-01-03)', 'queued (depends_on)', 'var(--warn)']
    ],
    actionBtn: { label: 'Disparar Builders (/proxima-feature) →', nextStep: 3 }
  },
  {
    stationLabel: 'Estação 04 · Execução & Portão A',
    agentName: '04. Fila de Builders',
    agentHex: 0x4fd1b5,
    agentState: 'working',
    chipColor: 'var(--ok)',
    chipText: 'Portão A: PASS',
    speech: 'Executamos as tarefas da fila respeitando as dependências e validamos tudo no Portão Determinístico Camada A!',
    logHtml: `
      <div class="cmd-input">&gt; /proxima-feature</div>
      <div class="accent">[Builder A &amp; Builder B] Executando Onda 1 em paralelo (feat-01-01 + feat-01-02)…</div>
      <div class="dim">[Builder A] concluído → liberando dependência de feat-01-03…</div>
      <div class="cmd-input">$ forge-sdd gate --mode bloqueio --branch feat/01-pagamento-pix</div>
      <div class="ok">  ✓ [P-01] Branch feat/01-pagamento-pix válida</div>
      <div class="ok">  ✓ [P-13] Zero segredos expostos no diff</div>
      <div class="ok">  ✓ [P-23] go test ./... &amp; go vet ./... (100% PASS)</div>
    `,
    files: [
      ['feat-01-01 & feat-01-02', 'done (Portão A PASS)', 'var(--ok)'],
      ['feat-01-03 (Webhook Handler)', 'done (Portão A PASS)', 'var(--ok)'],
      ['sdd/memory/progress.md', '712 bytes ≤ 1024 B', 'var(--ok)']
    ],
    actionBtn: { label: 'Rodar Auditoria Read-Only (/revisar) →', nextStep: 4 }
  },
  {
    stationLabel: 'Estação 05 → 01 · Governança & Líder',
    agentName: '05. Governança → 01. Líder',
    agentHex: 0x3ecf8e,
    agentState: 'finished',
    chipColor: 'var(--ok)',
    chipText: 'Veredito: Conforme',
    speech: 'Auditoria Read-Only aprovada sem ressalvas! O Líder consolidou seu Painel Executivo e sincronizou 1 nova Auto-Skill.',
    logHtml: `
      <div class="cmd-input">&gt; /revisar</div>
      <div class="accent">[Estação 05 · Governança Read-Only] Auditando diff contra ARQ-01, SEG-01, DUP-01, TST-01…</div>
      <div class="ok">✓ Parecer Camada B: {"veredito": "conforme", "bloqueios": [], "auto_skills_candidatas": ["webhook-hmac-idempotente"]}</div>
      <div class="cmd-input">$ forge-sdd skills approve webhook-hmac-idempotente</div>
      <div class="ok">✓ Auto-Skill sincronizada em .agents/skills/, .claude/skills/ e .gemini/skills/</div>
      <div class="ok">✓ Painel Executivo do Líder (Camada C): 3/3 tarefas aprovadas · 0 pendências na fila!</div>
    `,
    files: [
      ['sdd/.queue/queue.json (3/3 tarefas)', 'status: approved', 'var(--ok)'],
      ['.agents/skills/webhook-hmac-idempotente/', 'Auto-Skill ativa', 'var(--ok)'],
      ['Painel Executivo (Camada C)', 'entregue ao humano', 'var(--ok)']
    ],
    actionBtn: { label: 'Reiniciar demonstração ↺', nextStep: 0, style: 'ghost' }
  }
];
