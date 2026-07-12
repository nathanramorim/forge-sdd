# Feature 5ae2-03 — `doctor` Detecta Deriva de Convenção de Nomenclatura

Evita que um projeto acumule inconsistência entre a convenção sequencial (`feat-NN`) e a convenção por hash (`feat-xxxx`), o mesmo tipo de descompasso encontrado no próprio repositório do Forge-SDD (PR #26 aberta vs. `feat-43a2`/`discovery-9b2f` já em uso).

## Critérios de Aceitação Executáveis

1. `forge-sdd doctor` deve escanear recursivamente `sdd/features/*.md` e `sdd/discovery/*.md`, classificando cada nome encontrado como convenção `sequencial` (`feat-NN`, dois dígitos) ou `hash` (`feat-[0-9a-f]{4}`).
2. Se as duas convenções coexistirem no mesmo projeto, `doctor` deve emitir um aviso explícito listando os arquivos de cada convenção em conflito.
3. Teste unitário cobrindo: projeto só-sequencial (sem aviso), projeto só-hash (sem aviso), projeto misto (aviso emitido com os arquivos corretos listados).

## Status: done

Implementado em `cmd/forge-sdd/doctor.go` (`classifyNamingConvention` + varredura de `sdd/features/` e `sdd/discovery/`). Rodar `forge-sdd doctor .` no próprio repositório do Forge-SDD confirma o cenário descoberto na Discovery: os 41 arquivos `feat-NN` sequenciais coexistem com os arquivos hash (`feat-43a2`, `feat-5ae2-*`, `discovery-5ae2-*`). Testes em `cmd/forge-sdd/commands_test.go` (`TestDoctorCommand_NamingConventionConsistent`, `TestDoctorCommand_NamingConventionDriftDetected`).
