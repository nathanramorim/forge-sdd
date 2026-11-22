# fix-106-04 — A visualização pré-Gate 1 abre o editor e a jornada acaba

**Branch:** `fix/106-04-visualizador-de-artefatos`
**Origem:** Iniciativa 106, `feat-106-26a` (comandos de concepção) e `sdd/memory/clarify.md` (seção "Visualização pós-spec")
**Tamanho:** `L`
**Status:** `spec` (Gate 1 aprovado em 22/11/2026, `[Seguir]` com a opção C)

## Sintoma

O dono pede para visualizar os artefatos antes do Gate 1, escolhe "Abrir no VS Code" e a jornada para ali: o editor abre, ele lê, e **não há caminho de volta** para aprovar. O comando não retorna ao chat com uma decisão.

Observado na sessão de 22/11/2026, duas vezes (fix-80-01 e fix-106-03): depois do `code ...`, o agente respondeu com o Gate 1 em **texto solto**, e o dono teve de pedir de novo uma forma de decidir.

Três falhas somadas:

1. **Menu de uso único.** `clarify.md` termina em "Execute apenas a ação escolhida. Depois, prossiga normalmente ao Gate 1". Não há laço: ver e decidir são dois momentos separados.
2. **Gate 1 sem menu.** `[Seguir]`/`[Ajustar]` não é descrito como `AskUserQuestion` em nenhum arquivo. Contraria a regra do dono de que todo handoff é menu.
3. **Visualização incompleta.** O VS Code só renderiza Mermaid com a extensão `bierner.markdown-mermaid` (o próprio `clarify.md` avisa), e abre uma pasta crua: sem árvore pensada para a spec, sem tudo junto. O repositório não tem comando de visualização nem código que renderize Mermaid.

## Feature Correlacionada

- **Arquivo de origem:** `sdd/memory/clarify.md` (linhas da seção "Visualização pós-spec") e os passos finais de `.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md` e `.agents/skills/sdd-planejamento/references/split-features.md`.
- **Causa-raiz:** a visualização foi pensada como "abrir alguma coisa" (Finder, VS Code, lista), não como **etapa de uma decisão**. Como o menu não volta, o Gate fica órfão.
- **Evidência do defeito:** `grep -n 'Execute apenas\|prossiga' sdd/memory/clarify.md` devolve a linha 37; `forge --help` não lista comando de visualização; `grep -rIn -i mermaid --include='*.go' cmd internal` não devolve código.

## Esclarecimento (clarify)

Pedido do dono, registrado:

- Visualização **completa**: markdown renderizado e Mermaid quando houver.
- Ver **os demais arquivos da pasta** da iniciativa (discovery, criteria, plan, features, fixes, status).
- **Voltar ao chat** e decidir: ver, depois Gate 1.
- Preferência por ver "aqui". **Limite técnico:** o terminal do Claude Code não renderiza Mermaid inline. A opção realista é uma página no navegador, com o agente aguardando no chat o menu do Gate.
- Base `main`, worktree separada (a árvore principal está suja na `feat/115-02`).

## Modelo do módulo

**Completo** (`ARQ-11`): pacote novo `internal/view` (montar a árvore e a página), subcomando novo `forge view` (adaptador fino em `cmd/forge-sdd`) e um asset HTML embutido por `embed.FS`. A dependência aponta para dentro: `cmd` → `view`. `view` não importa `cmd`.

## Decisões do projeto aplicadas

- `ARQ-06` — a CLI expõe subcomandos determinísticos. **`forge view` é subcomando novo, então a Constituição muda** e exige **ADR no mesmo PR** (`FLX-05`, `GOV-01`).
- `ARQ-03` / `ARQ-04` — o asset HTML entra por `embed.FS`; nenhuma dependência Go nova no binário. As bibliotecas de navegador (markdown e Mermaid) são **assets**, não módulos Go.
- `ARQ-05` — `--dry-run` não grava nada.
- `SEG-01` — o id da iniciativa é validado na borda: caminho com `..` ou fora de `sdd/discovery/` é recusado.
- `SEG-07` — sem dependência de runtime nova; ver a decisão do CDN abaixo.
- **Segurança do conteúdo:** o texto da spec é dado não confiável (escrito por agentes). O HTML **nunca executa** nada vindo dele: tudo é escapado, Mermaid em `securityLevel: strict`, markdown sem HTML cru.
- `DUP-01` — um só ponto de visualização para os quatro comandos de concepção.
- Regra do dono: todo handoff é menu (`AskUserQuestion`).

