# Como atualizar o Forge SDD

Vale para qualquer versão nova (estável ou `@beta`). Seus arquivos não são sobrescritos: quando algo seu conflita com o que mudou, a versão nova fica ao lado, num arquivo `.novo.md`, para você comparar.

## Passo a passo

**1. Instale a versão nova**
```bash
npm install -g @nathanramorim/forge-sdd          # estável
npm install -g @nathanramorim/forge-sdd@beta     # prévia
```

**2. Veja o que vai mudar, sem mexer em nada**
```bash
forge update --dry-run
```
Lista o que será criado, atualizado ou preservado.

**3. Aplique a atualização**
```bash
forge update --upgrade
```
Rode no seu terminal, fora do agente.

**4. Reconfigure o projeto**

No chat do agente (Claude Code, Gemini, etc.):
```
/upgrade-sdd
```
Leva a estrutura, as regras e os prompts novos para o seu projeto.

**5. Confira**
```bash
forge doctor
```

## O que esperar

- **Arquivos seus ficam intactos.** Instruções e regras que você editou não são trocadas; a versão nova aparece como `.novo.md`.
- **Constituição:** campos novos entram por acréscimo, sem apagar o que você escreveu.
- **CI do GitHub:** o workflow novo é opcional e só entra se você pedir (`--ci`). Um CI próprio nunca é tocado.
- **Hooks e prompts novos:** chegam pelo `/upgrade-sdd`; o `forge doctor` avisa se algum faltar.
- **Novidades de cada versão:** veja [Releases](releases.md).

## Se algo estranho aparecer

Rode `forge doctor` e envie a saída a quem mantém o projeto.
