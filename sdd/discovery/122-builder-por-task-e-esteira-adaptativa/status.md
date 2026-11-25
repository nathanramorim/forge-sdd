# Status — 122 Builder por task e esteira adaptativa

**Gate 1:** aprovado pelo Nathan em 24/11/2026 23h ("vamos rodar o split"). Alvo: 6.0.0-beta.
**Status:** split em 8 features (`features/feat-122-01..08`); implementação em 3 trilhas paralelas.
**Próximo passo:** integrar as trilhas na branch e rodar build, vet e testes.
**Decisões (24/11, dono):** `task` padrão na beta; sem teto fixo, limite por disco 80% e memória com reserva de 2 GB; janela 85%; squash no PR da feature; publicação no npm só com ok do dono.
**A confirmar:** medição real (3.3) depois da beta.
