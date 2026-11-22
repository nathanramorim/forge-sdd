# Lições Aprendidas — forge-sdd

Padrões de erro já corrigidos, consultados por Builder/Revisor antes de implementar (lido no READ-MIN, feat-01-04). Entradas mais recentes primeiro; o arquivo é aparado automaticamente para respeitar o orçamento.

- criar a issue externa antes de ler a evidência local, duplicando a issue a cada reexecução → ler a evidência primeiro (fast path sem chamada externa) e só depois buscar e criar; gravar a evidência de forma atômica (feat-113-03, #238)
- teste de references que só confere a substring do comando citado (forge queue card) e passa com flag inexistente (--dir) → rodar o comando citado no texto (--dry-run) e afirmar no teste as flags reais e a ausência das inexistentes (feat-113-06, #241)
- lógica de comportamento (nomenclatura, telemetria) duplicada literalmente entre prompts de Claude/Gemini/Copilot → qualquer correção precisa lembrar de editar em 3-9 lugares; extrair para um único bloco referenciado evita drift (sdd/features/fix-50-telemetry-recording-gemini-only.md, fix-48-novo-fix-missing-copilot-agent.md)
- instrução de gravação de telemetria embutida só no último passo de um prompt longo, disparada por um único comando → mover a escrita para um mecanismo determinístico (forge-sdd session record) chamado por todos os comandos que encerram sessão (sdd/features/feat-01-simplificacao-e-aprendizado-continuo/feat-01-01-telemetria-code-enforced.md)
