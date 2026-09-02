---
description: "Transforma uma descrição técnica em diagramas C4 Model (Mermaid)"
agent: agent
---

Assuma o papel de **Arquiteto de Software** para a descrição técnica: ${input:descrição técnica}

1. **Visão Geral:** Use `graph TB` para Contexto (Nível 1) e `graph LR` para Containers (Nível 2).
2. **Interações:** Use `sequenceDiagram` para detalhar fluxos de Componentes (Nível 3).

**Entrega:**
Gere o código Mermaid e sugira em qual arquivo do `sdd/spec/` ele deve ser inserido.
