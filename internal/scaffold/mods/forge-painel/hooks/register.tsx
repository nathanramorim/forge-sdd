import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import { PAPEIS, ESTADOS, DIR_FILAS, DIR_LOGS, RE_ID_JOB, tokensDoLog, caudaDoLog, sanear } from './lib/esteira'
import { duracao, escolherIniciativa, estacaoAtual, lerJob, montarQuadro, resumoInbox, tokensCurtos } from './lib/quadro'
import type { JobLido } from './lib/quadro'
import { resumir, textoDoResumo } from './lib/gastos'
import { avaliarAlertas, horaDeRenovacao, conteudoJanela, janelaMaisCheia, linhaDeStatus, LIMITE_ORCAMENTO, rotuloJanela } from './lib/janelas'
import type { EstadoAlertas } from './lib/janelas'
import { DIR_METRICAS, RE_INICIATIVA, faseDaIniciativa, featureDoRegistro, montarRegistro, nomeBase, papelDoSubagente } from './lib/telemetria'
import type { Uso } from './lib/telemetria'
import type { Esteira } from '../types'

const PAINEL = 'forge-painel'
const DIR_PAINEL = '.forge/runtime/painel'
const ARQ_ALERTAS = `${DIR_PAINEL}/alertas.json`
const ARQ_JANELA = `${DIR_PAINEL}/janela.json`
const JANELA_A_CADA_MS = 60_000
const DIR_INICIATIVAS = 'sdd/discovery'
const MAX_REGISTROS_GASTOS = 500

// Cache do painel: não é fonte de verdade, só o que a tela mostra (o barramento é do Go).
const VAZIA: Esteira = { quadro: null, outras: [], agentes: [], erro: '', inbox: { naoLidas: 0, prConforme: [] }, temFilas: false }
const esteira = atom({ plugin: 'forge-painel', key: 'esteira' } as const, VAZIA)
// tool_use_id da chamada Agent -> id do subagente, para o cartão de consumo embaixo da chamada.
const cartoes = atom({ plugin: 'forge-painel', key: 'cartoes' } as const, {} as Record<string, string>)
const DIR_INBOX = '.forge/runtime/queues/discovery/inbox'
const DIR_TAREFAS = '.forge/runtime/tarefas'

// Estado em memória do módulo: some no reload, e tudo bem (sem estado próprio além de cache).
let mudoCache: boolean | undefined
let interativa = false
let iniciativa: string | undefined
let estacao = 'discovery'
let telemetriaOn: boolean | undefined
let ultimaMedida: any
let alertas: EstadoAlertas = {}
let janelaGravadaEm = 0
let inicioSessao = 0
let turnos = 0
const usoSessao: Uso[] = []
let modeloSessao = ''
const agentesGravados = new Set<string>()


// RN-11: dentro de um worker (FORGE_WORKER presente) o mod não grava nem abre painel.
async function mudo($: EngineInterface): Promise<boolean> {
  if (mudoCache === undefined) {
    const v = await $.env.get('FORGE_WORKER')
    mudoCache = v !== undefined && v !== ''
  }
  return mudoCache
}

// RN-2: a escrita só vale dentro de sdd/.metrics e .forge/runtime/painel.
async function gravar($: EngineInterface, caminho: string, texto: string): Promise<void> {
  const ok = (caminho.startsWith(`${DIR_METRICAS}/`) || caminho.startsWith(`${DIR_PAINEL}/`)) && !caminho.includes('..')
  if (!ok) throw new Error(`escrita fora do escopo do mod: ${caminho}`)
  await $.fs.write(caminho, texto)
}

async function lerJson($: EngineInterface, caminho: string): Promise<any | undefined> {
  try {
    return JSON.parse(await $.fs.read(caminho))
  } catch {
    return undefined
  }
}

async function iniciativaPorStatus($: EngineInterface): Promise<string | undefined> {
  try {
    const entradas = await $.fs.list(DIR_INICIATIVAS)
    let melhor: string | undefined
    let melhorT = -1
    for (const en of entradas) {
      if (en.kind !== 'dir' || !RE_INICIATIVA.test(en.name)) continue
      const st = await $.fs.stat(`${DIR_INICIATIVAS}/${en.name}/status.md`).catch(() => undefined)
      if (st !== undefined && st.mtimeMs > melhorT) {
        melhorT = st.mtimeMs
        melhor = en.name
      }
    }
    return melhor
  } catch {
    return undefined
  }
}

