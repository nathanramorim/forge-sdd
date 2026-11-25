import { expect, test } from 'claude-code/testing'

import { AGORA, cmd, montarMundo } from './mundo'

const PANE = {
  plugin: 'forge-painel',
  component: 'Pane',
  requestId: 'forge-painel',
  props: {
    title: 'Forge: esteira',
    isFocused: false,
    bodyColumns: 80,
    placement: 'dock',
    scroll: { offset: 0, bodyRows: 20 },
    view: {},
  },
} as const

const INI = '77-auth-oauth'
const job = (id: string, extra: object = {}) =>
  JSON.stringify({ id, type: 'task_build', initiative: INI, feature: 'auth', task_id: '77-01', task_ids: ['77-01', '77-02'], attempt: 1,
    created_at: '2026-11-25T17:50:00Z', updated_at: '2026-11-25T17:55:00Z', ...extra })

const FILAS = '.forge/runtime/queues'
const CENARIO: Record<string, string> = {
  [`${FILAS}/builder/processing/1-1-job-b1.json`]: job('job-b1', { lease: { worker: 'w', expires_at: '2026-11-25T18:10:00Z' }, events: [{ at: '2026-11-25T17:50:00Z', state: 'processing' }] }),
  [`${FILAS}/builder/done/1-0-job-b2.json`]: job('job-b2', { task_id: '77-02', updated_at: '2026-11-25T17:58:00Z' }),
  [`${FILAS}/reviewer/done/1-2-job-r2.json`]: job('job-r2', { type: 'pr_review', task_id: '77-02', updated_at: '2026-11-25T17:59:00Z' }),
  // de outra iniciativa, de dez dias atrás: não entra no quadro
  [`${FILAS}/builder/failed/0-0-job-velho.json`]: job('job-velho', { initiative: '55e0-outra', task_id: '', task_ids: ['55e0-01', '55e0-02'], created_at: '2026-11-15T10:00:00Z', updated_at: '2026-11-15T10:00:00Z' }),
  '.forge/runtime/logs/builder-job-b1.log': '{"type":"system"}\n{"type":"result","usage":{"input_tokens":900,"output_tokens":100}}\n',
  '.forge/runtime/tarefas/77-auth-oauth/feat-77-02.json': '{"task":"feat-77-02","estado":"mesclada"}',
  '.forge/runtime/logs/builder-worker.log': '[forge-builder] Aguardando jobs na fila...\n[builder] 🔨 Editando backend/src/auth.ts...\n  ✎ backend/src/auth.ts\n',
  'sdd/discovery/77-auth-oauth/queue.json': JSON.stringify({ items: [{ task_id: '77-01', status: 'building' }, { task_id: '77-02', status: 'approved' }, { task_id: '77-03', status: 'queued' }] }),
}

test('o painel tem um cartão por estação, como no protótipo: estado, progresso, o que faz agora, tokens e tempo', async ($, on) => {
  const m = montarMundo(on, CENARIO)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })

  expect(await ui.find({ type: 'Text', text: /^Esteira$/ })).toBeDefined()
  expect((await ui.find({ type: 'Text', text: /^iniciativa 77/ }))?.text).toBe('iniciativa 77')
  expect(await ui.find({ type: 'Text', text: /^discovery$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^builder$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^● rodando$/ })).toBeDefined()
  expect((await ui.find({ type: 'Text', text: /tasks/ }))?.text).toBe('▰▰▰▱▱▱▱▱▱▱ 1/3 tasks · 77-01')
  expect((await ui.find({ type: 'Text', text: /Editando/ }))?.text).toBe('Editando backend/src/auth.ts...')
  expect(await ui.find({ type: 'Text', text: /^1,0k tok$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^10m 00s$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^reviewer$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^● aprovado$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^iniciativa$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /55e0/ })).toBeUndefined()
  await ui.unmount()

  // só leitura: nenhuma escrita no barramento nem nos logs
  expect(m.escritas.filter(p => p.startsWith(FILAS) || p.startsWith('.forge/runtime/logs') || p.startsWith('.forge/runtime/tarefas'))).toEqual([])
})

