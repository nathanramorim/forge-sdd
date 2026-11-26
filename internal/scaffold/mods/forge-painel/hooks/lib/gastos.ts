// Resumo do `/gastos`: lê os registros de `sdd/.metrics` e soma por estação (papel).
// O dólar sai sempre rotulado "custo de tabela" (RN-4): estimativa do Claude Code, não fatura.

export type Resumo = {
  registros: number
  porEstacao: Record<string, { sessoes: number; tokens: number; custo: number }>
  reais: number
  estimados: number
}

type RegistroCru = {
  agent_path?: unknown
  tokens_input?: unknown
  tokens_output?: unknown
  cost_usd_tabela?: unknown
  medicao?: unknown
  feature?: unknown
  phase?: unknown
}

const num = (v: unknown): number => (typeof v === 'number' && Number.isFinite(v) ? v : 0)

// Estação do registro: último item de agent_path (orquestrador, builder, revisor).
function estacao(r: RegistroCru): string {
  if (Array.isArray(r.agent_path) && r.agent_path.length > 0) {
    const ultimo = r.agent_path[r.agent_path.length - 1]
    if (typeof ultimo === 'string') return ultimo.slice(0, 30)
  }
  return 'desconhecida'
}

// Os registros do archive (`medicao: resumo`) não entram nas somas, como no schema.
export function resumir(brutos: string[], iniciativa: string | undefined): Resumo {
  const r: Resumo = { registros: 0, porEstacao: {}, reais: 0, estimados: 0 }
  for (const bruto of brutos) {
    let reg: RegistroCru
    try {
      reg = JSON.parse(bruto) as RegistroCru
    } catch {
      continue
    }
    if (reg.medicao === 'resumo') continue
    if (iniciativa !== undefined) {
      const fase = iniciativa.split('-')[0]
      const nasIni =
        (typeof reg.feature === 'string' && reg.feature.includes(`/${iniciativa}/`)) ||
        (typeof reg.phase === 'string' && (reg.phase === fase || reg.phase.includes(iniciativa)))
      if (!nasIni) continue
    }
    const est = estacao(reg)
    const acc = (r.porEstacao[est] ??= { sessoes: 0, tokens: 0, custo: 0 })
    acc.sessoes += 1
    acc.tokens += num(reg.tokens_input) + num(reg.tokens_output)
    acc.custo += num(reg.cost_usd_tabela)
    r.registros += 1
    if (reg.medicao === 'real') r.reais += 1
    else r.estimados += 1
  }
  return r
}

function milhar(n: number): string {
  return String(Math.round(n)).replace(/\B(?=(\d{3})+(?!\d))/g, '.')
}

export function textoDoResumo(r: Resumo, iniciativa: string | undefined): string {
  const alvo = iniciativa === undefined ? 'projeto' : `iniciativa ${iniciativa}`
  if (r.registros === 0) return `Sem registros em sdd/.metrics para ${alvo}.`
  const linhas = [`Gastos de ${alvo} (${r.registros} registros: ${r.reais} reais, ${r.estimados} estimados)`]
  let tokens = 0
  let custo = 0
  for (const est of Object.keys(r.porEstacao).sort()) {
    const a = r.porEstacao[est] ?? { sessoes: 0, tokens: 0, custo: 0 }
    tokens += a.tokens
    custo += a.custo
    linhas.push(`- ${est}: ${a.sessoes} sessões, ${milhar(a.tokens)} tokens, custo de tabela US$ ${a.custo.toFixed(2)}`)
  }
  linhas.push(`Total: ${milhar(tokens)} tokens, custo de tabela US$ ${custo.toFixed(2)}`)
  linhas.push('Custo de tabela é a estimativa do Claude Code a preço de lista; não é fatura.')
  return linhas.join('\n')
}
