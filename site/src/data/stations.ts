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
    step: 'Sessão 01',
    shortName: '01. Discovery',
    name: 'Discovery (você conversa e decide)',
    trait: 'Concepção + Planejamento + Painel · a sessão que fala com você',
    color: '#f2a65a',
    hex: 0xf2a65a,
    mark: 'ring',
    defaultState: 'approval',
    inputArtifact: 'Sua ideia + contexto do projeto (forge brief)',
    outputArtifact: 'discovery.md + criteria.md + plan.md → job feature_build',
    handoffTo: '02. Builder',
    handoffPayload: 'Job feature_build (iniciativa, tasks e spec), criado por forge dispatch',
    handoffGate: 'Portão FLX-01 · sua aprovação da spec + gate pré-build',
    handoffNote: 'O dispatch só enfileira se o gate passar; a Discovery nunca constrói',
    desc: 'Você conversa, aprova a spec e fatia o plano em tasks. As skills de concepção, planejamento e operação (o Líder) rodam aqui, e é a única sessão que fala com você.',
    states: {
      working:  'Escrevendo discovery e critérios com o contexto do projeto…',
      thinking: 'Fatiando o plano em tasks P/M/G e mapeando dependências…',
      approval: 'Portão FLX-01 aberto: aguardando o seu "De Acordo" na spec.',
      finished: 'Job entregue ao Builder por forge dispatch.'
    }
  },
  {
    step: 'Sessão 02',
    shortName: '02. Builder',
    name: 'Builder (constrói e abre o PR)',
    trait: 'Execução em ondas · tasks sem conflito rodam juntas',
    color: '#4fd1b5',
    hex: 0x4fd1b5,
    mark: 'wave',
    defaultState: 'working',
    inputArtifact: 'Job feature_build (ou pr_rework, com o feedback do Reviewer)',
    outputArtifact: 'Código + testes verdes + PR aberto → job pr_review',
    handoffTo: '03. Reviewer',
    handoffPayload: 'Job pr_review com o PR registrado (forge pipeline pr-opened); tasks em review_pending',
    handoffGate: 'Antes do PR · build, testes e forge gate',
    handoffNote: 'Em rework, o Builder corrige só o que o Reviewer apontou',
    desc: 'Pega o job da fila, constrói as tasks em ondas (depends_on), roda os testes e o forge gate e abre o PR. Em modo live o painel segue aberto entre os jobs.',
    states: {
      working:  'Construindo tasks em ondas, sem conflito de arquivos…',
      approval: 'Rodando testes e forge gate antes de abrir o PR…',
      finished: 'PR aberto! Entregando ao Reviewer.'
    }
  },
  {
    step: 'Sessão 03',
    shortName: '03. Reviewer',
    name: 'Reviewer (confere, somente leitura)',
    trait: 'Camada A + Camada B · nunca edita nem mescla',
    color: '#ff8a5c',
    hex: 0xff8a5c,
    mark: 'tri',
    defaultState: 'searching',
    inputArtifact: 'Job pr_review: PR + spec + ADRs + rules',
    outputArtifact: 'Veredito JSON → aprova, devolve ao Builder ou escala',
    handoffTo: '01. Discovery',
    handoffPayload: 'Conforme: aviso "PR pronto" na Discovery e tasks approved. Violação: job pr_rework de volta ao Builder (até 3 vezes)',
    handoffGate: 'Camada B · somente leitura (SEG-06)',
    handoffNote: 'No limite de 3 tentativas, ou em dúvida, escala para você. Só você aprova o PR',
    desc: 'Confere o diff contra a spec, as ADRs e as rules, com evidência literal. Quem decide o fluxo é o Go: aprova e notifica, devolve ao Builder ou escala para a pessoa.',
    states: {
      searching: 'Conferindo o diff contra a spec, as ADRs e as rules…',
      error:     'Violação encontrada: devolvendo ao Builder (pr_rework, tentativa 2 de 3)…',
      finished:  'Conforme! Avisando você na Discovery: PR pronto para aprovar.'
    }
  }
];
