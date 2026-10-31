export type MascotState = 'working' | 'approval' | 'thinking' | 'searching' | 'finished' | 'error';
export type MascotMark = 'ring' | 'diamond' | 'bar' | 'wave' | 'tri';

export interface StateDef {
  label: string;
  orb: number;
  cssOrb: string;
  eyes: 'cap' | 'round' | 'happy' | 'x';
}

export const STATE_DEFS: Record<MascotState, StateDef> = {
  working:  { label: 'Trabalhando',  orb: 0x4f8ef7, cssOrb: '#4f8ef7', eyes: 'cap' },
  approval: { label: 'Aguarda você', orb: 0xf5b04a, cssOrb: '#f5b04a', eyes: 'round' },
  thinking: { label: 'Planejando',   orb: 0x9b7bff, cssOrb: '#9b7bff', eyes: 'cap' },
  searching:{ label: 'Auditando',    orb: 0x38c6e8, cssOrb: '#38c6e8', eyes: 'round' },
  finished: { label: 'Concluído',    orb: 0x3ecf8e, cssOrb: '#3ecf8e', eyes: 'happy' },
  error:    { label: 'Bloqueado',    orb: 0xf2677a, cssOrb: '#f2677a', eyes: 'x' }
};

export interface StationData {
  step: string;
  shortName: string;
  name: string;
  trait: string;
  color: string;
  hex: number;
  mark: MascotMark;
  defaultState: MascotState;
  inputArtifact: string;
  outputArtifact: string;
  handoffTo: string;
  handoffPayload: string;
  handoffGate: string;
  handoffNote: string;
  desc: string;
  states: Partial<Record<MascotState, string>>;
}

