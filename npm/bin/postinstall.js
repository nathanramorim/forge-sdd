#!/usr/bin/env node
'use strict';

const isColorSupported = !process.env.NO_COLOR && (process.stdout.isTTY || process.env.FORCE_COLOR);

const colors = {
  reset: isColorSupported ? '\x1b[0m' : '',
  bold: isColorSupported ? '\x1b[1m' : '',
  green: isColorSupported ? '\x1b[32m' : '',
  cyan: isColorSupported ? '\x1b[36m' : '',
  yellow: isColorSupported ? '\x1b[33m' : '',
  dim: isColorSupported ? '\x1b[90m' : '',
};

const width = 74;

function printLine(content, color = '') {
  // Remove escapes ANSI para calcular o comprimento visível
  const cleanLength = content.replace(/\x1b\[\d+m/g, '').length;
  
  // Calcula o espaço restante
  const spacesCount = Math.max(0, width - cleanLength - 1);
  console.log(`${colors.cyan}│${colors.reset} ${color}${content}${colors.reset}${' '.repeat(spacesCount)}${colors.cyan}│${colors.reset}`);
}

function printEmptyLine() {
  console.log(`${colors.cyan}│${colors.reset}${' '.repeat(width)}${colors.cyan}│${colors.reset}`);
}

function printSeparator() {
  console.log(`${colors.cyan}├${'─'.repeat(width)}┤${colors.reset}`);
}

function printBorderTop() {
  console.log(`${colors.cyan}┌${'─'.repeat(width)}┐${colors.reset}`);
}

function printBorderBottom() {
  console.log(`${colors.cyan}└${'─'.repeat(width)}┘${colors.reset}`);
}

console.log('');
printBorderTop();
printEmptyLine();
printLine(`🚀  ${colors.bold}${colors.cyan}Bem-vindo ao Forge-SDD!${colors.reset}`);
printEmptyLine();
printLine(`O CLI definitivo para scaffolding de Software Design Doc (SDD)`);
printLine(`pronto para uso com múltiplos agentes de Inteligência Artificial.`);
printEmptyLine();
printSeparator();
printEmptyLine();
printLine(`⚡  ${colors.bold}${colors.green}Comando global ativo: forge${colors.reset}`);
printEmptyLine();
printLine(`💡  ${colors.bold}Guia de Comandos Disponíveis:${colors.reset}`);
printEmptyLine();
printLine(`   ${colors.cyan}•${colors.reset} ${colors.bold}forge init${colors.reset}     ${colors.dim}─${colors.reset} Inicializa o SDD interativamente no projeto`);
printLine(`   ${colors.cyan}•${colors.reset} ${colors.bold}forge update${colors.reset}   ${colors.dim}─${colors.reset} Adiciona novos agentes ou atualiza a versão`);
printLine(`   ${colors.cyan}•${colors.reset} ${colors.bold}forge doctor${colors.reset}   ${colors.dim}─${colors.reset} Verifica a integridade da estrutura local`);
printLine(`   ${colors.cyan}•${colors.reset} ${colors.bold}forge destroy${colors.reset}  ${colors.dim}─${colors.reset} Remove a estrutura SDD local do projeto`);
printEmptyLine();
printSeparator();
printEmptyLine();
printLine(`👉  ${colors.bold}Como começar agora:${colors.reset}`);
printEmptyLine();
printLine(`   1. Navegue até o diretório do seu projeto`);
printLine(`   2. Execute o comando:`);
printLine(`      ${colors.bold}${colors.yellow}$ forge init${colors.reset}`);
printLine(`   3. Abra seu projeto com VS Code, Claude Code ou Gemini!`);
printEmptyLine();
printBorderBottom();
console.log('');
