# HOWTO — forge-sdd

Tutoriais para humanos. Não é lido por agentes.

## Iniciar uma sessão
1. Abra o VS Code neste repositório
2. Use `/proxima-feature` para continuar de onde parou
3. O Orquestrador lê `sdd/memory/progress.md` e reporta o próximo passo

## Criar nova feature
```
/nova-feature <descrição>
```

## Health check
```
/doctor
```

## Compactar progress.md
```
/archive
```

## Upgrade da versão Forge-SDD
```
/upgrade-sdd 1.2.0
```

## Layout do projeto Go
```
cmd/forge-sdd/     → entrypoint cobra
internal/config/   → struct Config + defaults
internal/scaffold/ → renderiza templates + escreve arquivos
internal/survey/   → prompts interativos (huh)
templates/         → artefatos embutidos via embed.FS
tests/fixtures/    → golden test fixtures
```

## Rodar testes
```bash
go test ./...
```

## Release local (snapshot)
```bash
goreleaser build --snapshot --clean
```
