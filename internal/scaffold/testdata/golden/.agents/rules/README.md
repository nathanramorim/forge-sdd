# `.agents/rules/` — Regras de Domínio e Governança Modulares

Pasta neutra, fora de `.claude/`, `.gemini/`, `.github/` ou `.openai/` — qualquer agente de IA configurado neste projeto consulta os arquivos aqui sob demanda, guiado pelo frontmatter (`description`, `globs`, `regras: [IDs]`) e pelo Mapa de Roteamento em `sdd/memory/constitution.md`.

## Rules Nativas de Governança SDD (`GOV-04`)

Estes arquivos detalham **como cumprir** na prática as regras da Constituição, referenciando seus IDs estáveis sem copiar ou reescrever o texto normativo:

- `sdd-arquitetura.md` — Fronteiras `cmd/` vs `internal/`, `embed.FS`, `stdlib`, `--dry-run` e modelo de módulo (`ARQ-01..06`, `ARQ-11`).
- `sdd-seguranca.md` — Validação de borda, escrita segura, proteção de segredos/logs e dependências (`SEG-01`, `SEG-02`, `SEG-04`, `SEG-05`, `SEG-07`).
- `sdd-reuso-e-duplicacao.md` — Catálogo de conceitos, Reuse Check (`O que já existe`), Regra de Três (Auto-Skills) e fonte única (`DUP-01`, `DUP-02`, `DUP-05`, `DUP-10`, `GOV-04`).
- `sdd-codigo-e-testes.md` — Responsabilidade única, `go vet`, higiene de `TODO(<id>)`, testes executáveis e golden files (`COD-01`, `COD-02`, `COD-04`, `COD-06`, `TST-01`, `TST-02`, `TST-05`).
- `sdd-fluxo-e-portao.md` — Gate de Spec, branch única por subpasta, Portão A/B/C, ADRs, exceções temporárias e fim de onda (`FLX-01..03`, `FLX-05..08`, `GOV-01..03`).

## Convenção para Regras de Domínio do Projeto

- Um arquivo Markdown por domínio de regra adicional (ex: `design-system.md`, `frontend.md`, `accessibility.md`).
- Inclua frontmatter YAML (`description`, `globs`, `regras: [...]`) quando a regra detalhar IDs registrados em `sdd/memory/constitution.md`.
- O Planejador (`planner`) e o Líder (`lider`) selecionam apenas os arquivos relevantes para cada task da fila do Builder — não é necessário carregar todos os arquivos ao mesmo tempo.
- `forge-sdd update` preserva regras customizadas criadas pelo usuário nesta pasta.

Veja `example.md.example` para um ponto de partida adicional (arquivos `.example` são ignorados pelos agentes).
