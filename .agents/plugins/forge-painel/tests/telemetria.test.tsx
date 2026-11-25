import { expect, test } from 'claude-code/testing'

import { AGORA, errosDoSchema, montarMundo, spawn } from './mundo'

const USO = {
  input_tokens: 1000,
  output_tokens: 200,
  cache_read_input_tokens: 50,
  cache_creation_input_tokens: 10,
  model: 'claude-teste',
}
const agente = (id: string, tipo = 'builder') => ({
  id,
  type: tipo,
  status: 'completed',
  startTime: AGORA - 60_000,
  model: 'claude-teste',
  tokenCount: 1200,
  usage: [USO],
  cost: { usd: 0.1234 },
})
const relatorio = (id: string, tipo = 'builder') => ({
  agentId: id,
  subagentType: tipo,
  provider: { plugin: 'engine', tier: 'core' } as const,
  status: 'completed' as const,
  text: 'feito',
})
const SDDRC = JSON.stringify({ version: '6.0.0-beta', telemetry: { enabled: true } })
const metricas = (m: { escritas: string[] }) => m.escritas.filter(p => p.startsWith('sdd/.metrics/'))
const turno = (id: string, extra: object = {}) => ({
  answer: 'a',
  durationMs: 10,
  isAborted: false,
  turnId: id,
  reason: 'answer' as const,
  usage: USO,
  ...extra,
})
const fim = (reason: 'prompt_input_exit' | 'other') => ({ reason, sessionId: 's1', resume: {} as never })

test('agent.report grava um registro válido no schema 1.0 com medicao real', async ($, on) => {
  const m = montarMundo(on, { 'sdd/.sddrc': SDDRC, 'sdd/discovery/123-painel/status.md': 'x' })
  m.agentes = [agente('ag1')]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.agent.report(relatorio('ag1'))

  const nomes = metricas(m)
  expect(nomes).toHaveLength(1)
  expect(nomes[0]).toBe('sdd/.metrics/session-2026-11-25T180000Z.json')
  const texto = m.arquivos.get(nomes[0] ?? '') ?? ''
  expect(errosDoSchema(texto)).toEqual([])
  const r = JSON.parse(texto)
  expect(r.medicao).toBe('real')
  expect(r.tokens_input).toBe(1000)
  expect(r.tokens_output).toBe(200)
  expect(r.tokens_cache_read).toBe(50)
  expect(r.duration_seconds).toBe(60)
  expect(r.cost_usd_tabela).toBe(0.1234)
  expect(r.agent_path).toEqual(['orquestrador', 'builder'])
  expect(r.phase).toBe('123')
  expect(r.sdd_version).toBe('6.0.0-beta')
})

test('o mesmo agente não é gravado duas vezes; outro no mesmo segundo ganha sufixo', async ($, on) => {
  const m = montarMundo(on, { 'sdd/.sddrc': SDDRC })
  m.agentes = [agente('ag1'), agente('ag2', 'revisor')]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.agent.report(relatorio('ag1'))
  await $.agent.report(relatorio('ag1'))
  await $.agent.report(relatorio('ag2', 'revisor'))

  expect(metricas(m)).toEqual([
    'sdd/.metrics/session-2026-11-25T180000Z.json',
    'sdd/.metrics/session-2026-11-25T180000Z-2.json',
  ])
})

test('sem uso medido do agente nada é gravado (nunca estima)', async ($, on) => {
  const m = montarMundo(on, { 'sdd/.sddrc': SDDRC })
  m.agentes = [{ ...agente('ag1'), usage: undefined }]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.agent.report(relatorio('ag1'))
  expect(metricas(m)).toHaveLength(0)
})

test('telemetria desligada no .sddrc: nada é gravado', async ($, on) => {
  const m = montarMundo(on, { 'sdd/.sddrc': JSON.stringify({ version: '6', telemetry: { enabled: false } }) })
  m.agentes = [agente('ag1')]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.agent.report(relatorio('ag1'))
  expect(metricas(m)).toHaveLength(0)
})

