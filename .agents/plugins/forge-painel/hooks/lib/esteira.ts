// Leitura (somente leitura, RN-10) do barramento `.forge/runtime/queues/<papel>/<estado>/` e dos
// logs `.forge/runtime/logs/<papel>-<id>.log`. O Go é o escritor único; aqui nada é gravado.

export const PAPEIS = ['builder', 'reviewer'] as const
export const ESTADOS = ['pending', 'processing', 'done', 'failed'] as const
export const DIR_FILAS = '.forge/runtime/queues'
export const DIR_LOGS = '.forge/runtime/logs'

export const RE_ID_JOB = /^[a-z0-9][a-z0-9._-]{0,80}$/

export type Linha = {
  papel: string
  estado: string
  id: string
  tarefa: string
  tentativa: number
  parado: boolean
  desde?: string
  tokens?: number
}

// SEG-08: texto de job e de log passa por aqui antes de ir à tela: sem caracteres de controle
// (inclui escapes ANSI) e com tamanho limitado.
export function sanear(texto: unknown, max = 80): string {
  if (typeof texto !== 'string') return ''
  // eslint-disable-next-line no-control-regex
  const limpo = texto.replace(/\u001b\[[0-9;?]*[ -/]*[@-~]/g, '').replace(/[\u0000-\u001f\u007f-\u009f]/g, ' ')
  const junto = limpo.replace(/\s+/g, ' ').trim()
  return junto.length > max ? `${junto.slice(0, max - 1)}…` : junto
}

type JobCru = {
  id?: unknown
  feature?: unknown
  task_ids?: unknown
  attempt?: unknown
  lease?: { expires_at?: unknown } | null
  updated_at?: unknown
}

// Converte o JSON de um job em linha do painel. Job ilegível ou sem id válido é ignorado.
export function linhaDoJob(papel: string, estado: string, bruto: string, agora: number): Linha | undefined {
  let job: JobCru
  try {
    job = JSON.parse(bruto) as JobCru
  } catch {
    return undefined
  }
  if (typeof job.id !== 'string' || !RE_ID_JOB.test(job.id)) return undefined
  const tasks = Array.isArray(job.task_ids)
    ? job.task_ids.filter((t): t is string => typeof t === 'string').slice(0, 3).map(t => sanear(t, 20))
    : []
  const tarefa = tasks.length > 0 ? tasks.join(',') : sanear(job.feature, 40)
  let parado = false
  const expira = job.lease?.expires_at
  if (estado === 'processing' && typeof expira === 'string') {
    const t = Date.parse(expira)
    parado = !Number.isNaN(t) && t < agora
  }
  return {
    papel,
    estado,
    id: job.id,
    tarefa,
    tentativa: typeof job.attempt === 'number' ? job.attempt : 0,
    parado,
    desde: typeof job.updated_at === 'string' ? job.updated_at : undefined,
  }
}

// Tokens de um log stream-json: o evento `result` final quando existe; em execução, o último
// uso que uma mensagem do assistente trouxe (entrada mais cache) somado às saídas, ao vivo.
export function tokensDoLog(texto: string): number | undefined {
  const linhas = texto.split('\n')
  const n = (u: Record<string, unknown> | undefined, k: string) => (typeof u?.[k] === 'number' ? (u[k] as number) : 0)
  for (let i = linhas.length - 1; i >= 0; i--) {
    const l = linhas[i] ?? ''
    if (!l.includes('"result"')) continue
    try {
      const ev = JSON.parse(l) as { type?: string; usage?: Record<string, unknown> }
      if (ev.type !== 'result' || ev.usage === undefined) continue
      return n(ev.usage, 'input_tokens') + n(ev.usage, 'output_tokens') + n(ev.usage, 'cache_read_input_tokens') + n(ev.usage, 'cache_creation_input_tokens')
    } catch {
      continue
    }
  }
  let saida = 0
  let entrada = 0
  let viu = false
  for (const l of linhas) {
    if (!l.includes('"usage"')) continue
    try {
      const ev = JSON.parse(l) as { type?: string; message?: { usage?: Record<string, unknown> } }
      const u = ev.type === 'assistant' ? ev.message?.usage : undefined
      if (u === undefined) continue
      viu = true
      saida += n(u, 'output_tokens')
      entrada = n(u, 'input_tokens') + n(u, 'cache_read_input_tokens') + n(u, 'cache_creation_input_tokens')
    } catch {
      continue
    }
  }
  return viu ? entrada + saida : undefined
}

// Últimas linhas de um log, saneadas, para o `/painel log`.
export function caudaDoLog(texto: string, linhas = 30): string {
  return texto
    .split('\n')
    .map(l => sanear(l, 200))
    .filter(l => l !== '')
    .slice(-linhas)
    .join('\n')
}

export function rotuloEstado(l: Linha): string {
  if (l.parado) return 'parado (lease vencido)'
  switch (l.estado) {
    case 'pending':
      return 'na fila'
    case 'processing':
      return 'em execução'
    case 'done':
      return 'concluído'
    case 'failed':
      return 'falhou'
    default:
      return sanear(l.estado, 20)
  }
}

export function tempoDecorrido(desde: string | undefined, agora: number): string {
  if (desde === undefined) return '-'
  const t = Date.parse(desde)
  if (Number.isNaN(t)) return '-'
  const s = Math.max(0, Math.round((agora - t) / 1000))
  if (s < 60) return `${s}s`
  if (s < 3600) return `${Math.floor(s / 60)}min`
  return `${Math.floor(s / 3600)}h${String(Math.floor((s % 3600) / 60)).padStart(2, '0')}`
}
