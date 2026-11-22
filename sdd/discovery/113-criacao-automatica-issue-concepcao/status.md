# Status — 113 Criação automática de issue/card na concepção

**Gate 1:** aprovado
- **Iniciativa:** `113-criacao-automatica-issue-concepcao`
- **Status:** split aprovado pelo dono em 22/11/2026 (6 features, 113-01..06); Reviewer conforme. Sem `queue.json` ainda.
- **Próximo passo:** o chat monta a fila e dispara o Builder pela onda 1, uma branch e um PR em draft por feature (sem `forge up`).
- **Ondas:** 1) 113-01 e 113-02 em paralelo; 113-03 (após 01 e 02). 2) 113-04 e 113-05 em paralelo (após 03); 113-06 (após 04 e 05).
- **Decidido:** 113-05 emite `spec.aberta` só como gancho aditivo (no-op) até a 109 mesclar (`internal/modulo` não está na `main`); 113-06 não toca `.claude/commands`; `forge queue card` com flags de registro.
