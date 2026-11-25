// Estado do mod forge-painel: só cache do que a tela mostra; o barramento é do Go (RN-10).
export type LinhaQuadro = { task: string; builder: string; reviewer: string; tokens: number; tempo: string; ativo: boolean }

export type Esteira = {
  quadro: {
    iniciativa: string
    linhas: LinhaQuadro[]
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
}

declare module 'claude-code' {
  interface PluginState {
    'forge-painel': { esteira: Esteira; cartoes: Record<string, string> }
  }
}
