# System Instructions — meu-projeto

Você é um agente de desenvolvimento para o projeto **meu-projeto**.

## Projeto
- **Stack:** go
- **Banco:** none
- **Idioma dos templates:** pt-BR

## Metodologia
Este projeto usa **Forge-SDD** — uma metodologia de desenvolvimento assistido por IA com memória persistente.

### Arquivos essenciais
| Arquivo | Função |
|---------|--------|
| `sdd/memory/progress.md` | Estado ativo — leia primeiro |
| `sdd/memory/constitution.md` | Regras imutáveis |
| `sdd/features/feat-XX.md` | Feature em andamento |
| `sdd/spec/overview.md` | Visão geral do sistema |

### Protocolo de sessão (Lifecycle)
1. **READ-MIN:** Leia `sdd/memory/progress.md` e a feature ativa.
2. **PLAN:** Identifique a próxima ação, reporte o plano e aguarde confirmação.
3. **ACT:** Execute no escopo definido (Builder para código, Specifier para specs, etc. - veja `.gemini/skills/`).
4. **WRITE:** Edite apenas arquivos sob sua responsabilidade na fase atual.
5. **CLOSE:** Se agir como Orquestrador, atualize `progress.md` e grave métricas ao concluir.

### Regras
- Nunca commite direto em `main`
- Uma branch por feature (`feat/<nome>`)
- Consulte `sdd/memory/constitution.md` antes de decisões arquiteturais
- Para libs externas, verifique a documentação com a versão exata antes de usar
- Mantenha `progress.md` abaixo de 1 KB (dispare Archivist se exceder).
