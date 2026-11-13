# Fix 5ae2-08 — Nomenclatura, Ativação de Telemetria e Estimativa de Tokens

Esta especificação define os critérios executáveis para a correção dos seguintes bugs na metodologia Forge-SDD:
1. **Convenção de Nomenclatura configurável e retrocompatível:** Configuração `naming_convention` em `.sddrc`. Auto-detectada pelo `doctor` em projetos existentes.
2. **Ativação da Telemetria:** Orquestrador lê `.sddrc` e só grava se `enabled` for `true`.
3. **Métricas de Tokens:** Estimativa realista dos tokens no JSON de sessão (evitando `0`).
4. **Fluxo de Fix e comando `/novo-fix`:** Criação de branches `fix/` e arquivos `fix-` automáticos.

## Critérios de Aceitação Executáveis

1. **Configuração da Convenção no Init:**
   * `forge-sdd init` interativo deve incluir uma pergunta com opções: `sequencial`, `hash` e `workitem`.
   * A escolha deve ser persistida na chave `naming_convention` de `sdd/.sddrc`.

2. **Auto-Healing no Doctor para Projetos Existentes:**
   * O comando `forge-sdd doctor` deve atualizar a chave `naming_convention` se ela estiver ausente no `.sddrc`.
   * A convenção é inferida varrendo os arquivos existentes: se só houver sequenciais, usa `sequencial`; se só houver hashes, usa `hash`; caso contrário, cai para o padrão `sequencial`.
   * O `doctor` deve aceitar arquivos com prefixo `fix-` nas expressões regulares de classificação de nomes (`classifyNamingConvention`).

3. **Leitura e Respeito da Telemetria no Orquestrador:**
   * O prompt/skill do Orquestrador (`orquestrador.chatmode.md`) deve ler `sdd/.sddrc` no passo `READ-MIN`.
   * No passo `CLOSE`, a gravação em `sdd/.metrics/session-*.json` só deve ser feita se `telemetry.enabled` for `true` no `.sddrc`.

4. **Estimativa de Tokens nas Métricas:**
   * O Orquestrador deve estimar `tokens_input` e `tokens_output` (1 token ≈ 4 caracteres ou 0.75 palavras) ao gerar a métrica, evitando valores fixados em `0`.

5. **Comando e Fluxo `/novo-fix`:**
   * Deve ser introduzido o prompt `/novo-fix` (Claude, Gemini) para gerar branches `fix/<nome>` e especificações `fix-XX-<nome>.md` (ou `fix-<hash>-<nome>.md`).
   * Se o comando `/nova-feature` receber uma descrição de fix ou for acionado com `fix:`, ele deve adotar o mesmo fluxo.
