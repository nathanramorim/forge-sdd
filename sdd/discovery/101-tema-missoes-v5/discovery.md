# Discovery 101 — Tema "Missões" para o lançamento da v5.0.0

## Problema
A v5.0.0 do Forge SDD saiu em 19/11/2026, a mesma semana em que todo mundo fala do lançamento do GTA VI. O Nathan quer aproveitar esse momento: "se o Forge SDD tivesse a temática idêntica do GTA VI, como seria", com o número da versão e um letreiro parecido com o do jogo.

## Decisão sobre o pedido (registrar)
- **Não fazer cópia nem letreiro parecido.** O logo, o numeral "VI", a tipografia do título, a arte oficial, os personagens, a trilha e a interface do jogo são marcas e obras da Rockstar/Take-Two. Imitar o letreiro para promover um produto é uso de marca de terceiros e pode causar denúncia, remoção de post e advertência no perfil e no repositório público.
- **Fazer uma homenagem ao clima,** que o público reconhece sem copiar marca: o tema "Missões".

## Proposta: "Forge SDD 5.0 — Edição Missões"
- **Conceito:** cada estação do Forge SDD é uma "missão". Concepção = "Missão 1: o plano"; Planejamento = "Missão 2: a rota"; Construção = "Missão 3: a obra"; Governança = "Missão 4: o vigia"; Operação = "Missão 5: missão cumprida". A promessa: autonomia com segurança, "a IA executa, o plano segura, você aprova".
- **Versão:** a v5.0.0 já existe (tag `v5.0.0`, 19/11/2026). O tema vira a "edição de lançamento" dela, sem mudar o número. Usar o "5" como destaque gráfico próprio (o "5" de 5.0), desenhado com a nossa tipografia, sem imitar o "VI".
- **Visual (homenagem, não cópia):** pôr do sol de costa tropical (rosa, laranja, magenta, azul-petróleo), neon, palmeiras, motel art déco com letreiro com palavras nossas, mosaico de painéis no estilo "tela de carregamento" com contorno grosso. Tipografia: a da marca do Forge SDD (ou uma condensada livre, escolhida por nós e diferente da do jogo), nunca uma fonte feita para imitar o logo.
- **Onde aplicar:** banner do README e do site `forge-sdd.org` (seção "novidades da 5.0"), página de notas da versão, card de compartilhamento (og-image), carrossel e reel "Minhas missões" (já em produção no hub-busize, PR #16), post de lançamento.
- **Texto de referência:** "GTA VI" só como assunto do momento, em texto comum, na legenda dos posts, com o aviso "conteúdo original, sem ligação com a Rockstar/Take-Two". Nunca no logo, no banner nem no README.

## Fora do escopo
Mudar o nome do produto, criar logo novo do Forge SDD, usar qualquer áudio, imagem ou fonte do jogo.

## Riscos
- Marca: mitigado pelas regras acima; revisão de propriedade intelectual em cada peça.
- Prazo: o interesse cai em 1 a 2 semanas; peças leves primeiro (banner, og-image, post).

## Dúvidas para o Nathan (Gate 1)
1. Aprova o tema "Missões" como homenagem, sem letreiro parecido?
2. Onde aplicar primeiro: README + site, ou só posts?
3. O tema fica só no lançamento (temporário) ou vira identidade da série 5.x?

## Decisões do Gate 1 (Nathan, 19/11/2026: "seguir recomendações")
1. Tema "Missões" aprovado como **homenagem ao clima**, sem letreiro, logo, "VI", personagens, trilha ou interface do jogo.
2. Aplicar primeiro no **README (banner), nas notas da versão 5.0.0 e nos posts**; o site `forge-sdd.org` e a imagem de compartilhamento vêm numa segunda etapa.
3. O tema vale como **edição de lançamento da série 5.0.x**; depois do lançamento, decidir se vira a identidade da série 5.x.
4. Enviar a decisão ao repositório como PR em rascunho.

## Segunda rodada de decisões (Nathan, 19/11/2026)
Contexto: o PR #113 publicou no README um banner e um design system "GTA VI", medidos no conceito de interface de Matías Rodríguez (Behance, galeria 224512205), sem conhecer esta discovery.
5. **A identidade segue esta discovery.** Do design system ficam a paleta do pôr do sol, o neon e as molduras. O nome é "Edição Missões", sem "GTA VI" escrito, sem letreiro parecido e com o "5" como destaque próprio. O que o #113 publicou fora disso é corrigido no `fix-101-01`.
6. **O tema é temporário.** Vale até segunda-feira, 23/11/2026, 0h de Brasília. A data fica em um parâmetro único e o site volta sozinho ao layout anterior (`feat-101-01`). Isso responde à dúvida 3 e substitui a decisão 3.
7. **O README volta por PR agendado**, aberto por workflow na virada e mesclado pelo Nathan, sem commit automático no `main` (`feat-101-03`).
8. **Agentes rebeldes nas fases.** Os mascotes dos agentes aparecem como rebeldes, com arminhas e porrete de brinquedo, nos cartões das estações, sem perder qualidade (`feat-101-02`). São personagens do próprio forge-sdd.
9. **O site entra agora**, como a segunda etapa prevista na decisão 2.

## Dono
Departamento de TI da Busize (modelo Hub Enterprise) para o produto; Marketing para as peças de divulgação.
