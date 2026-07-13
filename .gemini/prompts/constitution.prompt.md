# Prompt: constitution

**Uso:** Peça "/constitution" ou "alinhar arquitetura"

**Ação:**
Você deve atuar como o **Specifier** em uma sessão de alinhamento arquitetural.

0. **Idioma:** Antes de qualquer outra coisa, pergunte ao usuário (a) em qual idioma o chat deve interagir daqui em diante e (b) em qual idioma escrever commits e PRs (título e descrição). Registre a resposta em `sdd/memory/constitution.md`, seção Regras de Ouro, e siga essa escolha a partir desse ponto, inclusive nesta própria sessão.
1. **Discovery de Codebase:** Explore o diretório atual (`list_dir`, `read_file`) em busca de arquivos de configuração (ex: `go.mod`, `package.json`, `Cargo.toml`, `requirements.txt`), estrutura de pastas (`cmd/`, `internal/`, `src/`, `api/`) e padrões de código.
2. **Elaboração de Regras:** Identifique boas práticas implícitas no projeto (ex: "Usa Clean Architecture", "Testes em arquivos paralelos", "Config via .env").
3. **Povoamento:**
   - Preencha `sdd/memory/constitution.md` com a Missão, Stack Real e Regras de Ouro.
   - Preencha `sdd/spec/stack.md` com as dependências detectadas e o Layout do projeto.
   - Atualize `sdd/spec/overview.md` com diagramas **C4 Model (Mermaid)** refletindo a arquitetura detectada.

**Protocolo:**
- Se o projeto for novo (vazio), pergunte ao usuário sobre a visão arquitetural desejada.
- Garanta que as regras não conflitem com as diretrizes globais do Forge-SDD.
- Sempre reporte um resumo do que foi detectado antes de escrever nos arquivos.
