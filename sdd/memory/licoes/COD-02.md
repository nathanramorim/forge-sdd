# Lições Aprendidas — COD-02

Lições associadas à rule `COD-02` da Constituição (mais recentes primeiro).

- encadear 'gh pr ready' e 'gh pr merge' sem conferir os checks mesclou a 118-07 com o lint vermelho (variável sem uso); o ci-local.sh não roda golangci-lint, que não está instalado na máquina → conferir os 3 jobs (test, lint, pronto?) verdes antes do merge e nunca encadear ready e merge; instalar o golangci-lint local para o ci-local.sh cobrir o lint (sdd/discovery/118-upgrade-completo-da-beta/status.md)
- alterações em pacotes Go ou assinaturas sem rodar `go vet ./...` antes de encerrar a task → rodar `go vet ./...` ao final de cada alteração para capturar problemas estáticos cedo (sdd/memory/constitution.md)
