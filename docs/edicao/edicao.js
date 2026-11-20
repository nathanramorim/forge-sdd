/* Edição temporária do site. Este é o único lugar com o identificador e a data de fim. */
(function () {
  var RFC3339 = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/;
  var edicao = {
    id: "missoes",
    fim: "2026-11-23T00:00:00-03:00",
    vigente: function (agoraMs) {
      if (typeof this.fim !== "string" || !RFC3339.test(this.fim)) return false;
      var fim = Date.parse(this.fim);
      return typeof agoraMs === "number" && isFinite(agoraMs) && isFinite(fim) && agoraMs < fim;
    }
  };
  window.forgeEdicao = edicao;

  var forcado = /[?&]edicao=(on|off)(?:&|$)/.exec((window.location && window.location.search) || (typeof location !== "undefined" && location.search) || "");
  var ativa = forcado ? forcado[1] === "on" : edicao.vigente(Date.now());
  if (!ativa) return;

  document.documentElement.setAttribute("data-edicao", edicao.id);
  if (!document.addEventListener) return;

  /* Estação de cada cartão da LP, na ordem em que a página os desenha. */
  var estacaoDoCartao = ["Operação", "Concepção", "Planejamento", "Construção", "Governança"];

  function criar(tag, classe, texto) {
    var el = document.createElement(tag);
    el.className = classe;
    el.textContent = texto;
    return el;
  }

  function rotularCartoes(missoes) {
    var cartoes = document.querySelectorAll("#estacoes .agent-card");
    for (var i = 0; i < cartoes.length; i++) {
      var meta = cartoes[i].querySelector(".agent-meta");
      var estacao = estacaoDoCartao[Number(cartoes[i].dataset.idx)];
      if (!meta || !estacao || meta.querySelector(".edicao-missao")) continue;
      for (var k = 0; k < missoes.length; k++) {
        var item = missoes[k];
        if (item && typeof item.estacao === "string" && typeof item.titulo === "string" && item.estacao === estacao) {
          meta.insertBefore(criar("div", "edicao-missao", item.titulo), meta.querySelector("h3"));
        }
      }
    }
  }

  function tirarMascotes3D() {
    var palcos = document.querySelectorAll(".avatar-stage canvas");
    for (var i = 0; i < palcos.length; i++) palcos[i].parentNode.removeChild(palcos[i]);
  }

  function vestir(dados) {
    if (!dados || typeof dados.edicao !== "string" || typeof dados.promessa !== "string" || !Array.isArray(dados.missoes)) return;

    var marca = document.querySelector(".site-header .brand, .wiki-brand");
    if (marca && !marca.querySelector(".edicao-selo")) {
      marca.appendChild(criar("span", "edicao-selo", dados.edicao));
    }

    var hero = document.querySelector(".hero");
    if (hero && !hero.querySelector(".edicao-faixa")) {
      var faixa = criar("div", "edicao-faixa", "");
      faixa.appendChild(criar("span", "edicao-cinco", "5.0"));
      faixa.appendChild(criar("span", "edicao-nome", dados.edicao));
      faixa.appendChild(criar("span", "edicao-promessa", dados.promessa));
      hero.insertBefore(faixa, hero.firstChild);
      var ultima = dados.missoes[dados.missoes.length - 1];
      if (ultima && typeof ultima.titulo === "string") {
        var texto = ultima.titulo.split(": ").pop();
        faixa.parentNode.insertBefore(criar("div", "edicao-placa", texto), faixa.nextSibling);
      }
    }

    var estacoes = document.getElementById("estacoes");
    if (!estacoes) return;
    rotularCartoes(dados.missoes);
    tirarMascotes3D();
    if (window.MutationObserver) {
      new MutationObserver(function () { rotularCartoes(dados.missoes); tirarMascotes3D(); }).observe(estacoes, { childList: true, subtree: true });
    }
  }

  document.addEventListener("DOMContentLoaded", function () {
    var caminho = (typeof location !== "undefined" && /\/wiki\/$/.test(location.pathname || "")) ? "../design/edicao-missoes/missoes.json" : "design/edicao-missoes/missoes.json";
    fetch(caminho)
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (dados) { vestir(dados); })
      .catch(function () {});
  });
})();
