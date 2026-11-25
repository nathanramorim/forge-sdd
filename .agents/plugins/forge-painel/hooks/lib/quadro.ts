// Quadro da esteira por iniciativa (funções puras): lê o JSON dos jobs do barramento e monta o
// que o painel e a status line mostram. Nada aqui toca disco; o Go é o escritor único (RN-10).

import { RE_ID_JOB, sanear } from './esteira'

export type JobLido = {
  papel: string
  estado: string
  id: string
  tipo: string
  iniciativa: string
  taskId: string
  taskIds: string[]
  criado?: number
  atualizado?: number
  tokens?: number
  parado: boolean
  inicio?: number // quando o worker pegou o job (último evento processing)
  fim?: number // último evento done ou failed
  motivo: string // nota do último evento (motivo da falha)
}

export type LinhaQuadro = {
  task: string
  builder: string
  reviewer: string
  tokens: number
  tempo: string
  ativo: boolean
}

export type Cartao = {
  nome: string
  estado: string
  classe: 'ok' | 'run' | 'wait' | 'fail'
  tokens: number
  tempo: string
  jobId: string
}

export type Quadro = {
  iniciativa: string
  linhas: LinhaQuadro[]
  cartoes: Cartao[]
  modoFeature: boolean
  ativos: number
  mescladas: number
  total: number
  tokens: number
}

const RE_SLUG = /^[0-9a-z]+(-[a-z0-9]+)+$/

function data(v: unknown): number | undefined {
  if (typeof v !== 'string') return undefined
  const t = Date.parse(v)
  return Number.isNaN(t) ? undefined : t
}

export function lerJob(papel: string, estado: string, bruto: string, agora: number): JobLido | undefined {
  let j: any
  try {
    j = JSON.parse(bruto)
  } catch {
    return undefined
  }
  if (typeof j?.id !== 'string' || !RE_ID_JOB.test(j.id)) return undefined
  const ini = typeof j.initiative === 'string' && RE_SLUG.test(j.initiative) ? j.initiative : ''
  const taskIds = Array.isArray(j.task_ids) ? j.task_ids.filter((t: unknown): t is string => typeof t === 'string').map((t: string) => sanear(t, 30)) : []
  const expira = data(j.lease?.expires_at)
  const evs: any[] = Array.isArray(j.events) ? j.events : []
  let inicio: number | undefined
  let fim: number | undefined
  for (const ev of evs) {
    const t = data(ev?.at)
    if (t === undefined) continue
    if (ev.state === 'processing') inicio = t
    if (ev.state === 'done' || ev.state === 'failed') fim = t
  }
  const ultimo = evs[evs.length - 1]
  return {
    inicio,
    fim,
    motivo: typeof ultimo?.note === 'string' ? sanear(ultimo.note, 120) : '',
    papel,
    estado,
    id: j.id,
    tipo: sanear(j.type, 20),
    iniciativa: ini,
    taskId: typeof j.task_id === 'string' ? sanear(j.task_id, 30) : '',
    taskIds,
    criado: data(j.created_at),
    atualizado: data(j.updated_at),
    parado: estado === 'processing' && expira !== undefined && expira < agora,
  }
}

const ativo = (j: JobLido) => j.estado === 'pending' || j.estado === 'processing'
const quando = (j: JobLido) => j.atualizado ?? j.criado ?? 0

// A iniciativa atual: a do job ativo mais recente; senão a do job mais recente das últimas 24h;
// senão a pasta de discovery com status.md mais recente (porStatus, já ordenada).
export function escolherIniciativa(jobs: JobLido[], porStatus: string | undefined, agora: number): string | undefined {
  const com = jobs.filter(j => j.iniciativa !== '')
  const ativos = com.filter(ativo).sort((a, b) => quando(b) - quando(a))
  if (ativos.length > 0) return ativos[0]!.iniciativa
  const recentes = com.filter(j => agora - quando(j) < 24 * 3600_000).sort((a, b) => quando(b) - quando(a))
  if (recentes.length > 0) return recentes[0]!.iniciativa
  return porStatus
}

// Estação em curso na iniciativa: o papel do job em execução; sem job, a sessão é a Discovery.
export function estacaoAtual(jobs: JobLido[], ini: string | undefined): string {
  if (ini === undefined) return 'discovery'
  const rodando = jobs.find(j => j.iniciativa === ini && j.estado === 'processing')
  return rodando ? rodando.papel : 'discovery'
}

function rotulo(j: JobLido | undefined, papel: string): string {
  if (j === undefined) return papel === 'reviewer' ? '·' : 'na fila'
  if (j.parado) return 'parado'
  switch (j.estado) {
    case 'pending':
      return 'na fila'
    case 'processing':
      return papel === 'reviewer' ? 'revisando' : 'rodando'
    case 'done':
      return papel === 'reviewer' ? 'aprovado' : 'commitado'
    case 'failed':
      return 'falhou'
    default:
      return j.estado
  }
}