async function descobrirIniciativa($: EngineInterface): Promise<void> {
  await atualizar($).catch(() => undefined)
}

async function telemetriaLigada($: EngineInterface): Promise<boolean> {
  const rc = await lerJson($, 'sdd/.sddrc')
  return rc?.telemetry?.enabled !== false
}

async function versaoSdd($: EngineInterface): Promise<string> {
  const rc = await lerJson($, 'sdd/.sddrc')
  return typeof rc?.version === 'string' ? rc.version : 'desconhecida'
}

async function gravarRegistro($: EngineInterface, registro: unknown): Promise<string> {
  const base = nomeBase(await $.clock.now())
  let nome = `${base}.json`
  for (let i = 2; await $.fs.exists(`${DIR_METRICAS}/${nome}`); i++) nome = `${base}-${i}.json`
  await gravar($, `${DIR_METRICAS}/${nome}`, `${JSON.stringify(registro, null, 2)}\n`)
  return nome
}

async function lerTexto($: EngineInterface, caminho: string): Promise<string | undefined> {
  try {
    return await $.fs.read(caminho)
  } catch {
    return undefined
  }
}

async function listar($: EngineInterface, dir: string): Promise<string[]> {
  try {
    const ents = (await $.fs.list(dir)) as any[]
    return ents.filter(en => en.kind === 'file' && en.name.endsWith('.json') && !en.name.startsWith('.')).map(en => en.name as string).sort()
  } catch {
    return []
  }
}

// Relê o barramento e monta o quadro da iniciativa atual (RN-10: só leitura).
async function atualizar($: EngineInterface): Promise<void> {
  const agora = await $.clock.now()
  const jobs: JobLido[] = []
  let temFilas = false
  for (const papel of PAPEIS) {
    for (const estado of ESTADOS) {
      const dir = `${DIR_FILAS}/${papel}/${estado}`
      const nomes = await listar($, dir)
      if (nomes.length > 0) temFilas = true
      // Concluídos e falhos: os mais recentes bastam para o quadro (o histórico fica em /gastos).
      const recorte = estado === 'done' || estado === 'failed' ? nomes.slice(-60) : nomes.slice(0, 40)
      for (const nome of recorte) {
        const bruto = await lerTexto($, `${dir}/${nome}`)
        const j = bruto === undefined ? undefined : lerJob(papel, estado, bruto, agora)
        if (j !== undefined) jobs.push(j)
      }
    }
  }
  iniciativa = escolherIniciativa(jobs, await iniciativaPorStatus($), agora)
  estacao = estacaoAtual(jobs, iniciativa)
  let quadro: Esteira['quadro'] = null
  if (iniciativa !== undefined) {
    for (const j of jobs.filter(x => x.iniciativa === iniciativa)) {
      const log = await $.fs.read(`${DIR_LOGS}/${j.papel}-${j.id}.log`, { as: 'text', offset: 0, length: 262144 }).catch(() => undefined)
      if (typeof log === 'string') j.tokens = tokensDoLog(log)
    }
    const estadosTask: Record<string, string> = {}
    for (const nome of await listar($, `${DIR_TAREFAS}/${iniciativa}`)) {
      if (!nome.startsWith('feat-')) continue
      const st = JSON.parse((await lerTexto($, `${DIR_TAREFAS}/${iniciativa}/${nome}`)) ?? '{}')
      if (typeof st?.task === 'string' && typeof st?.estado === 'string') estadosTask[sanear(st.task, 40)] = sanear(st.estado, 20)
    }
    quadro = montarQuadro(jobs, iniciativa, estadosTask, agora)
  }
  const outras = [...new Set(jobs.filter(j => j.iniciativa !== '' && j.iniciativa !== iniciativa && (j.estado === 'pending' || j.estado === 'processing')).map(j => j.iniciativa))]
  const brutosInbox: string[] = []
  for (const nome of (await listar($, DIR_INBOX)).slice(-50)) {
    const b = await lerTexto($, `${DIR_INBOX}/${nome}`)
    if (b !== undefined) brutosInbox.push(b)
  }
  let agentes: Esteira['agentes'] = []
  let erro = ''
  try {
    const lista = await $.agent.list()
    agentes = lista.slice(-10).map((a: any) => ({
      tipo: sanear(a.type, 30),
      estado: sanear(a.status, 20),
      tokens: typeof a.tokenCount === 'number' ? a.tokenCount : 0,
    }))
  } catch {
    erro = 'lista de agentes indisponível'
  }
  telemetriaOn = await telemetriaLigada($)
  if (ultimaMedida !== undefined) $.ui.status(linhaDeStatus(ultimaMedida, iniciativa, estacao, telemetriaOn))
  try {
    await update($, esteira, () => ({ quadro, outras, agentes, erro, inbox: resumoInbox(brutosInbox), temFilas }))
  } catch {
    // a sessão pode ter terminado entre a leitura e a gravação do cache
  }
}

