# Critérios — 102 lp-visao-pipelines

1. A LP não cita "5 estações", "5 especialistas", Líder, Planejador nem "PM & Arquiteto" como parte do produto (fora de comentários históricos e de `sdd/releases/history.md`).
2. A esteira interativa mostra exatamente 3 sessões (Discovery, Builder, Reviewer), com o fechamento "você aprova".
3. O fluxo e os comandos do emulador e do guia batem com o README e com `docs/pipeline.md`.
4. Os ids das seções e o menu de navegação continuam funcionando.
5. Nenhum arquivo em `docs/edicao/` é alterado nesta iniciativa.
6. `go test ./...` e `go vet ./...` passam.
