# ADR-0017: Servidor MCP Somente Leitura com Saída Saneada

- **Data:** 2026-11-20
- **Status:** `aceito` (20/11/2026, aprovado pelo dono no chat)
- **Regras afetadas (IDs da constitution):** `ARQ-06` (subcomando `mcp`), `SEG-05`, `SEG-06`, `SEG-08` (criado em 20/11/2026), `FLX-08`, `GOV-01` (esta ADR). O teste dos IDs passou de 45 para 46.
- **Autor / Responsável:** Concepção; aceite do `mcp` em `ARQ-06` dado pelo dono no chat em 20/11/2026

## Relação com ADRs vigentes
- **ADR-0015: detalha** D-5 (diagnóstico exposto por MCP). **ADR-0008 e a skill `sdd-operacao`: convivem**; `forge acp` (IDEs) é outro protocolo e não muda.

## Contexto
O diagnóstico `forge doctor --harness` precisa estar ao alcance do agente em várias superfícies. Uma ferramenta MCP que escrevesse no projeto, ou devolvesse texto não tratado ao modelo, abriria um canal de injeção.

## Opções Consideradas
1. **A. MCP mínimo, só leitura**, com uma ferramenta.
2. **B. MCP com ferramentas de escrita** (despacho, fila).
3. **C. Só hook `SessionStart` e instrução, sem MCP.**

## Decisão
- **D-1:** adotar **A**: `forge mcp serve` (`cmd/forge-sdd/mcp.go`, `internal/acp/mcp.go`) fala JSON-RPC por stdio e expõe só `forge_doctor`, sem argumentos (`CodeInvalidParams` caso receba).
- **D-2:** a saída passa por `mcpSaneia` (remove controles, ESC e caracteres de formatação ou bidi; limita o tamanho); o conteúdo é dado (`FLX-08`).
- **D-3:** o hook `SessionStart` e a instrução nos 3 adaptadores cobrem onde não há MCP.

## Consequências
- **Positivas:** superfície mínima e sem efeito colateral. **Negativas:** o agente não aciona o pipeline por MCP; novas ferramentas exigem nova ADR.
- **Ações de acompanhamento:** `ARQ-06` já lista `mcp`; não verificado: hooks no app desktop.
