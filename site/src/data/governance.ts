import type { MascotState } from './stations';

export interface GovScenario {
  k: string;
  state: MascotState;
  chipLabel: string;
  chipColor: string;
  ruleCode: string;
  title: string;
  sub: string;
  termHtml: string;
  fixTitle: string;
  fixTerm: string;
}

export const GOV_SCENARIOS: GovScenario[] = [
  {
    k: 'var(--bad)',
    state: 'error',
    chipLabel: 'Bloqueado na Auditoria',
    chipColor: 'var(--bad)',
    ruleCode: 'SEG-01 · P-13',
    title: 'Chave privada detectada no diff da tarefa',
    sub: 'O Analista de Governança opera sem permissão de escrita: ele bloqueia a passagem e devolve a evidência exata para o Builder corrigir (05 ↺ 04).',
    termHtml: `
      <div>$ forge-sdd gate --mode bloqueio</div>
      <span class="bad-line">- [P-13 FALHOU] Credencial exposta detectada em internal/auth/client.go:42</span>
      <span class="bad-line">- Veredito Camada B: "bloqueado" (violação SEG-01 — retorna à Estação 04)</span>
    `,
    fixTitle: 'Segredo removido pelo Builder · Reinspeção aprovada (05 → 01)!',
    fixTerm: `
      <div>$ forge-sdd gate --mode bloqueio</div>
      <span class="ok-line">+ [P-13 PASS] Zero segredos expostos (lendo via variável de ambiente)</span>
      <span class="ok-line">+ Veredito Camada B: "conforme" (liberado para o Painel do Líder)</span>
    `
  },
  {
    k: 'var(--warn)',
    state: 'approval',
    chipLabel: 'Retornado para Reuso',
    chipColor: 'var(--warn)',
    ruleCode: 'DUP-01 · DUP-02',
    title: 'Função duplicada ignorou módulo que já existia no projeto',
    sub: 'O Auditor cruza a entrega com o Reuse Check da Spec e exige que o Builder reaproveite o pacote existente em vez de duplicar código.',
    termHtml: `
      <div>$ forge-sdd governance audit --feature feat-58</div>
      <span class="bad-line">- [DUP-01 RESSALVA] Novo helper criado em internal/util/hash.go duplica internal/gate</span>
      <span class="bad-line">- Ação exigida: reutilizar função canônica mapeada na Spec (retorna 05 ↺ 04)</span>
    `,
    fixTitle: 'Código refatorado reaproveitando o módulo existente!',
    fixTerm: `
      <div>$ forge-sdd governance audit --feature feat-58</div>
      <span class="ok-line">+ [DUP-01 PASS] Reuso confirmado conforme seção "O que já existe" da Spec</span>
      <span class="ok-line">+ Veredito Camada B: "conforme"</span>
    `
  },
  {
    k: 'var(--bad)',
    state: 'error',
    chipLabel: 'Reprovado sem Testes',
    chipColor: 'var(--bad)',
    ruleCode: 'ARQ-01 · TST-01',
    title: 'Regra de negócio dentro de cmd/ e sem teste unitário',
    sub: 'Como o Auditor tem as mãos atadas (Read-Only), ele não faz gambiarras silenciosas: reprova a entrega e exige separação limpa em internal/.',
    termHtml: `
      <div>$ forge-sdd gate --mode bloqueio</div>
      <span class="bad-line">- [ARQ-01 FALHOU] Lógica de domínio detectada em cmd/forge-sdd/queue.go</span>
      <span class="bad-line">- [TST-01 FALHOU] Pacote alterado sem cobertura *_test.go correspondente</span>
    `,
    fixTitle: 'Lógica movida para internal/ com 100% dos testes verdes!',
    fixTerm: `
      <div>$ forge-sdd gate --mode bloqueio</div>
      <span class="ok-line">+ [P-23 PASS] go test ./... & go vet ./... passaram sem avisos</span>
      <span class="ok-line">+ Veredito Camada B: "conforme" (fronteira cmd/ vs internal/ íntegra)</span>
    `
  },
  {
    k: 'var(--ok)',
    state: 'finished',
    chipLabel: 'Selo Conforme & Auto-Skill',
    chipColor: 'var(--ok)',
    ruleCode: 'FLX-03 · DUP-05',
    title: 'Entrega íntegra aprovada nas Camadas A e B',
    sub: 'O pacote sobe limpo para o Painel Executivo do Líder e o padrão repetido 3x vira uma Auto-Skill sincronizada para toda a equipe.',
    termHtml: `
      <div>$ forge-sdd skills approve novo-subcomando-cli-sdd</div>
      <span class="ok-line">+ [PORTÃO A & B PASS] Arquitetura, Segurança, Reuso e Testes validados</span>
      <span class="ok-line">+ [AUTO-SKILL ATIVA] Sincronizada em .agents/, .claude/ e .gemini/</span>
    `,
    fixTitle: 'Padrão já ativo e sincronizado para os próximos projetos!',
    fixTerm: `
      <div>$ forge-sdd skills list</div>
      <span class="ok-line">+ novo-subcomando-cli-sdd [status: active] (3 evidências validadas)</span>
    `
  }
];
