# Lições Aprendidas — SEG-01

Lições associadas à rule `SEG-01` da Constituição (mais recentes primeiro).

- script de deploy usou CLOUDFLARE_ACCOUNT_ID e XDG_CONFIG_HOME do shell e publicou a LP na conta errada, duas vezes → a conta é entrada explícita e validada (--conta, variável própria ou arquivo local fora do git), conferida contra o login e passada só ao comando de deploy (PR #284) (sdd/discovery/118-upgrade-completo-da-beta/status.md)
- criar a issue externa antes de ler a evidência local, duplicando a issue a cada reexecução → ler a evidência primeiro (fast path sem chamada externa) e só depois buscar e criar; gravar a evidência de forma atômica (feat-113-03, #238)
