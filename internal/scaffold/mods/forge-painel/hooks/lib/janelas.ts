// Funções puras da status line e dos alertas de janela. Nada aqui toca disco nem a API:
// recebe os números que o Claude Code já mediu (`session.measure`) e devolve texto.

export type Janela = { kind: string; percentUsed: number; resetsAt?: string }

export type Medida = {
  context: { percent?: number }
  rateLimits: Janela[]
  cost?: { usd: number }
}

// Limites dos alertas (decididos pelo dono em 24/11/2026, status.md da 123).
export const NIVEIS_ALERTA = [70, 90] as const

const ROTULO: Record<string, string> = { five_hour: '5h', seven_day: '7d' }

function dois(n: number): string {
  return n < 10 ? `0${n}` : String(n)
}

// Hora local de renovação: HH:MM, e dd/mm HH:MM quando não é hoje.
export function horaDeRenovacao(iso: string | undefined, agora: number): string {
  if (iso === undefined) return ''
  const t = new Date(iso)
  if (Number.isNaN(t.getTime())) return ''
  const hora = `${dois(t.getHours())}:${dois(t.getMinutes())}`
  const hoje = new Date(agora)
  const mesmoDia =
    t.getFullYear() === hoje.getFullYear() &&
    t.getMonth() === hoje.getMonth() &&
    t.getDate() === hoje.getDate()
  return mesmoDia ? hora : `${dois(t.getDate())}/${dois(t.getMonth() + 1)} ${hora}`
}

export function rotuloJanela(kind: string): string {
  return ROTULO[kind] ?? kind
}

function pct(n: number): string {
  return `${Math.round(n)}%`
}

// Só as janelas de assinatura entram na linha (o limite de gasto de gateway não é janela de tempo).
export function janelasDeTempo(rateLimits: Janela[]): Janela[] {
  return rateLimits.filter(j => j.kind in ROTULO)
}

// Linha de status: iniciativa, janelas de 5h e 7d e contexto. Fora de assinatura (rateLimits
// vazio) mostra só contexto e o custo de tabela (RN-4: nunca fatura).
export function linhaDeStatus(m: Medida, iniciativa: string | undefined, estacao?: string, telemetria?: boolean): string {
  const cabeca = ['forge']
  if (iniciativa) cabeca.push(iniciativa.length > 28 ? `${iniciativa.slice(0, 27)}…` : iniciativa)
  if (estacao) cabeca.push(estacao)
  const partes: string[] = [cabeca.join(' · ')]
  const janelas = janelasDeTempo(m.rateLimits)
  for (const j of janelas) {
    partes.push(`${rotuloJanela(j.kind)} ${pct(j.percentUsed)}`)
  }
  partes.push(m.context.percent === undefined ? 'ctx -' : `ctx ${pct(m.context.percent)}`)
  if (janelas.length === 0 && m.cost !== undefined) {
    partes.push(`custo de tabela US$ ${m.cost.usd.toFixed(2)}`)
  }
  if (telemetria !== undefined) partes.push(telemetria ? '● telemetria real' : '○ telemetria desligada')
  return partes.join(' │ ')
}

// Estado dos alertas por janela: qual nível já avisou, para a janela que vale agora (resetsAt).
export type EstadoAlertas = Record<string, { janela: string; nivel: number }>

export type Alerta = { kind: string; nivel: number; texto: string }

// Decide quais toasts disparar. Um por nível cruzado, uma vez por janela: quem já passou de 70%
// não avisa 70 de novo; se a leitura pula direto a 90% avisa só o 90. Janela nova (resetsAt
// diferente) zera o nível. Devolve os alertas e o estado novo (sem mutar o recebido).
export function avaliarAlertas(
  rateLimits: Janela[],
  anterior: EstadoAlertas,
  agora: number,
): { alertas: Alerta[]; estado: EstadoAlertas } {
  const estado: EstadoAlertas = { ...anterior }
  const alertas: Alerta[] = []
  for (const j of janelasDeTempo(rateLimits)) {
    const id = j.resetsAt ?? 'sem-renovacao'
    const visto = estado[j.kind]
    let nivelVisto = visto !== undefined && visto.janela === id ? visto.nivel : 0
    let cruzado = 0
    for (const nivel of NIVEIS_ALERTA) {
      if (j.percentUsed >= nivel && nivel > nivelVisto) cruzado = nivel
    }
    if (cruzado > 0) {
      const quando = horaDeRenovacao(j.resetsAt, agora)
      const renova = quando === '' ? '' : `; renova ${quando}`
      alertas.push({
        kind: j.kind,
        nivel: cruzado,
        texto: `Janela de ${rotuloJanela(j.kind)} em ${pct(j.percentUsed)} (passou de ${cruzado}%${renova})`,
      })
      nivelVisto = cruzado
    }
    estado[j.kind] = { janela: id, nivel: nivelVisto }
  }
  return { alertas, estado }
}

// Janela mais cheia entre as de tempo, para o aviso de orçamento em agent.spawn.
export function janelaMaisCheia(rateLimits: Janela[]): Janela | undefined {
  let pior: Janela | undefined
  for (const j of janelasDeTempo(rateLimits)) {
    if (pior === undefined || j.percentUsed > pior.percentUsed) pior = j
  }
  return pior
}

// Limite do aviso de orçamento: o mesmo em que o despachante do forge cai para 1 par (ADR-0029).
export const LIMITE_ORCAMENTO = 85

// Conteúdo de .forge/runtime/painel/janela.json: o forge lê esse arquivo para o teto de paralelo
// (LerJanelaDoPainel, válido por 15 minutos). Sem janelas de tempo, não há o que gravar.
export function conteudoJanela(rateLimits: Janela[], agora: number): string | undefined {
  let cinco: number | undefined
  let sete: number | undefined
  for (const j of janelasDeTempo(rateLimits)) {
    const p = Math.max(0, Math.min(100, Math.round(j.percentUsed * 10) / 10))
    if (j.kind === 'five_hour') cinco = p
    if (j.kind === 'seven_day') sete = p
  }
  if (cinco === undefined && sete === undefined) return undefined
  const o: Record<string, unknown> = {}
  if (cinco !== undefined) o.five_hour_pct = cinco
  if (sete !== undefined) o.seven_day_pct = sete
  o.at = new Date(agora).toISOString().replace(/\.\d{3}Z$/, 'Z')
  return `${JSON.stringify(o)}\n`
}
