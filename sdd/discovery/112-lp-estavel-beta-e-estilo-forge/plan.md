# Plano — Iniciativa 112

Ordem: 01 antes de 02 (as duas editam `docs/index.html`; a 01 remove o bloco da edição, a 02 mexe em chips, terminal, rodapé e instalação); a 03 vem depois das duas porque cria uma seção nova antes do rodapé e reaproveita os chips de versão da 02. Até 7 tasks por onda.

## Onda 1 — Encerrar a edição
- **feat-112-01** Encerrar a Edição Missões antes do prazo. **Depende de:** nenhuma. Rules: COD-04, TST-01.
  Tasks: (1) branch; (2) `bash .github/scripts/encerrar-edicao/encerrar.sh .`; (3) apagar o workflow (D-3); (4) ajustar `TestScaffold_DocsEdicao`; (5) rodar a prova de grep; (6) `go test ./...`.

## Onda 2 — Versões
- **feat-112-02** Estável e beta na LP, com npm ao vivo e fallback. **Depende de:** feat-112-01 (mesmo arquivo; sem a 01, a 02 colide nas linhas ~1099-1103 e ~1774-1776). Rules: SEG-01, SEG-07, TST-01.
  Tasks: (1) HTML estático com os dois valores; (2) bloco `/* versoes:inicio */` com função pura; (3) chamada com `AbortController`; (4) comando de instalação beta e aviso; (5) frase de privacidade; (6) `lp_versoes_test.go`; (7) nota no checklist (com o dono).

## Onda 3 — Novidades
- **feat-112-03** Seletor de novidades da versão na LP, alimentado pelo `history.md`. **Depende de:** feat-112-01 e feat-112-02 (mesmo arquivo; usa os `data-versao` da 02 e ocupa o espaço liberado pela 01). Rules: DUP-01, DUP-02, DUP-10, SEG-01, SEG-02, SEG-05, SEG-07, TST-01, TST-05, FLX-07.
  Tasks: (1) `scripts/novidades_lp.py` com `--check`; (2) gerar e versionar `docs/data/releases.json`; (3) seção `#novidades` com abas e lista compacta antes do `footer`; (4) `buscarNovidades()` com fallback JSON → raw → cache → vazio; (5) passo no `deploy-lp.yml`; (6) passo no `release.yml` (gerar e commitar o JSON, `[skip ci]`); (7) `lp_novidades_test.go` com os quatro testes.

Publicação: depende do deploy da Cloudflare corrigido (token sem Workers Scripts/Routes em Editar). A publicação manual funcionou em 21/11/2026.