export function duracao(ms: number): string {
  const s = Math.max(0, Math.round(ms / 1000))
  if (s < 60) return `${s}s`
  if (s < 3600) return `${Math.floor(s / 60)}m ${String(s % 60).padStart(2, '0')}s`
  return `${Math.floor(s / 3600)}h${String(Math.floor((s % 3600) / 60)).padStart(2, '0')}`
}

export function tokensCurtos(n: number): string {
  if (n < 1000) return String(n)
  if (n < 1_000_000) return `${(n / 1000).toFixed(1).replace('.', ',')}k`
  return `${(n / 1_000_000).toFixed(1).replace('.', ',')}M`
}

// Uma linha por task (par Builder e Reviewer), na ordem do plano; jobs por feature (modo antigo)
// viram uma linha só. estadosTask vem de .forge/runtime/tarefas (mesclada, bloqueada).
export function montarQuadro(jobs: JobLido[], ini: string, estadosTask: Record<string, string>, agora: number): Quadro {
  const daIni = jobs.filter(j => j.iniciativa === ini)
  const grupos = new Map<string, JobLido[]>()
  let modoFeature = false
  for (const j of daIni) {
    let chave = j.taskId
    if (chave === '') {
      modoFeature = modoFeature || (j.papel === 'builder' && j.taskIds.length > 1)
      chave = j.taskIds.length > 1 ? `feature (${j.taskIds.length} tasks)` : j.taskIds[0] ?? j.id
    }
    const g = grupos.get(chave) ?? []
    g.push(j)
    grupos.set(chave, g)
  }
  const linhas: LinhaQuadro[] = []
  const cartoes: Cartao[] = []
  let ativos = 0
  let tokens = 0
  for (const [task, g] of grupos) {
    const ultimo = (papel: string) => g.filter(j => j.papel === papel).sort((a, b) => quando(b) - quando(a))[0]
    const b = ultimo('builder')
    const r = ultimo('reviewer')
    const t = g.reduce((s, j) => s + (j.tokens ?? 0), 0)
    const inicio = Math.min(...g.map(j => j.criado ?? agora))
    const algumAtivo = g.some(ativo)
    const fim = algumAtivo ? agora : Math.max(...g.map(quando))
    let reviewer = rotulo(r, 'reviewer')
    const est = estadosTask[`feat-${task}`]
    if (est === 'mesclada') reviewer = 'mesclada'
    else if (est === 'bloqueada') reviewer = 'bloqueada'
    linhas.push({ task, builder: rotulo(b, 'builder'), reviewer, tokens: t, tempo: duracao(fim - inicio), ativo: algumAtivo })
    for (const [papel, j] of [['builder', b], ['reviewer', r]] as const) {
      if (j === undefined && !(papel === 'reviewer' && b?.estado === 'done')) continue
      const est = papel === 'reviewer' ? reviewer : rotulo(j, papel)
      const ini0 = j?.criado ?? agora
      const fim0 = j === undefined ? agora : ativo(j) ? agora : quando(j)
      cartoes.push({
        nome: `${papel} · ${task}`,
        estado: est,
        classe: classeDe(est),
        tokens: j?.tokens ?? 0,
        tempo: j === undefined || j.estado === 'pending' ? '·' : duracao(fim0 - ini0),
        jobId: j?.id ?? '',
      })
    }
    if (algumAtivo) ativos++
    tokens += t
  }
  linhas.sort((a, b) => a.task.localeCompare(b.task))
  // Em execução primeiro, depois na fila, depois os encerrados (mais recentes no fim da tela).
  const ordem = { run: 0, wait: 1, fail: 2, ok: 3 }
  cartoes.sort((a, b) => ordem[a.classe] - ordem[b.classe] || a.nome.localeCompare(b.nome))
  const mescladas = Object.values(estadosTask).filter(e => e === 'mesclada').length
  const total = Math.max(linhas.length, Object.keys(estadosTask).length)
  return { iniciativa: ini, linhas, cartoes, modoFeature, ativos, mescladas, total, tokens }
}

// Faixa acima do prompt: não lidas do inbox e PRs conformes entre elas.
export function resumoInbox(brutos: string[]): { naoLidas: number; prConforme: string[] } {
  let naoLidas = 0
  const prConforme: string[] = []
  for (const b of brutos) {
    try {
      const n = JSON.parse(b)
      if (n?.read === true) continue
      naoLidas++
      if (n?.type === 'pr_approved') {
        const m = typeof n.pr_url === 'string' ? /\/pull\/(\d+)/.exec(n.pr_url) : null
        prConforme.push(m ? `PR #${m[1]}` : sanear(n.feature, 30) || 'PR')
      }
    } catch {
      continue
    }
  }
  return { naoLidas, prConforme }
}

