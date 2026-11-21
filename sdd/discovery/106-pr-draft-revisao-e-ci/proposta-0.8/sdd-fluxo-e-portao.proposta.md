# Proposta de ajuste: `.agents/rules/sdd-fluxo-e-portao.md` (task 0.8)

Substitui o último sub-item do bloco "Fim de Onda e Release Beta (`GOV-03`, `FLX-07`)" (hoje uma linha sobre `-beta`) pelos sub-itens abaixo. Recebe o que as Regras 11 e 12 da Constituição tinham e o `FLX-07` encurtado não repete. A rule não copia o texto do `FLX-07` (P-23); só detalha.

```markdown
  - Versões `-beta` e `-rc`: documente os destaques reais e prévias de comandos em `sdd/releases/history.md`, publique a tag na própria branch da feature e mantenha o PR aberto até a promoção estável. A regra vale também para `-rc`.
  - A beta pode acumular fixes e features antes da promoção.
  - Sem destaques reais em `sdd/releases/history.md`, o workflow de release falha com o texto genérico.
  - Retenção no GitHub: as 3 releases estáveis mais recentes e as betas posteriores à mais antiga delas; o NPM não é afetado.
  - Novos comandos atualizam a documentação, com prévia e agente explícito.
```
