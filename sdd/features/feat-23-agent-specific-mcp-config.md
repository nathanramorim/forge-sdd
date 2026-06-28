# Feature: MCP e Habilidades Específicas por Agente

## Contexto
Atualmente, o scaffolding do SDD centraliza a configuração de MCP no arquivo `.vscode/mcp.json`. No entanto, Gemini, Claude e Copilot possuem formas diferentes de gerenciar MCPs e habilidades. A centralização em um arquivo do VS Code ignora as especificidades de instalação e uso de cada agente.

## Objetivos
- Separar a configuração de MCP e habilidades por agente.
- Garantir que cada agente tenha suas próprias instruções de configuração de MCP.
- Evitar que o Gemini dependa de arquivos do VS Code para funcionar corretamente.

## Critérios de Aceite
- [x] Criar templates específicos de MCP para cada agente (se aplicável).
- [x] Atualizar `internal/scaffold/scaffold.go` para distribuir as configurações de MCP corretamente.
- [x] Garantir que o `GEMINI.md` aponte para as configurações/instruções corretas do Gemini.
- [x] O comando `doctor` deve validar as configurações específicas de cada agente.

## Plano de Implementação
1. **Research:** Identificar onde o Gemini e o Claude preferem suas configurações de MCP.
2. **Strategy:** 
   - Mover a configuração de MCP do Gemini para dentro de `.gemini/`.
   - Adicionar orientações de configuração para o Claude Desktop.
   - Manter `.vscode/mcp.json` apenas para o Copilot/VS Code.
3. **Execution:**
   - Criar `internal/scaffold/templates/agents/gemini/.gemini/mcp.json.tmpl`.
   - Modificar `internal/scaffold/scaffold.go` para não forçar `.vscode/mcp.json` se o Copilot não estiver selecionado.
   - Atualizar templates de `doctor.prompt.md` para cada agente.
4. **Validation:**
   - Testar scaffolding apenas com Gemini.
   - Testar scaffolding apenas com Copilot.
   - Testar scaffolding com múltiplos agentes.