export function classeDe(estado: string): 'ok' | 'run' | 'wait' | 'fail' {
  if (estado === 'rodando' || estado === 'revisando') return 'run'
  if (estado === 'falhou' || estado === 'parado' || estado === 'bloqueada') return 'fail'
  if (estado === 'na fila' || estado === '·') return 'wait'
  return 'ok'
}

const BLOCOS = '▁▂▃▄▅▆▇█'

// Tendência ao vivo dos tokens de um agente em execução (as amostras vêm de cada releitura).
export function sparkline(amostras: number[], largura = 24): string {
  const v = amostras.slice(-largura)
  if (v.length < 2) return ''
  const min = Math.min(...v)
  const max = Math.max(...v)
  return v.map(x => BLOCOS[max === min ? 3 : Math.round(((x - min) / (max - min)) * 7)]).join('')
}

// Barra curta de percentual para a status line.
export function barra(pct: number, largura = 8): string {
  const cheio = Math.max(0, Math.min(largura, Math.round((pct / 100) * largura)))
  return '▰'.repeat(cheio) + '▱'.repeat(largura - cheio)
}

// ---- Painel como no protótipo: um cartão por estação (Builder e Reviewer) da iniciativa atual ----

export type Estacao = {
  papel: string
  estado: string
  classe: 'ok' | 'run' | 'wait' | 'fail'
  tokens: number
  tempo: string
  detalhe: string
  jobId: string
}

// O job mais recente de cada papel na iniciativa vira o cartão daquela estação. O tempo conta de
// quando o worker pegou o job (não de quando foi criado) e o detalhe diz o motivo da falha ou o que
// o agente está fazendo agora (atividade, da última linha útil do log do worker).
export function montarEstacoes(jobs: JobLido[], ini: string, agora: number, atividade: Record<string, string> = {}): Estacao[] {
  const out: Estacao[] = []
  for (const papel of ['builder', 'reviewer']) {
    const j = jobs.filter(x => x.iniciativa === ini && x.papel === papel).sort((a, b) => Number(ativo(b)) - Number(ativo(a)) || quando(b) - quando(a))[0]
    if (j === undefined) {
      out.push({ papel, estado: 'ocioso', classe: 'wait', tokens: 0, tempo: '·', detalhe: '', jobId: '' })
      continue
    }
    const estado = rotulo(j, papel) === 'commitado' ? 'concluído' : rotulo(j, papel)
    const classe = classeDe(estado)
    let tempo = '·'
    if (j.estado !== 'pending' && j.inicio !== undefined) {
      const fim = j.estado === 'processing' ? agora : j.fim ?? quando(j)
      tempo = duracao(fim - j.inicio)
    }
    let detalhe = ''
    if (classe === 'fail') detalhe = j.parado ? 'worker parou no meio (rode forge up)' : j.motivo
    else if (classe === 'run') detalhe = atividade[papel] ?? ''
    out.push({ papel, estado, classe, tokens: j.tokens ?? 0, tempo, detalhe, jobId: j.id })
  }
  return out
}

// Progresso da feature pelo queue.json da iniciativa: tasks prontas, total e a em andamento.
export function progressoDaFila(bruto: string | undefined): { feitas: number; total: number; atual: string } | null {
  if (bruto === undefined) return null
  let q: any
  try {
    q = JSON.parse(bruto)
  } catch {
    return null
  }
  const itens: any[] = Array.isArray(q?.items) ? q.items : []
  if (itens.length === 0) return null
  const prontas = new Set(['approved', 'success', 'review_pending', 'done'])
  let feitas = 0
  let atual = ''
  for (const it of itens) {
    const st = typeof it?.status === 'string' ? it.status : ''
    if (prontas.has(st)) feitas++
    else if (atual === '' && (st === 'building' || st === 'running')) atual = sanear(it.task_id, 30)
  }
  return { feitas, total: itens.length, atual }
}

// Última linha útil do log limpo do worker ("[builder] 🔨 Editando ..."), sem o prefixo do papel.
export function atividadeDoWorker(texto: string | undefined, papel: string): string {
  if (texto === undefined) return ''
  const linhas = texto.split('\n')
  for (let i = linhas.length - 1; i >= 0; i--) {
    const l = linhas[i] ?? ''
    if (l.startsWith('[forge-')) return ''
    if (!l.startsWith(`[${papel}]`)) continue
    return sanear(l.slice(papel.length + 2).trim().replace(/^[^\p{L}\d]+/u, ''), 70)
  }
  return ''
}