export const STATIONS: StationData[] = [
  {
    step: 'Estação 01',
    shortName: '01. O Líder',
    name: 'O Líder (Orquestrador)',
    trait: 'Maestro da operação · Seu ponto de contato',
    color: '#f2a65a',
    hex: 0xf2a65a,
    mark: 'ring',
    defaultState: 'working',
    inputArtifact: 'Pedido do Humano / Laudo Camada B',
    outputArtifact: 'Demanda estruturada → Estação 02',
    handoffTo: '02. PM & Arquiteto',
    handoffPayload: 'Contexto do projeto (progress.md) + abertura de Discovery & Reuse Check',
    handoffGate: 'Passo 1 → 2 · Início da Esteira',
    handoffNote: 'O Líder nunca implementa código direto: delega à estação certa',
    desc: 'Recebe sua demanda, coordena os especialistas em sequência ou paralelo e consolida o Painel Executivo.',
    states: {
      working:  'Recebendo sua demanda e acionando o PM + Arquiteto para concepção…',
      approval: 'Painel Executivo (Camada C) pronto: aguardando sua validação final.',
      finished: 'Ciclo completo! Entrega validada e progresso atualizado.'
    }
  },
  {
    step: 'Estação 02',
    shortName: '02. PM & Arquiteto',
    name: 'Produto & Arquiteto',
    trait: 'Concepção da Spec · Reuse Check (DUP-01)',
    color: '#8aa4ff',
    hex: 0x8aa4ff,
    mark: 'diamond',
    defaultState: 'approval',
    inputArtifact: 'Demanda do Líder + Código atual',
    outputArtifact: 'discovery.md + criteria.md',
    handoffTo: '03. O Planejador',
    handoffPayload: 'discovery.md + criteria.md (com seção "O que já existe" validada por você)',
    handoffGate: 'Portão FLX-01 · Aprovação Humana da Spec',
    handoffNote: 'Bloqueio estrito: o Planejador só recebe a Spec após o seu "De Acordo"',
    desc: 'Mapeia o que já existe para evitar duplicação e estrutura os critérios de aceite antes de qualquer código.',
    states: {
      working:  'Mapeando ativos existentes no repositório (Reuse Check DUP-01)…',
      approval: 'Portão FLX-01 aberto: aguardando seu "De Acordo" na especificação.',
      finished: 'Spec aprovada por você! Passando bastão para o Planejador.'
    }
  },
  {
    step: 'Estação 03',
    shortName: '03. O Planejador',
    name: 'O Planejador',
    trait: 'Fatiamento P/M/G · Regras sob demanda',
    color: '#c9a7ff',
    hex: 0xc9a7ff,
    mark: 'bar',
    defaultState: 'thinking',
    inputArtifact: 'Spec aprovada (FLX-01)',
    outputArtifact: 'plan.md + sdd/.queue/queue.json',
    handoffTo: '04. Fila de Builders',
    handoffPayload: 'Tarefas P/M/G enfileiradas em queue.json + seleção de .agents/rules/sdd-*.md',
    handoffGate: 'Regra GOV-04 · Contexto Seletivo',
    handoffNote: 'Separa tarefas independentes (paralelo) das que têm depends_on (fila)',
    desc: 'Quebra a Spec em tarefas atômicas e anexa apenas os módulos de regras necessários para cada Builder.',
    states: {
      thinking: 'Mapeando IDs da Constituição (ARQ, SEG, TST) por subtarefa…',
      working:  'Verificando colisão de arquivos para liberar execução em paralelo…',
      finished: 'Fila gravada em sdd/.queue/queue.json! Despachando para os Builders.'
    }
  },
  {
    step: 'Estação 04',
    shortName: '04. Builders',
    name: 'Fila de Builders',
    trait: 'Execução isolada · Paralelismo seguro',
    color: '#4fd1b5',
    hex: 0x4fd1b5,
    mark: 'wave',
    defaultState: 'working',
    inputArtifact: 'Tarefa da fila + Rules do escopo',
    outputArtifact: 'Código + Testes + Portão A (PASS)',
    handoffTo: '05. Governança Read-Only',
    handoffPayload: 'Diff completo da implementação + relatório determinístico forge-sdd gate (P-01..P-23)',
    handoffGate: 'Portão Camada A · Testes & Segredos',
    handoffNote: 'Só avança para o Analista de Governança quando go test e P-13 passam',
    desc: 'Consome a fila determinística: tarefas sem conflito rodam em paralelo; dependências aguardam liberação.',
    states: {
      working:  'Construindo em paralelo tarefas sem conflito de arquivos…',
      approval: 'Executando Portão A (forge-sdd gate): validando testes e limites…',
      finished: 'Portão A aprovado! Entregando pacote para auditoria da Governança.'
    }
  },
  {
    step: 'Estação 05',
    shortName: '05. Governança',
    name: 'Analista de Governança',
    trait: 'Auditoria Somente Leitura · Camada B',
    color: '#ff8a5c',
    hex: 0xff8a5c,
    mark: 'tri',
    defaultState: 'searching',
    inputArtifact: 'Diff + Spec + IDs Constituição',
    outputArtifact: 'Laudo JSON → Líder (ou ↺ Builder)',
    handoffTo: '01. O Líder (Painel Camada C)',
    handoffPayload: 'Veredito JSON ("conforme") + proposta de Auto-Skill (DUP-05) → Painel do Líder',
    handoffGate: 'Portão Camada B → Camada C',
    handoffNote: 'Se reprovar, devolve direto ao Builder (05 ↺ 04); se aprovar, fecha no Líder',
    desc: 'Proibido de editar código. Audita arquitetura, segurança e reuso antes de devolver o laudo ao Líder.',
    states: {
      searching:'Inspecionando conformidade com ARQ-*, SEG-*, DUP-* e TST-*…',
      error:    'Não-conformidade detectada: devolvendo tarefa para a fila do Builder (05 ↺ 04).',
      finished: 'Veredito CONFORME! Retornando resumo executivo e Auto-Skill para o Líder.'
    }
  }
];
