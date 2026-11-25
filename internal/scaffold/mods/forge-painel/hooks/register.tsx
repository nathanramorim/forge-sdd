import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import { PAPEIS, ESTADOS, DIR_FILAS, DIR_LOGS, RE_ID_JOB, linhaDoJob, tokensDoLog, caudaDoLog, rotuloEstado, tempoDecorrido, sanear } from './lib/esteira'
import type { Linha } from './lib/esteira'
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
const esteira = atom({ plugin: 'forge-painel', key: 'esteira' } as const, { jobs: [], agentes: [], erro: '' } as Esteira)

// Estado em memória do módulo: some no reload, e tudo bem (sem estado próprio além de cache).
let mudoCache: boolean | undefined
let interativa = false
let iniciativa: string | undefined
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

async function descobrirIniciativa($: EngineInterface): Promise<void> {
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
    iniciativa = melhor
  } catch {
    iniciativa = undefined
  }
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

async function atualizar($: EngineInterface): Promise<void> {
  const agora = await $.clock.now()
  const jobs: Linha[] = []
  let erro = ''
  for (const papel of PAPEIS) {
    for (const estado of ESTADOS) {
      const dir = `${DIR_FILAS}/${papel}/${estado}`
      let entradas: any[] = []
      try {
        entradas = await $.fs.list(dir)
      } catch {
        continue // fila ainda não existe: o forge up não rodou
      }
      const nomes = entradas.filter(en => en.kind === 'file' && en.name.endsWith('.json') && !en.name.startsWith('.')).map(en => en.name as string).sort()
      // Concluídos e falhos: só os últimos para a tela não crescer sem fim.
      const recorte = estado === 'done' || estado === 'failed' ? nomes.slice(-5) : nomes.slice(0, 20)
      for (const nome of recorte) {
        let bruto = ''
        try {
          bruto = await $.fs.read(`${dir}/${nome}`)
        } catch {
          continue
        }
        const l = linhaDoJob(papel, estado, bruto, agora)
        if (l === undefined) continue
        try {
          const log = await $.fs.read(`${DIR_LOGS}/${papel}-${l.id}.log`, { as: 'text', offset: 0, length: 65536 })
          l.tokens = tokensDoLog(log)
        } catch {
          // sem log ainda
        }
        jobs.push(l)
      }
    }
  }
  let agentes: Esteira['agentes'] = []
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
  try {
    await update($, esteira, () => ({ jobs, agentes, erro }))
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
    $.ui.status(linhaDeStatus(e, iniciativa))
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
    return next(e)
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
    const dados = await read($, esteira)
    const agora = await $.clock.now()
    const sala = Math.max(1, (e.viewport?.rows ?? 24) - 6)
    return (
      <Box flexDirection="column">
        <Text bold>Jobs da esteira</Text>
        {dados.jobs.length === 0 && <Text dimColor>Nenhum job no barramento (rode `forge up --terminal painel`).</Text>}
        {dados.jobs.slice(0, sala).map(j => (
          <Text key={`${j.papel}-${j.id}`} dimColor={j.estado === 'done'}>
            {j.papel} {j.tarefa} [{rotuloEstado(j)}] {tempoDecorrido(j.desde, agora)}
            {j.tokens === undefined ? '' : ` ${j.tokens} tokens`}
          </Text>
        ))}
        <Text bold>Subagentes desta sessão</Text>
        {dados.agentes.length === 0 && <Text dimColor>Nenhum.</Text>}
        {dados.agentes.map((a, i) => (
          <Text key={`ag-${i}`}>
            {a.tipo} [{a.estado}] {a.tokens} tokens
          </Text>
        ))}
        {dados.erro !== '' && <Text dimColor>{dados.erro}</Text>}
        <Text dimColor>Tokens aqui são medidos; dólar, quando aparece, é custo de tabela.</Text>
      </Box>
    )
  })
}
