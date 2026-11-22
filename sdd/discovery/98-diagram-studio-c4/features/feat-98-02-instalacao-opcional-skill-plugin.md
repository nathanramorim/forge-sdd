# feat-98-02 — instalação opcional de skill/plugin, limitada ao projeto aberto

**Iniciativa:** 98-diagram-studio-c4
**Status:** draft
**Tamanho:** M
**Depende de:** feat-98-01

## Objetivo

Definir e implementar em Go a política do dono: instalar skill ou plugin é **opcional por skill**, **limitado ao projeto aberto** e, havendo forge em subprojetos, **pergunta hub (master) ou subprojeto**. A política vale para skills **e plugins**: o plugin `diagram-studio` também fica restrito ao projeto aberto (ou ao destino hub/subprojeto escolhido) e **`/plugin install` não é usado** (grava em `~/.claude/`). Primeiro consumidor: `diagramas-c4-mermaid` (feat-98-01) e o plugin `diagram-studio`.

## Modelo do módulo

**Completo** (ARQ-11): grava arquivos no projeto do usuário e valida caminhos, então precisa de função Go testada. Pacote dono: `internal/scaffold`, que já concentra `shouldPreserve`, `checkSafeTarget` e `WriteFileSafe` (SEG-02). Arquivo novo `internal/scaffold/instalar_skill.go` (`InstalarSkill` e `InstalarPlugin` no mesmo pacote); sem pacote novo e sem dependência Go nova.

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
- **FLX-05 / GOV-01**: a política vale para todo o forge (qualquer skill ou plugin instalado), logo é decisão estrutural. **Follow-up declarado:** criar nova ADR em `sdd/spec/adr/` (arquivo `sdd/spec/adr/adr-0025-instalacao-de-skills-e-plugins-limitada-ao-projeto.md`; 0024 já existe na branch de outra iniciativa, 0025 é o próximo livre em 2026-11-22; se outro PR tomar o 0025 antes, o Builder usa o próximo livre e ajusta o critério 3) registrando a política. A ADR é entregue por esta feature no mesmo PR ou em PR dedicado antes do merge; sem ela o Reviewer deve bloquear.

## Conteúdo da alteração

Arquivo `internal/scaffold/instalar_skill.go` (e `_test.go`):

- `ListarDestinosInstalacao(projetoAberto string) ([]string, error)`: devolve o projeto aberto e, se houver, os subprojetos com forge (diretórios filhos com `sdd/.sddrc`), ordenados. Sem subprojetos devolve só o projeto aberto (o agente instala sem perguntar).
- `InstalarSkill(projetoAberto, destino, slug string, arquivos map[string][]byte) ([]string, error)`: instala em `<destino>/.agents/skills/<slug>/` e devolve os preservados.
- `InstalarPlugin(projetoAberto, destino, slug string, arquivos map[string][]byte) ([]string, error)`: mesmas regras, grava em `<destino>/.agents/plugins/<slug>/` (caminho único usado pela feat-98-01: `.agents/plugins/diagram-studio/scripts/diagrama.py`). Compartilha o validador e o gravador com `InstalarSkill` (DUP-01); só muda o prefixo.

### Origem do plugin e como o conteúdo é obtido

- Origem: repositório `github.com/nathanramorim/diagram-studio`, versão 0.2.0 (`plugin.json`), fixada pelo commit `3996fd2e2bcdc2462e6f0d1edd7f6d18f649bc03` (o repositório não tem tags; esse commit é o HEAD de `origin/main` na consulta).
- O Go não faz rede. A obtenção é feita fora do Go, em diretório temporário: `git clone https://github.com/nathanramorim/diagram-studio <tmp>` e `git -C <tmp> checkout 3996fd2e2bcdc2462e6f0d1edd7f6d18f649bc03`; só `scripts/`, `skills/`, `config/`, `exemplos/`, `LICENSE` e `.claude-plugin/plugin.json` entram no mapa. O comando lê o `<tmp>` e chama `InstalarPlugin`; o `.git` nunca é copiado.

### Chamador (TST-01)

Ponto de entrada: o grupo **existente** `forge-sdd skills` (já listado em ARQ-06; `cmd/forge-sdd/skills.go`), com dois subcomandos novos, sem comando de topo novo:

- `forge-sdd skills destinos`: imprime `ListarDestinosInstalacao` (o agente pergunta hub ou subprojeto só se houver mais de um).
- `forge-sdd skills instalar <skill|plugin> <slug> --de <diretório-origem> [--destino <dir>]`: lê `--de` e chama `InstalarSkill` ou `InstalarPlugin`; o `--destino` omitido vale o projeto aberto.

Quem aciona: a skill `diagramas-c4-mermaid` (passo 1 do fluxo B da feat-98-01), por escolha do usuário. O handler em `cmd/` só faz parse e chama `internal/scaffold` (ARQ-10). Como altera o contrato de `skills` listado em ARQ-06, a ADR nova registra os dois subcomandos e o Builder atualiza a linha do ARQ-06 se o Reviewer exigir.

