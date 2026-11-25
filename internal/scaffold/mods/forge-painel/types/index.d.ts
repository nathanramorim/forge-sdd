// Estado do mod forge-painel: só cache do que a tela mostra; o barramento é do Go (RN-10).
export type LinhaQuadro = { task: string; builder: string; reviewer: string; tokens: number; tempo: string; ativo: boolean }

export type Cartao = { nome: string; estado: string; classe: 'ok' | 'run' | 'wait' | 'fail'; tokens: number; tempo: string; jobId: string }

export type Estacao = { papel: string; estado: string; classe: 'ok' | 'run' | 'wait' | 'fail'; tokens: number; tempo: string; detalhe: string; jobId: string }

export type Esteira = {
  estacoes: Estacao[]
  progresso: { feitas: number; total: number; atual: string } | null
  quadro: {
    iniciativa: string
    linhas: LinhaQuadro[]
    cartoes: Cartao[]
    modoFeature: boolean
    ativos: number
    mescladas: number
    total: number
    tokens: number
  } | null
  outras: string[]
  agentes: { tipo: string; estado: string; tokens: number }[]
  erro: string
  inbox: { naoLidas: number; prConforme: string[] }
  temFilas: boolean
  series: Record<string, number[]>
}

declare module 'claude-code' {
  interface PluginState {
    'forge-painel': { esteira: Shaped<Esteira>; cartoes: Record<string, string> }
  }
}
