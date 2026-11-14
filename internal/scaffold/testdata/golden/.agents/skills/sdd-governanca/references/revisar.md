# Comando: revisar

**Uso:** `/revisar [task]`.

1. **Camada A:** rode os critérios de aceitação da spec e `forge gate` (checks `P-xx`).
2. **Camada B:** audite `S-01` a `S-07` conforme `SKILL.md`, sobre o diff da task.
3. **Parecer:** produza o JSON do contrato em `references/veredito.md` e registre com `forge queue review <task-id> --verdict-file <arquivo>`. Devolva ao Líder para o painel humano.
4. **Telemetria:** `forge session record --feature "<feature ativa>" --outcome approved|rejected --criterio-atendido=true|false`.

Não modifique código de produção: apenas audite e reporte.
