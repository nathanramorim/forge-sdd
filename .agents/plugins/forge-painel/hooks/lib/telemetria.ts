// Registro de telemetria no schema existente `forge-sdd/metrics/1.0` (RN-3, DUP-04): nenhum
// campo novo. Só ids, contagens e nomes de papel entram (RN-8, SEG-05): nada de texto de agente.

export type Uso = {
  input_tokens: number
  output_tokens: number
  cache_read_input_tokens: number
  cache_creation_input_tokens: number
  model?: string
}

export type Entrada = {
  feature: string
  phase?: string
  agentPath: string[]
  uso: Uso[]
  turns: number
  durationSeconds: number
  model: string
  custoTabelaUsd?: number
  ok: boolean
  sddVersion: string
}

export type Registro = {
  $schema: 'forge-sdd/metrics/1.0'
  feature: string
  phase?: string
  agent_path: string[]
  tokens_input: number
  tokens_output: number
  turns: number
  duration_seconds: number
  criterio_atendido: boolean
  model: string
  rework_lines: number
  context7_calls: number
  git_calls: number
  files_touched: string[]
  outcome: 'approved' | 'blocked'
  sdd_version: string
  tokens_cache_read: number
  tokens_cache_creation: number
  cost_usd_tabela?: number
  medicao: 'real'
  complexidade: null
}

function soma(uso: Uso[], campo: keyof Omit<Uso, 'model'>): number {
  let total = 0
  for (const u of uso) total += u[campo] ?? 0
  return total
}

export function montarRegistro(e: Entrada): Registro {
  const r: Registro = {
    $schema: 'forge-sdd/metrics/1.0',
    feature: e.feature,
    agent_path: e.agentPath,
    tokens_input: soma(e.uso, 'input_tokens'),
    tokens_output: soma(e.uso, 'output_tokens'),
    turns: e.turns,
    duration_seconds: Math.max(0, Math.round(e.durationSeconds)),
    criterio_atendido: e.ok,
    model: e.model,
    rework_lines: 0,
    context7_calls: 0,
    git_calls: 0,
    files_touched: [],
    outcome: e.ok ? 'approved' : 'blocked',
    sdd_version: e.sddVersion,
    tokens_cache_read: soma(e.uso, 'cache_read_input_tokens'),
    tokens_cache_creation: soma(e.uso, 'cache_creation_input_tokens'),
    medicao: 'real',
    complexidade: null,
  }
  if (e.phase) r.phase = e.phase
  // O Claude Code estima o dólar a preço de tabela (RN-4): vai em cost_usd_tabela, nunca como fatura.
  if (e.custoTabelaUsd !== undefined && e.custoTabelaUsd > 0) {
    r.cost_usd_tabela = Math.round(e.custoTabelaUsd * 10000) / 10000
  }
  return r
}

// Papel do subagente pelo tipo: nome de papel, nunca descrição livre.
export function papelDoSubagente(tipo: string): string[] {
  const t = tipo.toLowerCase()
  if (/review|revis/.test(t)) return ['orquestrador', 'revisor']
  if (/build|implement/.test(t)) return ['orquestrador', 'builder']
  return ['orquestrador']
}

function dois(n: number): string {
  return n < 10 ? `0${n}` : String(n)
}

// session-<ISO8601 sem separadores>.json, o mesmo formato de RecordSessionMetrics (UTC).
export function nomeBase(agora: number): string {
  const d = new Date(agora)
  const iso =
    `${d.getUTCFullYear()}-${dois(d.getUTCMonth() + 1)}-${dois(d.getUTCDate())}` +
    `T${dois(d.getUTCHours())}${dois(d.getUTCMinutes())}${dois(d.getUTCSeconds())}Z`
  return `session-${iso}`
}

export const DIR_METRICAS = 'sdd/.metrics'

// Slug de iniciativa `<id>-<nome>` (mesma regra de internal/layout): minúsculo, sem separadores.
export const RE_INICIATIVA = /^[0-9a-z]+(-[a-z0-9]+)+$/

export function faseDaIniciativa(iniciativa: string | undefined): string | undefined {
  if (iniciativa === undefined) return undefined
  const m = /^([0-9a-z]+)-/.exec(iniciativa)
  return m ? m[1] : undefined
}

export function featureDoRegistro(iniciativa: string | undefined, sufixo: string): string {
  return iniciativa === undefined ? sufixo : `sdd/discovery/${iniciativa}/${sufixo}`
}
