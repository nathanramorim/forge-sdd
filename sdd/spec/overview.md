# Overview — forge-sdd

CLI Go de comando único que scaffolda a estrutura Forge-SDD em qualquer projeto. Produz `.github/`, `sdd/` e `.vscode/` prontos para uso com GitHub Copilot.

## Índice
- `stack.md` — tecnologias, versões, layout
- `modules.md` — componentes internos e responsabilidades
- `flows.md` — fluxo principal do `init`
- `decisions.md` — decisões de design resolvidas e abertas

## Entrada / Saída
- **Entrada:** flags cobra + survey interativo (huh)
- **Saída:** árvore de diretórios preenchida com valores do projeto-alvo

## Escopo do binário
- `forge-sdd init` — único comando público
- Sem subcomandos de runtime (upgrade, doctor, archive = chatmodes Copilot)
- Sem acesso a rede em runtime