Regras:

1. **Opcional por skill:** a função só é chamada para o slug que o usuário escolheu; não há instalação em lote nem implícita.
2. **Validação do destino (SEG-01/SEG-02):** `destino` precisa ser exatamente um item de `ListarDestinosInstalacao`; `slug` passa por `skills.ValidateSlug`; chaves de `arquivos` são relativas, sem `..` e sem caminho absoluto; o caminho final, resolvido com `EvalSymlinks`, precisa ficar dentro de `destino`; symlink no alvo é recusado.
3. **Limitada ao projeto aberto:** nunca grava em `~/.claude/` nem fora de `projetoAberto`.
4. **Preservação:** arquivo já existente em `.agents/skills/<slug>/` não é sobrescrito (estender `shouldPreserve` para os prefixos `.agents/skills/` e `.agents/plugins/`); devolve a lista dos preservados.
5. **Gravação atômica:** cada arquivo via `WriteFileSafe` (tmp + rename).

## Testes a criar (TST-01)

Em `internal/scaffold/instalar_skill_test.go`:

- `TestInstalaSkillSoNoProjetoAberto`: instala com `HOME` apontando para um tmp e confirma que nada é gravado em `~/.claude/` nem fora do projeto; destino fora da lista, `slug` com `..`, chave com `..` e symlink para fora retornam erro e não gravam.
- `TestInstalaPerguntaHubOuSubprojeto`: com forge em dois subprojetos, `ListarDestinosInstalacao` devolve hub e os dois subprojetos (3 itens); sem subprojetos devolve 1 item; instalar num subprojeto grava só nele.
- `TestInstalaPluginSoNoProjetoAberto`: instala `diagram-studio` com `HOME` num tmp; grava só em `<destino>/.agents/plugins/diagram-studio/` (inclui `scripts/diagrama.py`); nada em `~/.claude/` nem fora do projeto; chave com `..`, caminho absoluto, destino fora da lista e symlink para fora retornam erro e não gravam; arquivo existente é preservado e reportado.
- `TestInstalaOpcionalPorSkill`: instalar a skill A não cria a skill B; chamar sem escolha (slug vazio) retorna erro e não grava; arquivo pré-existente do usuário em `.agents/skills/<slug>/` permanece intacto e é reportado como preservado; falha no meio não deixa arquivo `.tmp` para trás.

## Dependências novas

Nenhuma dependência Go. Dependências externas de runtime do fluxo diagram-studio (SEG-07), consultadas em 2026-11-22:

| Dependência | Versão exata | Fonte |
|---|---|---|
| plugin `diagram-studio` | 0.2.0, commit `3996fd2e2bcdc2462e6f0d1edd7f6d18f649bc03` | `plugin.json` e README lidos localmente (somente leitura); `git ls-remote` do repositório confirma o mesmo commit no HEAD e nenhuma tag; página de releases do GitHub: sem releases |
| `uv` | 0.9.7 (`uv --version`, 2025-10-30) | docs.astral.sh/uv/guides/scripts/: `uv run` respeita metadados inline do script (versão do Python e dependências); a página não cita versão mínima do uv. O `diagrama.py` declara `requires-python >=3.10` e `resvg-py>=0.1` |

Resultado: o context7 não estava disponível para esta sessão, então a consulta usou a documentação oficial do uv e os arquivos do plugin. **Pendente (SEG-07):** `resvg-py>=0.1` é faixa aberta, não versão exata; o Builder fixa a versão exata do `resvg-py` no `diagrama.py` copiado ou no comando antes de usar, e confirma no context7 se houver ferramenta disponível.

## Rules e IDs da Constituição aplicáveis

`ARQ-11`, `DUP-01`, `COD-01`, `SEG-01`, `SEG-02`, `SEG-07`, `TST-01`, `FLX-05`, `GOV-01`

## Critérios de Aceitação Executáveis

```bash
# 1. Os quatro testes existem e passam
go test ./internal/scaffold -run 'TestInstalaSkillSoNoProjetoAberto|TestInstalaPluginSoNoProjetoAberto|TestInstalaPerguntaHubOuSubprojeto|TestInstalaOpcionalPorSkill' -v

# 2. Reuso da escrita segura, sem gravação direta
grep -n "WriteFileSafe\|writeFileSafe" internal/scaffold/instalar_skill.go

# 3. ADR da política criada
ls sdd/spec/adr/adr-0025-instalacao-de-skills-e-plugins-limitada-ao-projeto.md

# 3b. Chamador existe e a spec não usa /plugin install
go run ./cmd/forge-sdd skills destinos
! grep -rn "/plugin install" .agents/skills/diagramas-c4-mermaid/

# 4. Build e suíte
go build ./... && go test ./...
```
