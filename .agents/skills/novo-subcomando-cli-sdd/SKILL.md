---
name: novo-subcomando-cli-sdd
description: Procedimento padronizado (DUP-05) para adicionar um novo subcomando CLI Go no forge-sdd mantendo fronteiras cmd/ vs internal/ e testes
status: approved
origin: auto-governance
---

# Skill: novo-subcomando-cli-sdd

1. Criar a lógica de domínio em internal/<pacote>/ sem importar cobra.\n2. Criar testes unitários em internal/<pacote>/<pacote>_test.go.\n3. Criar o adaptador Cobra em cmd/forge-sdd/<subcomando>.go e registrar em rootCmd.AddCommand no init() de main.go.\n4. Rodar go test ./... e go vet ./... antes de marcar complete na fila.
