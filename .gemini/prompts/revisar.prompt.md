# Prompt: revisar

**Uso:** Peça "/revisar" ou "valide a feature"

**Ação:**
1. Acione a lógica de **Revisor**.
2. **Guardrail (Review):** Valide se apenas os arquivos declarados em "Arquivos gerados" na feature ativa foram modificados (\`git status\` ou \`git diff --name-only\`).
3. Valide a feature atual.
4. **Gravação de Métricas:** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "<feature ativa>" --outcome approved|rejected --criterio-atendido=true|false` — garante telemetria mesmo quando a sessão não chega a `/proxima-feature`.
