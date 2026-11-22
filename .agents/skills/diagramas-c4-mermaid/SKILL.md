---
name: diagramas-c4-mermaid
description: Cria diagramas de arquitetura C4 (contexto, container, componente, dinâmico, implantação). Pergunta ao agente qual motor usar — Mermaid (renderiza no GitHub) ou diagram-studio (SVG/PNG dark-mode com assinatura). Ative sempre que o pedido for diagrama, C4, desenho de arquitetura, fluxo ou proposta visual.
---

# Diagramas C4 — Mermaid ou diagram-studio

Regra central: **nunca entregue um diagrama que você não renderizou e olhou.** O C4 nativo do Mermaid não dá controle de layout e desenha setas por cima de caixas. Este método foi validado numa proposta real (nove diagramas) e resolve isso. O diagram-studio gera SVG/PNG dark-mode com assinatura a partir de JSON estruturado.

## 0. Escolha do motor (obrigatório antes de qualquer esboço)

Use `AskUserQuestion` com as opções abaixo antes de iniciar o diagrama:

**"Qual motor de renderização usar?"**

- A) Mermaid — blocos ` ```mermaid `, renderiza no GitHub/VS Code nativamente
- B) diagram-studio — SVG/PNG dark-mode com assinatura (nathanramorim/diagram-studio)

Se **A (Mermaid)**: siga as seções 1–7 abaixo.
Se **B (diagram-studio)**: vá direto para a seção 8.

---

## 1. Qual notação usar em cada nível

| Nível C4 | Notação | Por quê |
|---|---|---|
| 1 Contexto | `C4Context` nativo, com `UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")` e relações direcionais (`Rel_D`, `Rel_R`, `Rel_U`) | Poucos elementos, o nativo aguenta |
| 2 Container | `flowchart LR` com **fronteiras (subgraph)** e a convenção visual C4 | O nativo cruza setas; o fluxo deixa você controlar a ordem |
| 3 Componente | `flowchart LR`, uma cadeia linear com saídas em leque | Idem |
| Dinâmico (uma operação) | `sequenceDiagram` | É o formato natural |
| Implantação (onde roda) | `flowchart LR` com subgraphs aninhados, só as chamadas estruturais | Menos setas |
| Fases ou linha do tempo | `flowchart LR` de caixas encadeadas | Simples e legível |

## 2. Regras contra setas embaralhadas
1. **No máximo 9 nós e 12 setas por diagrama.** Passou disso, divida em dois diagramas (por exemplo, "quem vigia quem" e "como o alerta chega").
2. **Uma pergunta por diagrama.** Não misture "o que vigia o quê" com "para onde vai o alerta".
3. **Pares alinhados:** quando A vigia X, declare A e X na mesma posição de suas colunas, na mesma ordem. Setas ficam retas e paralelas.
4. **Cadeias lineares:** prefira `A --> B --> C --> D` da esquerda para a direita. Convergência (vários → um) e leque (um → vários) são bons; malha não.
5. **Fronteiras (subgraph) lado a lado**, com `direction TB` dentro de cada uma; setas cruzam a fronteira no máximo uma vez.
6. **Sem seta longa cruzando o diagrama.** Se uma seta precisa atravessar três caixas, troque de lugar os nós, ou tire a seta e descreva na legenda ("os logs de todas as plataformas vão ao armazém, não repetido aqui").
7. **Rótulo curto** na seta (uma ou duas palavras). Detalhe vai no texto ao lado.
8. **Ordem de declaração = ordem visual.** O Mermaid posiciona pelo que vem primeiro.
9. Sistemas externos e pessoas ficam nas pontas (início e fim da cadeia), nunca no meio.
10. Se o alerta de todo mundo vai para o mesmo destino, **converge para um nó** em vez de desenhar cada caminho até a pessoa.

## 3. Modelo visual C4 para flowchart
```
%%{init: {"flowchart": {"curve": "linear", "nodeSpacing": 40, "rankSpacing": 80, "htmlLabels": true}}}%%
flowchart LR
  classDef pessoa fill:#08427b,stroke:#052e56,color:#ffffff
  classDef cont fill:#438dd5,stroke:#2e6295,color:#ffffff
  classDef comp fill:#85bbf0,stroke:#5d82a8,color:#000000
  classDef ext fill:#999999,stroke:#6b6b6b,color:#ffffff
  subgraph fr["Nome da fronteira"]
    direction TB
    a["<b>Nome</b><br/>[tecnologia]<br/>o que faz"]:::cont
    b[("<b>Banco</b><br/>[Postgres]<br/>o que guarda")]:::cont
  end
  p(["<b>Pessoa</b><br/>[Pessoa]<br/>papel"]):::pessoa
  x["<b>Sistema externo</b><br/>[SaaS]"]:::ext
  a -->|rótulo curto| b
```
Todo elemento tem **nome em negrito, tecnologia entre colchetes e uma frase de responsabilidade**. Cilindro `[( )]` para banco, `([ ])` para pessoa.

