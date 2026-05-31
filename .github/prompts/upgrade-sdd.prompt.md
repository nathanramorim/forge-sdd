---
description: "Migra a estrutura SDD do forge-sdd para uma versão mais nova"
mode: agent
---

Invoque o Migrator para migrar a estrutura Forge-SDD para a versão: ${input:versão alvo, ex: 1.2.0}

O Migrator deve: ler `sdd/.sdd-version`, aplicar o diff estrutural, atualizar `.sdd-version` e não tocar em conteúdo de domínio.
