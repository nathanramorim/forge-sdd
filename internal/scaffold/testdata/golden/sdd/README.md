# SDD — demo

Navegação rápida:

| Arquivo | Função |
|---------|--------|
| `discovery/<id>-<nome>/status.md` | Estado da iniciativa — leia primeiro |
| `memory/constitution.md` | Regras imutáveis |
| `spec/overview.md` | Visão geral do sistema |
| `discovery/` | Iniciativas: discovery, criteria, plan, status, fila, features e fixes |
| `plan.md` | Roadmap faseado |
| `skills/index.md` | Skills disponíveis |

> Tutoriais: `HOWTO.md` | Métricas: `.metrics/`

---

## 🚀 Guia de Início Rápido

Dependendo do estado do seu projeto, siga uma das trilhas lógicas:

### 🆕 Novo Projeto (Do Zero)
1. `/constitution` → Ative para definir a arquitetura e as regras iniciais.
2. `/discovery "[ideia]"` → **Discovery de Produto:** Use para estruturar o "que" e o "para quem" o sistema será construído.
3. `/nova-feature` → Crie as primeiras tarefas no roadmap baseado no discovery.
4. `/implementar` → Execute a fila da iniciativa.

### 🏗️ Projeto Existente (Adoção)
1. `/constitution` → O agente fará o **scan do seu codebase** para aprender suas regras.
2. `/status` → Veja o diagnóstico inicial.
3. `/nova-feature` → Mapeie dívidas técnicas ou novas evoluções.

### 💡 Comandos Universais
Use `/status`, `/doctor`, `/revisar` e `/archive` conforme necessário durante as sessões.

> Mais detalhes em: \`HOWTO.md\` | Metodologia: [forge-sdd.vercel.app](https://forge-sdd.vercel.app)


