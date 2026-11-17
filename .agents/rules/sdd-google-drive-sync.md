---
description: Sincroniza documentação executiva, releases, diagramas e ativos de marketing do Forge-SDD com a pasta do Google Drive (Projetos/forge-sdd), sem copiar código-fonte.
globs: ["sdd/releases/**", "sdd/spec/adr/**", "site/**"]
regras: [COD-05, GOV-01]
---

# Guia Prático — Sincronização com o Hub do Google Drive (`Projetos/forge-sdd`)

Caminho local sincronizado pelo Google Drive for Desktop:
`/Users/nathanramorim/Library/CloudStorage/GoogleDrive-nathan.ramorim@gmail.com/Meu Drive/Projetos/forge-sdd/`

## Regras de Sincronização
1. **Sem código-fonte no Drive:** Nunca copie arquivos `.go`, `node_modules`, pacotes compilados ou configurações sensíveis para o Google Drive. Apenas documentação em Markdown/HTML, diagramas Mermaid, imagens, vídeos e materiais de marketing.
2. **Ao publicar nova release ou fechar iniciativa (`/archive`, bump de versão em `sdd/releases/history.md`):**
   - Atualize `03-Releases-e-Roadmap/Historico-de-Atualizacoes-e-Roadmap.md` no Google Drive com os destaques em linguagem de produto e o estado do roadmap.
3. **Ao criar ou alterar comandos, sessões ou ADRs (`sdd/spec/adr/`):**
   - Atualize `01-Documentacao-Base-IA/Contexto-Mestre-para-IA.md` e `02-Arquitetura-e-Diagramas/Arquitetura-Pipeline-5-Estacoes-e-ADRs.md`.
4. **Ao atualizar a Landing Page (`site/`) ou criar peças de divulgação:**
   - Sincronize o copy em `04-Marketing-e-Landing-Page/Copy-Landing-Page-e-Posicionamento.md`, os ativos visuais em `04-Marketing-e-Landing-Page/assets/` e novos roteiros/posts em `04-Marketing-e-Landing-Page/roteiros-e-posts/`.
