# Fluxos — forge-sdd

## Fluxo principal: `forge-sdd init [dir]`

```
forge-sdd init [targetDir] [flags]
  │
  ├─ parse flags (cobra)
  ├─ se --dry-run → config.DryRun = true
  │
  ├─ se --yes OU DryRun:
  │     config = Config.Defaults() + flags
  │   senão:
  │     config = survey.Run()   ← formulário huh interativo
  │
  ├─ scaffold.Run(config, targetDir)
  │   ├─ walk templates/ (embed.FS) → []filePath
  │   ├─ para cada filePath:
  │   │   ├─ render: text/template.Execute(filePath, config)
  │   │   ├─ se DryRun:  fmt.Println("[DRY]", destPath)
  │   │   └─ senão:      criar dirs + escrever arquivo
  │   ├─ se arquivo existir: coletar conflito, não sobrescrever
  │   └─ retornar ([]criados, []conflitos, error)
  │
  ├─ se conflitos → imprimir lista + Exit 1
  └─ imprimir "✓ Estrutura Forge-SDD criada."
     imprimir próximos passos
```

## Fluxo dry-run
Idêntico ao principal, mas `scaffold.Run` imprime a árvore sem criar arquivos. Conflitos de arquivo existente não são erros no dry-run — apenas avisam.

## Tratamento de erros

| Situação | Comportamento |
|----------|---------------|
| Diretório alvo não existe | Criar recursivamente |
| Arquivo já existe no alvo | Coletar em lista de conflitos; não sobrescrever; Exit 1 ao final |
| Template inválido (parse error) | Erro fatal com nome do template + número de linha |
| Survey cancelado (Ctrl+C) | Exit 0 sem criar nada |
