# Telemetria e Métricas no Forge-SDD

Este documento explica como funciona o sistema de telemetria e coleta de métricas do Forge-SDD, seus objetivos, dados coletados, controle de privacidade e como habilitá-la ou desativá-la.

---

## 1. O que é e por que existe?

O Forge-SDD inclui um mecanismo de telemetria local e assíncrona com o objetivo exclusivo de **mensurar a eficiência das regras de IA** e **detectar loops ou falhas frequentes de templates de prompts**.

As métricas coletadas ajudam a refinar o framework para responder questões como:
- *O template do Specifier gerou loops de entendimento na IA?*
- *Quanto tempo os agentes levam para transicionar sessões de trabalho de PM para Engenheiro?*
- *Quais as stacks e agentes mais adotados pela comunidade?*

---

## 2. O que é coletado? (Apenas Metadados Técnicos)

A telemetria foca estritamente em metadados de execução técnica e uso.

### Dados Coletados:
- **CLI logs:** Comandos executados (`init`, `update`, `doctor`) e status (sucesso/falha).
- **Métricas de Sessão (IA):** Tempo de sessão ativa de trabalho, contagem de arquivos editados e quantidade de chamadas de ferramentas.
- **Configuração Geral:** Idioma dos templates (ex: `pt-BR`), stack (ex: `go`), banco de dados (ex: `postgres`) e lista de agentes ativos.

---

## 3. Segurança e Privacidade (O que NÃO é coletado)

Para garantir a segurança do seu código comercial e dados pessoais, o Forge-SDD **NUNCA** acessa, armazena ou transmite:
- ❌ Código-fonte do seu projeto.
- ❌ Conteúdo de arquivos de especificações (`sdd/features/`).
- ❌ Chaves de APIs, tokens ou credenciais.
- ❌ Nomes de usuário do sistema ou dados pessoais de identificação.
- ❌ Caminhos absolutos de arquivos que revelem a estrutura do seu computador local.

---

## 4. Gerenciamento e Configuração

A telemetria é controlada por meio do arquivo de configuração do projeto em `sdd/.sddrc`.

### Como desabilitar em novos projetos:
Ao inicializar um projeto, você pode pular a telemetria passando a flag correspondente:
```bash
npx @nathanramorim/forge-sdd init --no-telemetry
```

### Como habilitar ou desativar em projetos existentes:
Você pode editar diretamente o bloco `"telemetry"` no arquivo [sdd/.sddrc](file:///Users/nathanramorim/git/forge-sdd/sdd/.sddrc):

```json
{
  "telemetry": {
    "enabled": true,       // true para ativar, false para desativar
    "anonymous": true,      // envia métricas sem nenhum ID de rastreamento
    "endpoint": "local"     // "local" para manter dados estritamente em disco no diretório
  }
}
```

### Como atualizar projetos antigos:
Se o seu projeto foi criado em versões antigas do CLI que não continham telemetria, você pode injetar as chaves padrão atualizando o projeto com o CLI:
```bash
npx @nathanramorim/forge-sdd@latest update --upgrade
```
