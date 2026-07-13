# Feature 43 — Atalho forge e Interface de Boas-vindas Pós-Instalação

Esta especificação define os requisitos para simplificar o uso do CLI, permitindo o acionamento global através do comando `forge`, além de criar uma interface de boas-vindas rica e instrutiva que é exibida imediatamente após a instalação do pacote NPM.

## Requisitos e Critérios de Aceitação

1. **Atalho de Execução Simples (`forge`):**
   * O pacote NPM do `forge-sdd` deve expor tanto o binário `forge-sdd` quanto o alias `forge` sob a chave `"bin"` em `package.json`.
   * Quando o pacote for instalado globalmente (`npm i -g @nathanramorim/forge-sdd`), ambos os comandos `forge` e `forge-sdd` devem estar disponíveis no PATH e apontar para o wrapper JavaScript do CLI (`npm/bin/run.js`).

2. **Interface Visual Pós-Instalação (`postinstall`):**
   * Adicionar um script de `postinstall` no `package.json`.
   * O script deve executar `npm/bin/postinstall.js`.
   * O design no terminal deve ser limpo, moderno, comunicativo e visualmente atraente (utilizando estilos de caixa, cores ANSI e emojis) para causar uma ótima primeira impressão.
   * O conteúdo exibido deve incluir:
     * Uma mensagem calorosa de boas-vindas ao ecossistema Forge-SDD.
     * Indicação clara de que o comando global `forge` está pronto para uso.
     * Um mini-guide com os comandos principais e suas descrições:
       * `forge init` — Cria uma nova estrutura SDD interativa
       * `forge update` — Adiciona agentes e atualiza a versão
       * `forge doctor` — Checa a integridade dos arquivos e diretórios
       * `forge destroy` — Remove a estrutura local do projeto
     * Sugestões de próximos passos amigáveis (ex: abrir o projeto na IDE com o Copilot, ler o progresso).

3. **Validação:**
   * Um teste simulado de instalação local (`npm link` ou empacotamento) deve demonstrar a renderização da interface e que o atalho foi associado corretamente.
