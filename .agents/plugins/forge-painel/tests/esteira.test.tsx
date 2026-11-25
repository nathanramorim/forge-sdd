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

const job = (id: string, extra: object = {}) =>
  JSON.stringify({ id, feature: 'feat-123-02', task_ids: ['2.1'], attempt: 1, updated_at: '2026-11-25T17:50:00Z', ...extra })

const FILAS = '.forge/runtime/queues'
const CENARIO: Record<string, string> = {
  [`${FILAS}/builder/processing/1-1-job-b1.json`]: job('job-b1', { lease: { worker: 'w', expires_at: '2026-11-25T18:10:00Z' } }),
  [`${FILAS}/reviewer/done/1-2-job-r1.json`]: job('job-r1', { task_ids: ['2.2'] }),
  '.forge/runtime/logs/builder-job-b1.log':
    '{"type":"system"}\n{"type":"result","usage":{"input_tokens":900,"output_tokens":100}}\n',
}

test('o painel mostra um job em execução e um concluído, com papel, task e estado', async ($, on) => {
  const m = montarMundo(on, CENARIO)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })

  const b = await ui.find({ type: 'Text', text: /builder 2\.1/ })
  expect(b?.text).toContain('em execução')
  expect(b?.text).toContain('1000 tokens')
  const r = await ui.find({ type: 'Text', text: /reviewer 2\.2/ })
  expect(r?.text).toContain('concluído')
  await ui.unmount()

  // só leitura: nenhuma escrita no barramento nem nos logs
  expect(m.escritas.filter(p => p.startsWith(FILAS) || p.startsWith('.forge/runtime/logs'))).toEqual([])
  expect(m.arquivos.get(`${FILAS}/builder/processing/1-1-job-b1.json`)).toBe(CENARIO[`${FILAS}/builder/processing/1-1-job-b1.json`])
})

test('lease vencido aparece como parado', async ($, on) => {
  montarMundo(on, {
    [`${FILAS}/builder/processing/1-1-job-b1.json`]: job('job-b1', { lease: { worker: 'w', expires_at: '2026-11-25T17:00:00Z' } }),
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect((await ui.find({ type: 'Text', text: /builder/ }))?.text).toContain('parado')
  await ui.unmount()
})

test('texto do job passa por saneamento (sem escapes de terminal)', async ($, on) => {
  montarMundo(on, {
    [`${FILAS}/builder/pending/1-1-job-b1.json`]: job('job-b1', { task_ids: ['\u001b[31mvermelho\u001b[0m\u0007'] }),
  })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  const t = (await ui.find({ type: 'Text', text: /vermelho/ }))?.text ?? ''
  expect(t).toContain('builder vermelho')
  expect(t).not.toContain('\u001b')
  await ui.unmount()
})

test('sem filas o painel diz como abrir a esteira e o mod não abre sozinho', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  expect(m.abertos).toEqual([])
  await $.command.run(cmd('painel'))
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect((await ui.find({ type: 'Text', text: /Nenhum job/ }))?.text).toContain('forge up --terminal painel')
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