## 4. Renderizar e conferir (obrigatório)
Um comando, sem `pip` nem instalação manual:
```
uv run .agents/skills/diagramas-c4-mermaid/scripts/renderizar-mermaid.py ARQUIVO.md --saida PASTA
```
Ele extrai todos os blocos ` ```mermaid `, instala o `mermaid-cli` uma vez em `~/.cache/mermaid-cli` (usa o Chrome do sistema, sem baixar outro), renderiza cada diagrama em PNG com até 3 tentativas, grava os fontes `.mmd` ao lado e devolve a lista OK ou FALHOU.

Depois **abra cada PNG** (a ferramenta de leitura de imagem mostra) e confira:
- [ ] Alguma seta passa por cima de uma caixa?
- [ ] Alguma seta cruza outra sem necessidade?
- [ ] Algum rótulo está sobreposto a uma caixa ou a outra seta?
- [ ] Mais de 9 nós ou 12 setas? Divida.
- [ ] Pessoa e sistema externo estão nas pontas?
- [ ] Dá para ler tudo sem ampliar?

Corrija e renderize de novo até tudo estar limpo. Renderizar é barato; entregar diagrama embolado não é.

## 5. Armadilhas conhecidas
- **"Timed out waiting for the WS endpoint"**: o Chrome demorou a abrir (comum no Mac); o script tenta de novo. Não é erro de sintaxe.
- **Aviso de Rosetta / arm64**: é só desempenho (Node x64 no Mac Apple Silicon). O render funciona.
- **Opção de largura**: nesta versão do mermaid-cli o tamanho é `-s` (escala), não `-w`.
- **Zsh**: um `rm m*.png` sem arquivos aborta o comando; use `bash -c` ou `setopt no_nomatch`.
- **Seta em C4 nativo passando por dentro de uma caixa**: não tem correção por configuração; mude para flowchart.

## 6. O que entregar
1. O markdown com os blocos ` ```mermaid ` (renderiza no GitHub, VS Code e Obsidian).
2. Uma pasta com os **PNG e os `.mmd`**, para quem abre em outro visualizador, e uma seção "Diagramas em imagem" no documento listando os arquivos.
3. Uma frase dizendo a convenção usada (C4 nativo só no Contexto; Container e Componente em fluxo com cores C4).

## 7. Cuidados de conteúdo (Opção A e B)
- Rótulos no idioma do leitor (português do Brasil, sem jargão desnecessário).
- **Nenhum segredo, chave, endereço interno ou dado pessoal** em rótulos.
- Use os nomes reais das peças do cliente; onde algo é hipótese, escreva "a confirmar".
- Um diagrama que afirma o que não foi verificado é pior que nenhum: marque proposta como proposta.

---

## 8. Motor B — diagram-studio

### 8.1. Verificar o plugin (sem instalar nada por conta própria)
1. Rode `forge-sdd skills destinos`. Se o plugin já estiver disponível no projeto aberto ou no hub, use-o de lá e **não instale nos demais**.
2. Enquanto esse subcomando não existir, confira `ls .agents/plugins/diagram-studio/scripts/diagrama.py 2>/dev/null || echo AUSENTE` (caminho único, relativo ao destino: projeto aberto ou hub).
3. Se AUSENTE, pare e oriente a pessoa a pedir a instalação restrita ao projeto (feat-98-02 do diagram-studio). Não instale por outro meio nem fora do projeto.

### 8.2. Assinatura e modelo
Siga o passo de assinatura do `SKILL.md` do plugin (`.agents/plugins/diagram-studio/`). Escolha o modelo: Fases, Personas, Fluxo ou Camadas.

### 8.3. Mapeamento C4 → diagram-studio

| Nível C4 | Elemento JSON |
|---|---|
| Pessoa | `caixa` com `estilo: "contorno"` ou `"pill"` |
| Sistema / Contêiner | `caixa` com `estilo: "normal"` ou `"destaque"` |
| Sistema externo | `caixa` com `estilo: "apagado"` ou `"tracejado"` |
| Fronteira / Boundary | `caixa` com `estilo: "grupo"` |
| Relacionamento | `seta` com `forma: "horizontal"` ou `"vertical"` |

### 8.4. JSON de referência
```json
{
  "kicker": "// C4 · nível",
  "titulo": "A mensagem em uma frase",
  "largura": 1280,
  "altura": 500,
  "elementos": [
    {"tipo": "caixa", "id": "a", "x": 56, "y": 56, "w": 240, "h": 80, "texto": "Nome", "sub": "tecnologia", "estilo": "normal"},
    {"tipo": "caixa", "id": "b", "x": 360, "y": 56, "w": 240, "h": 80, "texto": "Sistema Externo", "sub": "SaaS", "estilo": "apagado"},
    {"tipo": "seta", "de": "a.dir", "para": "b.esq", "forma": "horizontal", "rotulo": "usa"}
  ]
}
```

### 8.5. Plano, render e entrega
1. Mostre um plano em até 5 linhas (título, nós, arestas) e aguarde confirmação.
2. Escreva o JSON em `docs/diagramas/<nome>.json`.
3. Renderize: `uv run .agents/plugins/diagram-studio/scripts/diagrama.py docs/diagramas/<nome>.json --saida docs/diagramas/<nome>` (mesmo caminho do passo 8.1).
4. Abra o PNG e confira texto apertado, setas sobre caixas e desalinhamento; corrija e renderize de novo.
5. Entregue SVG para README/docs (`![nome](docs/diagramas/<nome>.svg)`), PNG para slides e chat, e uma frase sobre o que o diagrama mostra.
6. Aplique os cuidados da seção 7.
