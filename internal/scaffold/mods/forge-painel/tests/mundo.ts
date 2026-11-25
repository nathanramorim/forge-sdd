// Mundo simulado dos testes: um disco em memória, sob o plugin, no lugar do sistema de arquivos.
// Nada aqui é do mod: serve só para os `*.test.tsx` conferirem o que ele lê e escreve.
import type { On } from 'claude-code'
import { mock } from 'claude-code/testing'

export type Mundo = {
  arquivos: Map<string, string>
  escritas: string[]
  status: (string | undefined)[]
  toasts: string[]
  abertos: string[]
  usage: { startedAt: number; context: { window: number; percent?: number }; rateLimits: unknown[]; cost?: { usd: number } }
  agentes: unknown[]
  agora: number
}

export const AGORA = Date.parse('2026-11-25T18:00:00Z')

// No teste o diretório de trabalho é a pasta do plugin: o motor entrega o caminho absoluto.
const rel = (p: string): string => p.replace(/^.*\/\.agents\/plugins\/forge-painel\//, '')

export function montarMundo(on: On, inicial: Record<string, string> = {}, variaveis: Record<string, string> = {}): Mundo {
  const m: Mundo = {
    arquivos: new Map(Object.entries(inicial)),
    escritas: [],
    status: [],
    toasts: [],
    abertos: [],
    usage: { startedAt: AGORA, context: { window: 200000 }, rateLimits: [] },
    agentes: [],
    agora: AGORA,
  }
  mock.env(on, variaveis)
  mock.clock(on, { now: AGORA })
  // O que o motor responde por baixo: o mod só reage a estes eventos.
  on('session.start', (_, e) => ({ cwd: e.cwd }))
  on('session.measure', (_, e) => ({ changed: e.changed }))
  on('session.end', (_, e) => ({ sessionId: e.sessionId }))
  on('turn.complete', (_, e) => ({ text: e.answer }))
  on('agent.report', (_, e) => ({ text: e.text }))
  on('agent.spawn', () => ({ model: 'm', agentId: 'a1' }))
  on('session.usage', () => ({ value: m.usage as never }))
  on('agent.list', () => ({ value: m.agentes as never }))

  const filhos = (dir: string): { name: string; kind: 'file' | 'dir' }[] => {
    const base = dir.replace(/\/$/, '') + '/'
    const vistos = new Map<string, 'file' | 'dir'>()
    for (const caminho of m.arquivos.keys()) {
      if (!caminho.startsWith(base)) continue
      const resto = caminho.slice(base.length)
      const i = resto.indexOf('/')
      if (i < 0) vistos.set(resto, 'file')
      else vistos.set(resto.slice(0, i), 'dir')
    }
    return [...vistos].map(([name, kind]) => ({ name, kind }))
  }

  on('fs.read', (_, e) => {
    const t = m.arquivos.get(rel(e.path))
    if (t === undefined) throw new Error(`ENOENT ${e.path}`)
    const ini = e.offset ?? 0
    return { value: e.length === undefined ? t.slice(ini) : t.slice(ini, ini + e.length) }
  })
  on('fs.write', (_, e) => {
    m.arquivos.set(rel(e.path), e.text)
    m.escritas.push(rel(e.path))
    return { value: undefined }
  })
  on('fs.exists', (_, e) => ({
    value: m.arquivos.has(rel(e.path)) || filhos(rel(e.path)).length > 0,
  }))
  on('fs.list', (_, e) => {
    const lista = filhos(rel(e.path))
    if (lista.length === 0) throw new Error(`ENOENT ${e.path}`)
    return { value: lista.map(f => ({ ...f, size: 0, mtimeMs: 0, isLink: false })) }
  })
  on('fs.stat', (_, e) => {
    const t = m.arquivos.get(rel(e.path))
    if (t === undefined) throw new Error(`ENOENT ${e.path}`)
    return { value: { kind: 'file', size: t.length, mtimeMs: 1000, isLink: false } }
  })
  on('ui.status', (_, e) => {
    m.status.push(e.text)
    return { value: undefined }
  })
  on('ui.toast', (_, e) => {
    m.toasts.push(e.text)
    return { value: undefined }
  })
  on('command.register', (_, e) => ({ value: { command: e.name } }))
  on('ui.open', (_, e) => {
    m.abertos.push(e.id)
    return { value: { isPlaced: true } }
  })
  return m
}

// Espelho de sdd/.metrics/schema.json (forge-sdd/metrics/1.0): campos obrigatórios e opcionais.
// O teste recusa campo novo (RN-3, DUP-04) e tipo errado.
const OBRIGATORIOS: Record<string, string> = {
  $schema: 'string',
  feature: 'string',
  agent_path: 'array',
  tokens_input: 'integer',
  tokens_output: 'integer',
  turns: 'integer',
  duration_seconds: 'integer',
  criterio_atendido: 'boolean',
  model: 'string',
  rework_lines: 'integer',
  context7_calls: 'integer',
  git_calls: 'integer',
  files_touched: 'array',
  outcome: 'string',
  sdd_version: 'string',
}
const OPCIONAIS: Record<string, string> = {
  phase: 'string',
  tokens_cache_read: 'integer',
  tokens_cache_creation: 'integer',
  cost_usd_tabela: 'number',
  medicao: 'string',
  complexidade: 'null',
  spec: 'object',
}

export function errosDoSchema(texto: string): string[] {
  const erros: string[] = []
  let r: Record<string, unknown>
  try {
    r = JSON.parse(texto)
  } catch {
    return ['JSON inválido']
  }
  const tipo = (v: unknown): string => {
    if (v === null) return 'null'
    if (Array.isArray(v)) return 'array'
    if (typeof v === 'number') return Number.isInteger(v) ? 'integer' : 'number'
    return typeof v
  }
  const conforme = (esperado: string, real: string) => esperado === real || (esperado === 'number' && real === 'integer')
  for (const [k, t] of Object.entries(OBRIGATORIOS)) {
    if (!(k in r)) erros.push(`falta ${k}`)
    else if (!conforme(t, tipo(r[k]))) erros.push(`${k}: esperado ${t}`)
  }
  for (const k of Object.keys(r)) {
    if (k in OBRIGATORIOS) continue
    const t = OPCIONAIS[k]
    if (t === undefined) erros.push(`campo fora do schema: ${k}`)
    else if (!conforme(t, tipo(r[k]))) erros.push(`${k}: esperado ${t}`)
  }
  if (r.$schema !== 'forge-sdd/metrics/1.0') erros.push('$schema errado')
  if (!['approved', 'rejected', 'blocked'].includes(String(r.outcome))) erros.push('outcome inválido')
  if (!['real', 'estimado', 'resumo'].includes(String(r.medicao))) erros.push('medicao inválida')
  return erros
}

// Os testes chamam $.command.run e $.agent.spawn só com o que o mod lê; o resto é do motor.
export const cmd = (command: string, args?: string): never => ({ command, args }) as never
export const spawn = (subagentType: string): never => ({ prompt: 'x', subagentType }) as never
