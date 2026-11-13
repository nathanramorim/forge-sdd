# Fix 45 — `update --upgrade` Não Detecta Versão Beta Mais Recente

**Branch:** `fix/update-beta-version-detection`
**Depende de:** —

## Descrição
O comando `forge-sdd update --upgrade` não exibe/aplica a versão beta mais recente publicada no NPM Registry (`dist-tags.beta`), mostrando apenas a versão atual do projeto e a `latest`. Investigação identificou três causas na implementação de `runUpdateFlow` (`cmd/forge-sdd/main.go`):

1. **`--upgrade` ignorado fora do modo `--yes`:** o flag só é lido dentro do bloco `if yes { ... }` (`main.go:191-193`). Sem `--yes`, `forge-sdd update --upgrade` cai silenciosamente no fluxo interativo, ignorando a intenção do flag.
2. **Fallback silencioso ao falhar a consulta ao NPM:** no fluxo interativo (`main.go:223-231`), se `config.FetchNpmVersions()` retornar erro (inclusive timeout), o código usa `latest = version` (constante local do binário) e `beta = ""`. Como `survey.RunSmartUpgradePrompt` (`internal/survey/survey.go:180-185`) só adiciona a opção "Beta" quando `betaVersion != ""`, o usuário vê apenas "manter atual" + "latest", sem indicação de que a consulta falhou.
3. **Timeout agressivo:** `config.FetchNpmVersions()` usa `http.Client{Timeout: 2 * time.Second}` (`internal/config/config.go:236-238`), curto demais para redes mais lentas, aumentando a chance do fallback do item 2.
4. **`--upgrade --yes` nunca consulta o NPM:** quando `upgradeFlag` é `true`, `targetVersion = version` (`main.go:207-208`) — a versão **compilada no binário local**, não a mais recente do registry. Ou seja, mesmo em modo não-interativo, `--upgrade` nunca pode levar a uma beta mais nova que a já instalada localmente.

## Critérios de Aceitação Executáveis

1. `forge-sdd update --upgrade` (sem `--yes`) deve ser equivalente a solicitar upgrade — não pode cair silenciosamente no fluxo interativo padrão sem ao menos considerar a intenção do flag.
2. `forge-sdd update --upgrade --yes` deve consultar o NPM Registry (`config.FetchNpmVersions`) e usar a versão `beta` ou `latest` retornada (a mais recente aplicável) como `targetVersion`, em vez de reusar a constante `version` compilada no binário.
3. Se `config.FetchNpmVersions()` falhar (erro de rede ou timeout), o comando deve **reportar a falha de forma explícita e não silenciosa** (ex: retornar erro claro ou emitir aviso destacado) em vez de degradar silenciosamente para `beta = ""` sem contexto suficiente para o usuário perceber a causa.
4. O timeout de `FetchNpmVersions` deve ser aumentado para um valor mais tolerante (ex: 5-10s) e/ou configurável.
5. Teste automatizado cobrindo: (a) `FetchNpmVersions` retorna `latest` e `beta` corretamente a partir de um mock de resposta do NPM Registry; (b) fluxo `--upgrade --yes` resolve `targetVersion` a partir dos dados buscados no registry, não da constante local; (c) falha simulada de rede resulta em erro/aviso explícito, não em fallback silencioso sem sinalização.

## Status: done

Implementado em `cmd/forge-sdd/main.go` (`runUpdateFlow`), `internal/config/config.go` (`ResolveUpgradeTarget`, timeout de `FetchNpmVersions` elevado para 8s via `npmFetchTimeout`, `NpmRegistryURL` exportado para testes) e `internal/survey/survey.go` (`RunSmartUpgradePrompt` ganhou o parâmetro `preselectUpgrade`).

- `--upgrade --yes` agora consulta `FetchNpmVersions` e resolve o alvo com `ResolveUpgradeTarget` (beta vs. latest, o mais recente aplicável) em vez de reusar a constante `version` do binário; falha de rede retorna erro explícito (`--upgrade: falha ao consultar o NPM Registry: ...`).
- `--upgrade` sem `--yes` agora pré-seleciona a opção de upgrade mais recente no prompt interativo (em vez de "manter versão atual"), honrando a intenção do flag.
- Falha ao consultar o NPM Registry no modo interativo agora emite aviso explícito e visível (`⚠ Falha ao consultar o NPM Registry...`) em vez de degradar silenciosamente para "sem beta".
- Testes: `TestFetchNpmVersions_NetworkFailureReturnsExplicitError`, `TestResolveUpgradeTarget` (`internal/config/config_test.go`), `TestUpdateCommand_UpgradeYesResolvesFromNpmRegistry`, `TestUpdateCommand_UpgradeYesNpmFailureReturnsExplicitError` (`cmd/forge-sdd/commands_test.go`). Verificação manual confirmou `update --yes --upgrade` resolvendo para a beta real publicada (`1.9.1-beta`), não para a versão compilada localmente.