test('falha mostra o motivo e o próximo passo, com o tempo do job e não desde a criação', async ($, on) => {
  montarMundo(on, {
    [`${FILAS}/builder/failed/1-1-job-f.json`]: job('job-f', {
      type: 'feature_build', task_id: '', task_ids: ['77-01', '77-02', '77-03'], created_at: '2026-11-17T11:05:00Z',
      events: [{ at: '2026-11-25T17:00:00Z', state: 'processing' }, { at: '2026-11-25T17:18:03Z', state: 'failed', note: 'o agente saiu com código -1' }],
    }),
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect(await ui.find({ type: 'Text', text: /^● falhou$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^o agente saiu com código -1$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^próximo: \/implementar 77$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^18m 03s$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /anterior à 6.0/ })).toBeUndefined()
  await ui.unmount()
})

test('lease vencido aparece como parado, com o que fazer', async ($, on) => {
  montarMundo(on, {
    [`${FILAS}/builder/processing/1-1-job-b1.json`]: job('job-b1', { lease: { worker: 'w', expires_at: '2026-11-25T17:00:00Z' } }),
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect(await ui.find({ type: 'Text', text: /^● parado$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /rode forge up/ })).toBeDefined()
  await ui.unmount()
})

test('texto do log passa por saneamento (sem escapes de terminal)', async ($, on) => {
  montarMundo(on, {
    ...CENARIO,
    '.forge/runtime/logs/builder-worker.log': '[builder] \u001b[31mvermelho\u001b[0m\u0007 rodando\n',
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  const t = (await ui.find({ type: 'Text', text: /vermelho/ }))?.text ?? ''
  expect(t).toContain('vermelho')
  expect(t).not.toContain('\u001b')
  await ui.unmount()
})

test('sem filas o painel diz como abrir a esteira e o mod não abre sozinho', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  expect(m.abertos).toEqual([])
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect((await ui.find({ type: 'Text', text: /Esteira parada/ }))?.text).toContain('forge up --terminal painel')
  await ui.unmount()
})

test('a status line diz forge, a iniciativa do job e a estação em curso', async ($, on) => {
  const m = montarMundo(on, CENARIO)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.session.measure({ context: { window: 200000, percent: 41 }, rateLimits: [{ kind: 'five_hour', percentUsed: 72 }], changed: ['context' as const] })
  const linha = m.status[m.status.length - 1] ?? ''
  expect(linha).toContain('forge · 77-auth-oauth · builder')
  expect(linha).toContain('│ 5h 72% ▰▰▰▰▰▰▱▱')
  expect(linha).toContain('│ ctx 41%')
  expect(linha).toContain('● telemetria real')
})

test('a faixa acima do prompt mostra o inbox e o PR conforme', async ($, on) => {
  montarMundo(on, {
    ...CENARIO,
    [`${FILAS}/discovery/inbox/n1.json`]: JSON.stringify({ id: 'n1', type: 'pr_approved', pr_url: 'https://github.com/o/r/pull/77', read: false }),
    [`${FILAS}/discovery/inbox/n2.json`]: JSON.stringify({ id: 'n2', type: 'info', read: false }),
    [`${FILAS}/discovery/inbox/n3.json`]: JSON.stringify({ id: 'n3', type: 'info', read: true }),
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  const ui = await $.ui.mount({ plugin: 'forge-painel', component: 'AbovePrompt', surface: 'terminal', props: { hasSurvey: false, isCollapsed: false, isWorking: false } } as never)
  expect((await ui.find({ type: 'Text', text: /inbox/ }))?.text).toContain('forge · inbox 2 · PR #77 conforme')
  await ui.unmount()
})

test('com as filas criadas pelo forge up o painel abre sozinho na sessão interativa', async ($, on) => {
  const m = montarMundo(on, CENARIO)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  expect(m.abertos).toEqual(['forge-painel'])
})

test('/painel log mostra o fim do log sem escapes e recusa id fora do padrão', async ($, on) => {
  montarMundo(on, {
    '.forge/runtime/logs/builder-job-b1.log': 'linha um\n\u001b[32mlinha dois\u001b[0m\n',
    '.forge/runtime/logs/../segredo.log': 'nao',
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })

  const ok = await $.command.run(cmd('painel', 'log builder job-b1'))
  expect(ok.text).toContain('linha um')
  expect(ok.text).toContain('linha dois')
  expect(ok.text).not.toContain('\u001b')

  const ruim = await $.command.run(cmd('painel', 'log builder ../segredo'))
  expect(ruim.text).toContain('Uso:')
  const papel = await $.command.run(cmd('painel', 'log discovery job-b1'))
  expect(papel.text).toContain('Uso:')
})

test('/gastos resume a iniciativa por estação e rotula o dólar como custo de tabela', async ($, on) => {
  const reg = (extra: object) => JSON.stringify({ tokens_input: 1000, tokens_output: 500, medicao: 'real', ...extra })
  montarMundo(on, {
    'sdd/discovery/123-painel/status.md': 'x',
    'sdd/.metrics/schema.json': '{"$schema":"forge-sdd/metrics/1.0"}',
    'sdd/.metrics/session-2026-11-25T170000Z.json': reg({ feature: 'sdd/discovery/123-painel/agentes/builder', agent_path: ['orquestrador', 'builder'], cost_usd_tabela: 1.5 }),
    'sdd/.metrics/session-2026-11-25T170100Z.json': reg({ feature: 'sdd/discovery/123-painel/agentes/revisor', agent_path: ['orquestrador', 'revisor'], cost_usd_tabela: 0.5 }),
    'sdd/.metrics/session-2026-11-25T170200Z.json': reg({ feature: 'sdd/discovery/99-outra/x', agent_path: ['builder'], cost_usd_tabela: 9 }),
    'sdd/.metrics/session-2026-11-25T170300Z.json': reg({ feature: 'sdd/discovery/123-painel/y', agent_path: ['builder'], medicao: 'resumo' }),
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })

  const r = await $.command.run(cmd('gastos'))
  expect(r.text).toContain('iniciativa 123-painel')
  expect(r.text).toContain('- builder: 1 sessões, 1.500 tokens, custo de tabela US$ 1.50')
  expect(r.text).toContain('- revisor: 1 sessões')
  expect(r.text).toContain('Total: 3.000 tokens, custo de tabela US$ 2.00')
  expect(r.text).toContain('não é fatura')
  expect(r.text).not.toContain('9.00')

  const todos = await $.command.run(cmd('gastos', 'todos'))
  expect(todos.text).toContain('Gastos de projeto')
  expect(todos.text).toContain('custo de tabela US$ 11.00')
})

test('/gastos sem registros avisa em vez de falhar', async ($, on) => {
  montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  const r = await $.command.run(cmd('gastos'))
  expect(r.text).toContain('Sem registros')
  expect(AGORA).toBeGreaterThan(0)
})

test('agente em execução ganha a tendência de tokens ao vivo', async ($, on) => {
  const m = montarMundo(on, CENARIO)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  // o log cresce entre duas releituras
  m.arquivos.set('.forge/runtime/logs/builder-job-b1.log', '{"type":"assistant","message":{"usage":{"input_tokens":3000,"output_tokens":800}}}\n')
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect(await ui.find({ type: 'Text', text: /^[▁▂▃▄▅▆▇█]+$/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^3,8k tok$/ })).toBeDefined()
  await ui.unmount()
})
