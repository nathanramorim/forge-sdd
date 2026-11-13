# Especificação Técnica — Fase 40: Refinamento do Comando init e Relatório de Inicialização

Esta especificação define o refinamento do comportamento do comando `init` no CLI `forge-sdd` em relação à criação de pastas, extração automática do nome do projeto do diretório e geração de relatórios de conclusão.

## 1. Regras de Diretório e Criação de Pastas

O CLI deve determinar o diretório alvo e o nome padrão do projeto de acordo com a forma como o comando `init` é chamado:

### Caso A: `forge-sdd init .` (ou diretório especificado explicitamente como `.`)
- **Diretório Alvo:** O diretório atual onde o comando está sendo rodado.
- **Nome do Projeto:**
  - Se a flag `--name` for passada, usa o valor da flag.
  - Caso contrário, extrai automaticamente o nome da pasta física atual (ex: se o caminho atual for `/Users/user/git/meu-app`, o nome padrão será `meu-app`).
- **Comportamento:** A inicialização é feita diretamente na pasta atual. Nenhuma pasta filha é criada para o projeto.

### Caso B: `forge-sdd init meu-projeto` (diretório especificado explicitamente com outro nome)
- **Diretório Alvo:** A pasta `./meu-projeto/`.
- **Nome do Projeto:**
  - Se a flag `--name` for passada, usa o valor da flag.
  - Caso contrário, o nome do projeto será `meu-projeto`.
- **Comportamento:** O CLI cria a pasta `./meu-projeto/` no diretório atual e inicializa toda a metodologia dentro dela.

### Caso C: `forge-sdd init` (chamado sem argumentos)
- O CLI abre o formulário interativo.
- O usuário insere o nome do projeto no formulário (ex: `novo-servico`).
- **Diretório Alvo:** A pasta `./novo-servico/` (criada no diretório atual com o nome do projeto inserido).
- **Nome do Projeto:** `novo-servico`.
- **Comportamento:** O CLI cria a pasta `./novo-servico/` e gera toda a estrutura do Forge-SDD dentro dela.

---

## 2. Relatório de Conclusão do CLI

Tanto no comando `init` quanto no `update` (ou quando redirecionado do `init` para o `update` em projetos existentes), o CLI deve apresentar no terminal um relatório resumido e visual das configurações aplicadas.

### Conteúdo do Relatório:
```
✓ Estrutura Forge-SDD inicializada/atualizada com sucesso!

Resumo do Projeto:
- Nome: [Nome do Projeto]
- Diretório: [Diretório de Destino]
- Stack: [Stack]
- Banco de Dados: [Banco de Dados]
- Idioma: [Idioma]
- Agentes Configurados: [gemini, claude, etc.]
- Telemetria: [Ativo/Inativo]
```
- Esse relatório ajuda o usuário a ver imediatamente as opções ativas e o nome oficial que foi atribuído ao projeto.
