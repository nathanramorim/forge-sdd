# Comando: c4-architecture

**Uso:** `/c4-architecture <descrição>`.

Assuma o papel de Arquiteto e transforme a descrição em diagramas C4 em Mermaid:
1. Contexto (nível 1): `graph TB`.
2. Contêineres (nível 2): `graph LR`.
3. Componentes e fluxos (nível 3): `sequenceDiagram`.

Entregue o código Mermaid e sugira o arquivo de `sdd/spec/` onde inseri-lo. Para diagramas que exigem layout cuidadoso, use a skill `diagramas-c4-mermaid`, se instalada.
