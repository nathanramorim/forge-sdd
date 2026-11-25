import { expect, test } from 'claude-code/testing'

import { errosDoSchema, montarMundo } from './mundo'

const ISO = '2026-11-25T21:30:00Z'
const medida = (cinco: number, sete: number, contexto = 35) => ({
  context: { window: 200000, percent: contexto },
  rateLimits: [
    { kind: 'five_hour', percentUsed: cinco, resetsAt: ISO },
    { kind: 'seven_day', percentUsed: sete, resetsAt: '2026-11-29T10:00:00Z' },
  ],
  changed: ['rateLimits' as const],
})

test('a status line mostra iniciativa, janelas e contexto', async ($, on) => {
  const m = montarMundo(on, { 'sdd/discovery/123-painel/status.md': 'x' })
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.session.measure(medida(42, 12))

  const linha = m.status[m.status.length - 1] ?? ''
  expect(linha).toContain('123-painel')
  expect(linha).toContain('5h 42%')
  expect(linha).toContain('7d 12%')
  expect(linha).toContain('ctx 35%')
})

test('sem janelas (fora de assinatura) mostra só contexto e custo de tabela', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.session.measure({
    context: { window: 200000, percent: 20 },
    rateLimits: [],
    cost: { usd: 1.234 },
    changed: ['context'],
  })

  const linha = m.status[m.status.length - 1] ?? ''
  expect(linha).toContain('ctx 20%')
  expect(linha).toContain('custo de tabela US$ 1.23')
  expect(linha).not.toContain('5h')
  expect(linha).not.toContain('7d')
})

test('alerta uma vez ao cruzar 70% e uma vez ao cruzar 90%, por janela', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })

  await $.session.measure(medida(60, 10))
  expect(m.toasts).toHaveLength(0)
  await $.session.measure(medida(71, 10))
  expect(m.toasts).toHaveLength(1)
  expect(m.toasts[0]).toContain('5h')
  expect(m.toasts[0]).toMatch(/renova \d\d:\d\d/)
  await $.session.measure(medida(75, 10))
  expect(m.toasts).toHaveLength(1)
  await $.session.measure(medida(91, 10))
  expect(m.toasts).toHaveLength(2)
  expect(m.toasts[1]).toContain('90%')
  await $.session.measure(medida(95, 10))
  expect(m.toasts).toHaveLength(2)
  // a janela de 7d é independente
  await $.session.measure(medida(95, 72))
  expect(m.toasts).toHaveLength(3)
  expect(m.toasts[2]).toContain('7d')
})

test('salto direto a 90% avisa só o 90, e janela renovada avisa de novo', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })

  await $.session.measure(medida(92, 10))
  expect(m.toasts).toHaveLength(1)
  expect(m.toasts[0]).toContain('90%')

  const nova = medida(75, 10)
  nova.rateLimits[0] = { kind: 'five_hour', percentUsed: 75, resetsAt: '2026-11-26T02:30:00Z' }
  await $.session.measure(nova)
  expect(m.toasts).toHaveLength(2)
  expect(m.toasts[1]).toContain('70%')
})

test('só grava em sdd/.metrics e .forge/runtime/painel', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.session.measure(medida(95, 95))

  expect(m.escritas.length).toBeGreaterThan(0)
  for (const caminho of m.escritas) {
    expect(caminho.startsWith('.forge/runtime/painel/') || caminho.startsWith('sdd/.metrics/')).toBe(true)
  }
})

test('grava a janela para o forge decidir o teto de paralelo', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.session.measure(medida(87.44, 40))

  const txt = m.arquivos.get('.forge/runtime/painel/janela.json') ?? ''
  const j = JSON.parse(txt)
  expect(j.five_hour_pct).toBe(87.4)
  expect(j.seven_day_pct).toBe(40)
  expect(typeof j.at).toBe('string')
})

test('sem janelas de tempo não grava janela.json', async ($, on) => {
  const m = montarMundo(on)
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })
  await $.session.measure({ context: { window: 200000, percent: 10 }, rateLimits: [], changed: ['context' as const] })
  expect(m.arquivos.has('.forge/runtime/painel/janela.json')).toBe(false)
})