## Decisão técnica: script Python ou subcomando?

O pedido sugeria avaliar as duas. **Resultado: o script Python é inviável aqui.**

| Opção | Vantagem | Problema | Veredito |
|---|---|---|---|
| Script em `scripts/` (`uv run scripts/ver.py`) | Não toca `ARQ-06`; sem ADR; precedente (`licoes.py`, `pr_guard.py`) | **Não chega ao usuário.** Não há nenhum `.py` em `internal/scaffold/templates`; os scripts existem só no repositório do forge-sdd. O `/discovery` roda nos projetos dos usuários, onde o script não existe | **descartada** |
| Subcomando `forge view` | Vai dentro do binário, em todo projeto; testável em Go; sem Python nem `uv` na máquina do usuário | Muda `ARQ-06` e pede ADR | **escolhida** |

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Menu de visualização | `sdd/memory/clarify.md` seção "Visualização pós-spec" | menu de uso único, macOS-only para `open` | **reescrever** como laço (ver, depois Gate) |
| Passos finais dos 4 comandos | `references/{discovery,nova-feature,novo-fix,split-features}.md` | cada um repete "AskUserQuestion com abrir pasta/VS Code/lista/pular" | **trocar** por um ponteiro único para o `clarify.md` (`DUP-01`) |
| Diagrama Mermaid | skills de usuário `diagramas-c4-mermaid` e `blueprint-navegavel` | geram HTML, mas **estão fora do repositório** e não chegam ao projeto | **não depender**; só como referência de formato |
| Renderização de Markdown/Mermaid | Go (`cmd`, `internal`) | **nenhuma** | **criar** em `internal/view` |
| Abrir arquivo no sistema | Go: `open`/`xdg-open`/`start` | nenhum helper | **criar** pequeno, com `--no-open` e caminho impresso como fallback |
| Pasta de saída ignorada pelo git | `.gitignore` (`.forge/`) | `.forge/` já é ignorado | **gravar** em `.forge/runtime/view/<id>.html`, fora do versionamento |
| Script Python de apoio | `scripts/*.py` | só no repositório do forge-sdd | **não usar** (acima) |

## Dependências novas (`SEG-07`)

`Nenhuma` no Go. **Decisão do dono no Gate 1 (22/11/2026): opção C, híbrido.** Como a página obtém o renderizador de Mermaid e de Markdown:

| Opção | Custo | Offline |
|---|---|---|
| **A. CDN** (`marked` e `mermaid` por URL) | binário pequeno | **não** renderiza sem internet; cai no fallback (código-fonte do diagrama em `<pre>`) |
| **B. Embutir** os dois `.js` em `embed.FS` | binário **cresce** (Mermaid minificado é da ordem de MB) | funciona offline |
| **C. Híbrido** | embute `marked` (pequeno) e usa CDN só para Mermaid | Markdown sempre; Mermaid com internet |

**Escolhida: C.** Markdown, que está em toda spec, nunca depende de rede; Mermaid, que aparece em poucas, degrada com elegância. **O CDN carrega código de terceiros no navegador do dono:** a URL fica fixa com versão exata e hash de integridade (`SRI`), e o HTML é gerado localmente, nunca servido.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-02`, `ARQ-03`, `ARQ-04`, `ARQ-05`, `ARQ-06`, `ARQ-11`, `COD-03`, `DUP-01`, `FLX-01`, `FLX-05`, `GOV-01`, `SEG-01`, `SEG-07`, `TST-01`, `TST-03`, `TST-05`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Escopo

**Dentro:**
1. **`forge view <id-da-iniciativa>`** (ou `forge view` na iniciativa da branch): gera `.forge/runtime/view/<id>.html` e abre no navegador. Flags: `--no-open`, `--dir`, `--dry-run`. Retorna **imediatamente**; não prende o terminal.
2. **A página:** árvore de arquivos à esquerda (todos os `.md` e `.json` da pasta, com `discovery`, `criteria`, `plan`, `status`, `features/` e `fixes/` em destaque), conteúdo à direita com Markdown renderizado; blocos ` ```mermaid ` renderizados; busca simples por nome; link para o arquivo no disco.
3. **O laço da jornada** em `clarify.md`, para os quatro comandos (`/discovery`, `/nova-feature`, `/novo-fix`, `/split-features`):
   - **Menu A (ver):** abrir no navegador (`forge view`) · abrir no VS Code · abrir pasta · listar aqui · pular.
   - **Menu B (Gate 1), sempre depois de A:** `[Seguir]` · `[Ajustar]` · `[Ver de novo ou outro arquivo]` (volta ao Menu A) · `[Decidir depois]`. Tudo por `AskUserQuestion`.
