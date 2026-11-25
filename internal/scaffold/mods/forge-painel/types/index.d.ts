// Estado do mod forge-painel: só cache do que a tela mostra; o barramento é do Go (RN-10).
export type Esteira = {
  jobs: {
    papel: string
    estado: string
    id: string
    tarefa: string
    tentativa: number
    parado: boolean
    desde?: string
    tokens?: number
  }[]
  agentes: { tipo: string; estado: string; tokens: number }[]
  erro: string
}

declare module 'claude-code' {
  interface PluginState {
    'forge-painel': { esteira: Esteira }
  }
}
