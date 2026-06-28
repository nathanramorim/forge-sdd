---
description: "Instala e adapta uma skill a partir de uma URL externa"
mode: agent
---

Invoque o Specifier para importar a skill desta URL: ${input:URL da Skill no GitHub}

**Ações:**
1. Leia o conteúdo remoto (GitHub/Raw).
2. Adapte a instrução para o formato agnóstico Forge-SDD (válido para Copilot, Gemini e Claude).
3. Salve em \`sdd/skills/<nome>.chatmode.md\`.
4. Atualize \`sdd/skills/index.md\`.

**Regra:** Se a skill contiver ferramentas específicas (MCPs), documente-as na seção de requisitos da skill.
