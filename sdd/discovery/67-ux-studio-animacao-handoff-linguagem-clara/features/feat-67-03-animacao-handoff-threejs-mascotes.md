# Tarefa feat-67-03 — Animação de Handoff entre os Mascotes 3D Three.js

## 1. Contexto e Objetivo
Prover feedback visual imersivo e compreensível para qualquer usuário durante as transições de tarefas entre agentes. Quando uma estação conclui sua atividade e realiza o handoff para a próxima estação (ex: Líder para Concepção, Concepção para Builder, ou Builder para Governança), os mascotes 3D Three.js devem executar uma animação de diálogo/passagem de tarefa:
- Os mascotes de origem e destino reagem mutuamente com uma leve rotação de cabeça/olhar.
- Um arco de luz/partícula viaja dinamicamente da estação de origem para a de destino.
- Um balão/tooltip sutil de passagem de bastão ("Enviando tarefa...", "Recebendo tarefa...") é exibido temporariamente.

## 2. O que já existe (Reuse Check — DUP-02)
- Controladores `MascotController` e cena Three.js embutida em `internal/studio/assets/studio.js`.
- Trilho e eventos de handoff em `applySnapshot` e `sendHandoff`.

## 3. Critérios de Aceitação (Gherkin)
### Cenário 1: Disparo da Animação de Handoff
- **Dado** que ocorre uma transferência de artefato entre duas estações
- **Quando** o evento for processado no frontend
- **Então** os mascotes envolvidos devem executar a animação de handoff visual
- **E** a animação deve terminar de forma suave sem travar a taxa de quadros (60 FPS) nem vazar memória WebGL.

## 4. Arquivos Afetados
- `internal/studio/assets/studio.js`
- `internal/studio/assets/studio.css`
- `internal/studio/assets_test.go`
