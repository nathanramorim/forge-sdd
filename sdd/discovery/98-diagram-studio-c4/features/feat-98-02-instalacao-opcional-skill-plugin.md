# feat-98-02 — instalação opcional de skill/plugin, limitada ao projeto aberto

**Iniciativa:** 98-diagram-studio-c4
**Status:** draft
**Tamanho:** M
**Depende de:** feat-98-01

## Objetivo

Definir e implementar em Go a política do dono: instalar skill ou plugin é **opcional por skill**, **limitado ao projeto aberto** e, havendo forge em subprojetos, **pergunta hub (master) ou subprojeto**. Primeiro consumidor: `diagramas-c4-mermaid` (feat-98-01) e o plugin `diagram-studio`.

## Modelo do módulo

**Completo** (ARQ-11): grava arquivos no projeto do usuário e valida caminhos, então precisa de função Go testada. Pacote dono: `internal/scaffold`, que já concentra `shouldPreserve`, `checkSafeTarget` e `WriteFileSafe` (SEG-02). Arquivo novo `internal/scaffold/instalar_skill.go`; sem pacote novo e sem dependência Go nova.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Escrita segura (alvo, symlink, atômica) | `internal/scaffold/hooks.go` | `checkSafeTarget`, `writeFileSafe`, `WriteFileSafe` (tmp exclusivo + rename; recusa symlink) | Reusar | Evita duplicar SEG-02 (DUP-01) |
| Preservação do usuário | `internal/scaffold/scaffold.go` | `shouldPreserve(dest, targetDir)`; hoje protege só `sdd/`; `.agents/skills/` não é coberta | Estender | A skill já existente do usuário em `.agents/skills/` não pode ser sobrescrita |
| Skills por slug | `internal/skills/skills.go` | `Propose`, `Approve`, `Sync`, `List` (skills propostas, não instalação de skill pronta) | Não reusar | Outro contrato (proposta/aprovação); mantém fronteira |
| Subprojetos / hub | `internal/` e `cmd/` (grep por subprojeto/subproject) | nenhum código existente | Criar | Listagem de subprojetos com forge é nova |
| Pergunta ao usuário | skills e comandos `.claude/commands/` | menus via AskUserQuestion no agente | Reusar | Go só valida e grava; a pergunta é do agente, que passa o destino escolhido |

## Decisões do projeto aplicadas

- **SEG-01 / SEG-02**: destino validado na borda; escrita atômica; sem symlink nem path traversal; `shouldPreserve`.
- **SEG-07**: ver Dependências novas.
- **DUP-01**: reusa `WriteFileSafe`.
- **FLX-05 / GOV-01**: a política vale para todo o forge (qualquer skill ou plugin instalado), logo é decisão estrutural. **Follow-up declarado:** criar nova ADR em `sdd/spec/adr/` (próximo número livre, após adr-0023) registrando a política. A ADR é entregue por esta feature no mesmo PR ou em PR dedicado antes do merge; sem ela o Reviewer deve bloquear.

## Conteúdo da alteração

Arquivo `internal/scaffold/instalar_skill.go` (e `_test.go`):

- `ListarDestinosInstalacao(projetoAberto string) ([]string, error)`: devolve o projeto aberto e, se houver, os subprojetos com forge (diretórios filhos com `sdd/.sddrc`), ordenados. Sem subprojetos devolve só o projeto aberto (o agente instala sem perguntar).
- `InstalarSkill(projetoAberto, destino, slug string, arquivos map[string][]byte) error`: instala em `<destino>/.agents/skills/<slug>/`.

Regras:

1. **Opcional por skill:** a função só é chamada para o slug que o usuário escolheu; não há instalação em lote nem implícita.
2. **Validação do destino (SEG-01/SEG-02):** `destino` precisa ser exatamente um item de `ListarDestinosInstalacao`; `slug` passa por `skills.ValidateSlug`; chaves de `arquivos` são relativas, sem `..` e sem caminho absoluto; o caminho final, resolvido com `EvalSymlinks`, precisa ficar dentro de `destino`; symlink no alvo é recusado.
3. **Limitada ao projeto aberto:** nunca grava em `~/.claude/` nem fora de `projetoAberto`.
4. **Preservação:** arquivo já existente em `.agents/skills/<slug>/` não é sobrescrito (estender `shouldPreserve` para esse prefixo); devolve a lista dos preservados.
5. **Gravação atômica:** cada arquivo via `WriteFileSafe` (tmp + rename).

## Testes a criar (TST-01)

Em `internal/scaffold/instalar_skill_test.go`:

- `TestInstalaSkillSoNoProjetoAberto`: instala com `HOME` apontando para um tmp e confirma que nada é gravado em `~/.claude/` nem fora do projeto; destino fora da lista, `slug` com `..`, chave com `..` e symlink para fora retornam erro e não gravam.
- `TestInstalaPerguntaHubOuSubprojeto`: com forge em dois subprojetos, `ListarDestinosInstalacao` devolve hub e os dois subprojetos (3 itens); sem subprojetos devolve 1 item; instalar num subprojeto grava só nele.
- `TestInstalaOpcionalPorSkill`: instalar a skill A não cria a skill B; chamar sem escolha (slug vazio) retorna erro e não grava; arquivo pré-existente do usuário em `.agents/skills/<slug>/` permanece intacto e é reportado como preservado; falha no meio não deixa arquivo `.tmp` para trás.

## Dependências novas

Nenhuma dependência Go. Dependências externas de runtime do fluxo diagram-studio (SEG-07):

| Dependência | Versão exata | Fonte |
|---|---|---|
| plugin `diagram-studio` | 0.2.0 (`plugin.json` em `~/.claude/plugins/marketplaces/diagram-studio/.claude-plugin/`, leitura local) | marketplace `nathanramorim/diagram-studio` |
| `uv` | 0.9.7 (`uv --version`, 2025-10-30) | astral-sh/uv |

Consulta ao context7/documentação oficial dessas duas versões: **versão a confirmar antes do Builder (SEG-07)** quanto à consulta; o Builder registra o resultado em `status.md` antes de qualquer uso.

## Rules e IDs da Constituição aplicáveis

`ARQ-11`, `DUP-01`, `COD-01`, `SEG-01`, `SEG-02`, `SEG-07`, `TST-01`, `FLX-05`, `GOV-01`

## Critérios de Aceitação Executáveis

```bash
# 1. Os três testes existem e passam
go test ./internal/scaffold -run 'TestInstalaSkillSoNoProjetoAberto|TestInstalaPerguntaHubOuSubprojeto|TestInstalaOpcionalPorSkill' -v

# 2. Reuso da escrita segura, sem gravação direta
grep -n "WriteFileSafe\|writeFileSafe" internal/scaffold/instalar_skill.go

# 3. ADR da política criada
ls sdd/spec/adr/ | grep -i "instalacao"

# 4. Build e suíte
go build ./... && go test ./...
```