test('a sessão interativa grava um registro ao fim, com o uso somado dos turnos', async ($, on) => {
  const m = montarMundo(on, { 'sdd/.sddrc': SDDRC })
  m.usage.cost = { usd: 2.5 }
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.turn.complete(turno('t1'))
  await $.turn.complete(turno('t2'))
  // turno de subagente não conta na sessão principal
  await $.turn.complete(turno('t3', { agentId: 'ag1' }))
  await $.session.end(fim('prompt_input_exit'))

  const nomes = metricas(m)
  expect(nomes).toHaveLength(1)
  const texto = m.arquivos.get(nomes[0] ?? '') ?? ''
  expect(errosDoSchema(texto)).toEqual([])
  const r = JSON.parse(texto)
  expect(r.turns).toBe(2)
  expect(r.tokens_input).toBe(2000)
  expect(r.agent_path).toEqual(['orquestrador'])
  expect(r.cost_usd_tabela).toBe(2.5)
  expect(r.medicao).toBe('real')
})

test('sessão não interativa (claude -p) não grava registro', async ($, on) => {
  const m = montarMundo(on, { 'sdd/.sddrc': SDDRC })
  await $.session.start({ cwd: '/', surface: null, isInteractive: false })
  await $.turn.complete(turno('t1'))
  await $.session.end(fim('other'))
  expect(metricas(m)).toHaveLength(0)
})

test('com FORGE_WORKER o mod não grava, não mostra status e não abre painel', async ($, on) => {
  const m = montarMundo(
    on,
    { 'sdd/.sddrc': SDDRC, '.forge/runtime/queues/builder/pending/x.json': '{}' },
    { FORGE_WORKER: 'forge-builder' },
  )
  m.agentes = [agente('ag1')]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.session.measure({
    context: { window: 1, percent: 99 },
    rateLimits: [{ kind: 'five_hour', percentUsed: 99 }],
    changed: ['rateLimits'],
  })
  await $.turn.complete(turno('t1'))
  await $.agent.report(relatorio('ag1'))
  await $.session.end(fim('other'))

  expect(m.escritas).toEqual([])
  expect(m.status).toEqual([])
  expect(m.toasts).toEqual([])
  expect(m.abertos).toEqual([])
})

test('sem FORGE_WORKER, o registro de job de um worker já existente fica intacto, sem duplicata', async ($, on) => {
  const jobPronto = '{"feature":"spec.md","medicao":"real","agent_path":["builder"]}'
  const m = montarMundo(on, {
    'sdd/.sddrc': SDDRC,
    'sdd/.metrics/session-2026-11-25T175900Z.json': jobPronto,
  })
  m.agentes = [agente('ag1')]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.agent.report(relatorio('ag1'))
  await $.agent.report(relatorio('ag1'))

  expect(m.arquivos.get('sdd/.metrics/session-2026-11-25T175900Z.json')).toBe(jobPronto)
  expect(metricas(m)).toEqual(['sdd/.metrics/session-2026-11-25T180000Z.json'])
})

test('aviso de orçamento em agent.spawn acima do limite não bloqueia', async ($, on) => {
  const m = montarMundo(on)
  m.usage.rateLimits = [{ kind: 'five_hour', percentUsed: 85, resetsAt: '2026-11-25T21:30:00Z' }]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  const r = await $.agent.spawn(spawn('builder'))

  expect(m.toasts).toHaveLength(1)
  expect(m.toasts[0]).toContain('5h')
  expect(m.toasts[0]).toContain('85%')
  expect(r.deny).toBeUndefined()
  expect(r.agentId).toBe('a1')
})

test('abaixo do limite o spawn passa sem aviso', async ($, on) => {
  const m = montarMundo(on)
  m.usage.rateLimits = [{ kind: 'five_hour', percentUsed: 40 }]
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.agent.spawn(spawn('builder'))
  expect(m.toasts).toHaveLength(0)
})