async function abrirPainel($: EngineInterface): Promise<void> {
  await atualizar($)
  void $.ui.open({ id: PAINEL, title: 'Forge: esteira' }).catch(() => undefined)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    if (await mudo($)) return next(e)
    interativa = e.isInteractive
    inicioSessao = await $.clock.now()
    await $.command.register({ name: 'gastos', description: 'Gastos da iniciativa por estação (custo de tabela)' })
    await $.command.register({ name: 'painel', description: 'Abre o painel da esteira; "/painel log <papel> <job>" mostra o fim de um log' })
    if (!interativa) return next(e)
    await descobrirIniciativa($)
    const salvo = await lerJson($, ARQ_ALERTAS)
    if (salvo !== undefined && typeof salvo === 'object') alertas = salvo as EstadoAlertas
    void $.clock.every(5000, () => atualizar($).catch(() => undefined))
    // Se o `forge up` já criou as filas, o painel abre sozinho; senão, `/painel`.
    if (await $.fs.exists(DIR_FILAS)) await abrirPainel($)
    return next(e)
  })

  on('session.measure', async ($, e, next) => {
    if (await mudo($)) return next(e)
    const agora = await $.clock.now()
    ultimaMedida = e
    $.ui.status(linhaDeStatus(e, iniciativa, estacao, telemetriaOn))
    // A janela vai para o forge decidir o teto de paralelo (no máximo uma escrita por minuto).
    if (agora - janelaGravadaEm >= JANELA_A_CADA_MS) {
      const txt = conteudoJanela(e.rateLimits, agora)
      if (txt !== undefined) {
        janelaGravadaEm = agora
        await gravar($, ARQ_JANELA, txt).catch(() => undefined)
      }
    }
    const r = avaliarAlertas(e.rateLimits, alertas, agora)
    if (r.alertas.length > 0) {
      for (const a of r.alertas) $.ui.toast(a.texto)
      alertas = r.estado
      try {
        await gravar($, ARQ_ALERTAS, `${JSON.stringify(alertas)}\n`)
      } catch {
        // o alerta já saiu; o cache só evita repetir na próxima abertura
      }
    } else {
      alertas = r.estado
    }
    return next(e)
  })

  on('turn.complete', async ($, e, next) => {
    if (await mudo($)) return next(e)
    if (e.agentId === undefined) {
      turnos += 1
      if (e.usage !== undefined) {
        usoSessao.push(e.usage)
        modeloSessao = e.usage.model
      }
      await descobrirIniciativa($)
    }
    return next(e)
  })

  // Telemetria do subagente que terminou, com o uso que o Claude Code mediu (medicao real).
  on('agent.report', async ($, e, next) => {
    if (await mudo($)) return next(e)
    if (agentesGravados.has(e.agentId)) return next(e)
    try {
      if (await telemetriaLigada($)) {
        const lista = await $.agent.list()
        const info = lista.find((a: any) => a.id === e.agentId)
        if (info !== undefined && info.usage !== undefined && info.usage.length > 0) {
          agentesGravados.add(e.agentId)
          const agora = await $.clock.now()
          const reg = montarRegistro({
            feature: featureDoRegistro(iniciativa, `agentes/${sanear(e.subagentType, 40).replace(/[^A-Za-z0-9._-]/g, '-')}`),
            phase: faseDaIniciativa(iniciativa),
            agentPath: papelDoSubagente(e.subagentType),
            uso: info.usage,
            turns: info.usage.length,
            durationSeconds: (agora - info.startTime) / 1000,
            model: info.model ?? info.usage[0]?.model ?? 'desconhecido',
            custoTabelaUsd: info.cost?.usd,
            ok: e.status === 'completed',
            sddVersion: await versaoSdd($),
          })
          await gravarRegistro($, reg)
        }
      }
    } catch {
      // telemetria nunca derruba o relatório do agente
    }
    return next(e)
  }).catch(($, e, next) => next(e))

  // Sessão interativa: um registro ao fim, com o uso somado dos turnos da conversa principal.
  on('session.end', async ($, e, next) => {
    if (await mudo($)) return next(e)
    try {
      if (interativa && turnos > 0 && usoSessao.length > 0 && (await telemetriaLigada($))) {
        const usage = await $.session.usage()
        const agora = await $.clock.now()
        const reg = montarRegistro({
          feature: featureDoRegistro(iniciativa, 'sessao-interativa'),
          phase: faseDaIniciativa(iniciativa),
          agentPath: ['orquestrador'],
          uso: usoSessao,
          turns: turnos,
          durationSeconds: (agora - inicioSessao) / 1000,
          model: modeloSessao,
          custoTabelaUsd: usage.cost?.usd,
          ok: true,
          sddVersion: await versaoSdd($),
        })
        await gravarRegistro($, reg)
      }
    } catch {
      // idem
    }
    turnos = 0
    usoSessao.length = 0
    inicioSessao = await $.clock.now()
    return next(e)
  })

  // Aviso de orçamento (RN-5): só avisa; nunca bloqueia o spawn.
  on('agent.spawn', async ($, e, next) => {
    if (await mudo($)) return next(e)
    try {
      const { rateLimits } = await $.session.usage()
      const j = janelaMaisCheia(rateLimits)
      if (j !== undefined && j.percentUsed >= LIMITE_ORCAMENTO) {
        const quando = horaDeRenovacao(j.resetsAt, await $.clock.now())
        $.ui.toast(`Janela de ${rotuloJanela(j.kind)} em ${Math.round(j.percentUsed)}%: o subagente consome dela${quando ? ` (renova ${quando})` : ''}`)
      }
    } catch {
      // sem leitura de janela, sem aviso
    }
    const r = await next(e)
    if (e.tool_use_id && typeof (r as any)?.agentId === 'string') {
      await update($, cartoes, c => ({ ...c, [e.tool_use_id]: (r as any).agentId })).catch(() => undefined)
    }
    return r
  }).catch(($, e, next) => next(e))

  on('command.run', { command: 'painel' }, async ($, e) => {
    const partes = (e.args ?? '').trim().split(/\s+/).filter(p => p !== '')
    if (partes[0] === 'log') {
      const [, papel, id] = partes
      if (!(PAPEIS as readonly string[]).includes(papel ?? '') || id === undefined || !RE_ID_JOB.test(id)) {
        return { text: 'Uso: /painel log <builder|reviewer> <id-do-job>' }
      }
      try {
        const tamanho = (await $.fs.stat(`${DIR_LOGS}/${papel}-${id}.log`)).size
        const inicio = Math.max(0, tamanho - 32768)
        const texto = await $.fs.read(`${DIR_LOGS}/${papel}-${id}.log`, { as: 'text', offset: inicio, length: 32768 })
        return { text: caudaDoLog(texto) || 'Log vazio.' }
      } catch {
        return { text: `Sem log para ${papel} ${id}.` }
      }
    }
    await abrirPainel($)
    return { text: 'Painel da esteira aberto.' }
  })

  on('command.run', { command: 'gastos' }, async ($, e) => {
    const todos = (e.args ?? '').trim() === 'todos'
    const entradas = (await $.fs.list(DIR_METRICAS).catch(() => [])) as any[]
    const nomes = entradas
      .filter(en => en.kind === 'file' && /^session-.*\.json$/.test(en.name))
      .map(en => en.name as string)
      .sort()
      .slice(-MAX_REGISTROS_GASTOS)
    const brutos: string[] = []
    for (const n of nomes) {
      try {
        brutos.push(await $.fs.read(`${DIR_METRICAS}/${n}`))
      } catch {
        continue
      }
    }
    const alvo = todos ? undefined : iniciativa
    return { text: textoDoResumo(resumir(brutos, alvo), alvo) }
  })

  on('ui.render', { component: 'Pane', requestId: PAINEL }, async ($, e) => {
    const { Box, Text } = $.ui.resolve(e)
    const d = { ...VAZIA, ...(await read($, esteira)) }
    const q = d.quadro ?? null
    const col = (t: string, n: number) => (t.length > n ? `${t.slice(0, n - 1)}…` : t.padEnd(n))
    return (
      <Box flexDirection="column">
        <Text bold>Esteira{q ? ` · ${q.iniciativa}` : ''}</Text>
        {!d.temFilas && <Text dimColor>Nenhum job no barramento (rode `forge up --terminal painel`).</Text>}
        {d.temFilas && q === null && <Text dimColor>Nenhuma iniciativa em andamento.</Text>}
        {q !== null && (
          <Text dimColor>
            pares ativos {q.ativos}
            {q.total > 0 ? ` · ${q.mescladas}/${q.total} tasks mescladas` : ''}
          </Text>
        )}
        {q !== null && q.linhas.length > 0 && <Text dimColor>{`${col('task', 18)} ${col('builder', 11)} ${col('reviewer', 11)} ${col('tokens', 8)} tempo`}</Text>}
        {q?.linhas.map(l => (
          <Text key={`t-${l.task}`} bold={l.ativo} dimColor={!l.ativo}>
            {`${col(l.task, 18)} ${col(l.builder, 11)} ${col(l.reviewer, 11)} ${col(l.tokens > 0 ? tokensCurtos(l.tokens) : '·', 8)} ${l.tempo}`}
          </Text>
        ))}
        {q?.modoFeature && <Text color="yellow">⚠ um Builder para a feature inteira: o forge que despachou é anterior à 6.0 (rode `forge --version`)</Text>}
        {q !== null && <Text>Total da iniciativa {tokensCurtos(q.tokens)} tokens</Text>}
        {d.outras.length > 0 && <Text dimColor>Também em andamento: {d.outras.join(', ')}</Text>}
        <Text bold>Subagentes desta sessão</Text>
        {d.agentes.length === 0 && <Text dimColor>Nenhum.</Text>}
        {d.agentes.map((a, i) => (
          <Text key={`ag-${i}`}>
            {a.tipo} [{a.estado}] {tokensCurtos(a.tokens)} tok
          </Text>
        ))}
        {d.erro !== '' && <Text dimColor>{d.erro}</Text>}
        <Text dimColor>Tokens medidos pelo Claude Code; dólar, quando aparece, é custo de tabela.</Text>
      </Box>
    )
  })

  // Faixa acima do prompt: inbox e PR conforme sem precisar perguntar.
  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    if (await mudo($)) return next(e)
    const d = { ...VAZIA, ...(await read($, esteira)) }
    if (e.props.hasSurvey || (d.inbox?.naoLidas ?? 0) === 0) return next(e)
    const { Box, Text } = $.ui.resolve(e)
    const pr = d.inbox.prConforme.length > 0 ? ` · ${d.inbox.prConforme.slice(0, 3).join(', ')} conforme` : ''
    return (
      <Box>
        <Text>
          forge · inbox {d.inbox.naoLidas}
          {pr}
        </Text>
        <Text dimColor> (forge inbox)</Text>
      </Box>
    )
  })

  // Cartão embaixo da chamada de subagente: consumo daquele agente.
  on('ui.render', { component: 'ToolUse' }, async ($, e, next) => {
    const linha = await next(e)
    if (e.props.tool !== 'Agent' && e.props.tool !== 'Task') return linha
    const id = ((await read($, cartoes)) ?? {})[e.props.tool_use_id]
    if (id === undefined) return linha
    let info: any
    try {
      info = (await $.agent.list()).find((a: any) => a.id === id)
    } catch {
      return linha
    }
    if (info === undefined) return linha
    const { Box, Text } = $.ui.resolve(e)
    const agora = await $.clock.now()
    const ctx = typeof info.contextWindowSize === 'number' && info.contextWindowSize > 0 && typeof info.tokenCount === 'number'
      ? ` · contexto ${Math.min(100, Math.round((100 * info.tokenCount) / info.contextWindowSize))}%`
      : ''
    return (
      <Box flexDirection="column">
        {linha}
        <Text dimColor>
          {`  ● ${sanear(info.type, 30)} · ${tokensCurtos(info.tokenCount ?? 0)} tok · ${duracao(agora - (info.startTime ?? agora))}${ctx}`}
        </Text>
      </Box>
    )
  })
}