4. Corrigir a nota "macOS-only": `forge view` abre no navegador em macOS, Linux e Windows, com fallback de imprimir o caminho.
5. **ADR** do subcomando novo, e a Constituição passa a listar `view` em `ARQ-06`.
6. Goldens e templates dos quatro comandos regenerados.

**Fora:**
- Renderizar Mermaid dentro do terminal do Claude Code (não é possível).
- Editar a spec pela página (é só leitura).
- Servidor local ou porta aberta: a página é um arquivo estático.
- Salvar o HTML no git.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. Árvore e página: lista os arquivos da iniciativa e embute cada um.
go test ./internal/view/... -run 'TestMontaArvoreDaIniciativa|TestPaginaTrazTodosOsArquivos'
# 2. Segurança: conteúdo da spec é escapado; id com '..' é recusado.
go test ./internal/view/... -run 'TestEscapaConteudoDaSpec|TestIdForaDeDiscoveryRecusado'
# 3. Mermaid: bloco vira diagrama; sem renderizador, cai no código-fonte.
go test ./internal/view/... -run 'TestBlocoMermaidMarcado|TestFallbackMostraFonteDoDiagrama'
# 4. Comando: gera o arquivo e retorna sem abrir quando --no-open; --dry-run não grava.
go test ./cmd/forge-sdd/... -run 'TestViewGeraHTML|TestViewDryRunNaoGrava|TestViewNaoPrendeOTerminal'
# 5. A jornada é um laço: o clarify.md descreve os dois menus e o retorno.
test "$(grep -c 'Ver de novo' sdd/memory/clarify.md)" -ge 1
test "$(grep -l 'clarify.md' .agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md .agents/skills/sdd-planejamento/references/split-features.md | wc -l)" = 4
! grep -n 'Execute apenas a ação escolhida' sdd/memory/clarify.md
# 6. ARQ-06 e ADR no mesmo PR.
grep -q 'view' sdd/memory/constitution.md
test -n "$(git diff --name-only origin/main...HEAD -- sdd/spec/adr/)"
# 7. Sem dependência Go nova.
go mod tidy && git diff --exit-code go.mod go.sum
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** uma spec com `<script>alert(1)</script>` e `<img onerror=...>` aparece **escapada** no HTML (nenhuma tag executável vinda do conteúdo). Um id `../../etc` e um id inexistente recebem erro claro, sem gravar arquivo.
- **Prova do laço:** o teste do `clarify.md` falha se o texto "uma vez e termina" voltar.

## Riscos

- **R1 — Binário maior** se o Mermaid for embutido. Mitigação: recomendação C (híbrido); medir o tamanho antes e depois no PR.
- **R2 — Navegador não abre** (SSH, sem display). Mitigação: `--no-open` e caminho impresso; o menu A continua com VS Code, pasta e lista.
- **R3 — CDN indisponível ou comprometido.** Mitigação: versão fixa + `SRI`; fallback mostra o código-fonte do diagrama; o conteúdo da spec nunca é executado.
- **R4 — Subcomando novo mexe em `ARQ-06`.** Mitigação: ADR no mesmo PR (`FLX-05`); o PR é M+ e passa pelo Gate 1 do dono.
- **R5 — Agentes que não são o Claude Code** (Gemini, Copilot) não têm `AskUserQuestion`. Mitigação: o `clarify.md` já prevê texto simples nesses agentes; o laço continua, só a forma muda.
