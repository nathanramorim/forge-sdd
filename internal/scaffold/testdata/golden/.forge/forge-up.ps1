# Abre as sessões forge-discovery, forge-builder e forge-reviewer do pipeline Forge SDD.
# Uso: .forge\forge-up.ps1 [--cli <agente>] [--discovery|--builder|--reviewer <agente>] [--dry-run]
# Agentes: claude, gemini, copilot ou codex
# A lógica fica em `forge up`; este script só garante o diretório e o binário.
$ErrorActionPreference = 'Stop'
Set-Location (Join-Path $PSScriptRoot '..')
if (-not (Get-Command forge -ErrorAction SilentlyContinue)) {
    Write-Error 'forge não está no PATH. Instale com: npm install -g @nathanramorim/forge-sdd@beta'
    exit 1
}
& forge up @args
exit $LASTEXITCODE
