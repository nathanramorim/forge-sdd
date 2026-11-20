/* Acessórios 3D dos agentes rebeldes para a Edição Missões (geometria pura Three.js, estilo brinquedo). */
(function () {
  var edicao = window.forgeEdicao || (window.forgeEdicao = {});

  edicao.acessoriosPorMarca = {
    ring: ["panama", "charuto", "corrente", "arminha"],
    diamond: ["oculos", "colarinho", "brinco", "arminha"],
    bar: ["bonereto", "corrente", "porrete"],
    wave: ["bandana", "oculos", "brinco", "arminha"],
    tri: ["bone", "corrente", "porrete"]
  };

  var COR = {
    moldura: 0x151515,
    botao: 0x232123,
    pessego: 0xffaa6e,
    coral: 0xfe818d,
    rosa: 0xf061af,
    foco: 0xfe7694,
    ambar: 0xe28642,
    aviso: 0xe6d364,
    violeta: 0x824cc8,
    indigo: 0x4053c6,
    ouro: 0xf2b84b,
    charuto: 0x7a4a2b,
    brasa: 0xff6a3d
  };

  function mat(hex, roughness, metalness) {
    return new window.THREE.MeshStandardMaterial({
      color: hex,
      roughness: roughness === undefined ? 0.32 : roughness,
      metalness: metalness === undefined ? 0.14 : metalness
    });
  }

  function criarBandana(THREE, marca) {
    var grupo = new THREE.Group();
    var corFaixa = marca === "ring" ? COR.rosa : COR.violeta;
    var faixa = new THREE.Mesh(new THREE.TorusGeometry(0.495, 0.048, 18, 36), mat(corFaixa, 0.42, 0.08));
    faixa.rotation.x = Math.PI / 2;
    faixa.scale.set(1.0, 0.88, 1.0);
    faixa.position.y = 0.49;
    grupo.add(faixa);

    var friso = new THREE.Mesh(new THREE.TorusGeometry(0.502, 0.015, 16, 36), mat(COR.pessego, 0.35, 0.12));
    friso.rotation.x = Math.PI / 2;
    friso.scale.set(1.0, 0.88, 1.0);
    friso.position.y = 0.515;
    grupo.add(friso);

    var no = new THREE.Mesh(new THREE.SphereGeometry(0.065, 16, 16), mat(corFaixa, 0.4, 0.08));
    no.position.set(-0.48, 0.49, -0.06);
    grupo.add(no);

    var pontaA = new THREE.Mesh(new THREE.CylinderGeometry(0.018, 0.038, 0.17, 16), mat(COR.pessego, 0.4, 0.08));
    pontaA.position.set(-0.56, 0.43, -0.10);
    pontaA.rotation.z = 0.55;
    pontaA.rotation.x = -0.25;
    grupo.add(pontaA);

    var pontaB = new THREE.Mesh(new THREE.CylinderGeometry(0.016, 0.032, 0.14, 16), mat(corFaixa, 0.4, 0.08));
    pontaB.position.set(-0.55, 0.54, -0.11);
    pontaB.rotation.z = 1.15;
    grupo.add(pontaB);

    return grupo;
  }

  function criarOculos(THREE, marca) {
    var grupo = new THREE.Group();
    grupo.position.set(0, -0.10, 0.725);

    var corAro = marca === "diamond" ? COR.pessego : COR.rosa;
    var mAro = mat(corAro, 0.25, 0.22);
    var mLente = new THREE.MeshStandardMaterial({
      color: COR.moldura,
      roughness: 0.18,
      metalness: 0.2,
      transparent: true,
      opacity: 0.38,
      depthWrite: false
    });

    [-0.21, 0.21].forEach(function (x) {
      var aro = new THREE.Mesh(new THREE.TorusGeometry(0.148, 0.024, 16, 28), mAro);
      aro.scale.set(1.16, 0.92, 1.0);
      aro.position.set(x, 0, 0);
      grupo.add(aro);

      var lente = new THREE.Mesh(new THREE.CircleGeometry(0.136, 24), mLente);
      lente.scale.set(1.14, 0.9, 1.0);
      lente.position.set(x, 0, -0.008);
      grupo.add(lente);
    });

    var ponte = new THREE.Mesh(new THREE.CylinderGeometry(0.018, 0.018, 0.11, 16), mAro);
    ponte.rotation.z = Math.PI / 2;
    ponte.position.set(0, 0.02, 0.004);
    grupo.add(ponte);

    [-0.38, 0.38].forEach(function (x) {
      var lado = x < 0 ? -1 : 1;
      var haste = new THREE.Mesh(new THREE.CylinderGeometry(0.016, 0.016, 0.24, 16), mAro);
      haste.rotation.x = Math.PI / 2;
      haste.rotation.z = lado * -0.14;
      haste.position.set(x, 0.02, -0.12);
      grupo.add(haste);
    });

    return grupo;
  }

  function criarBone(THREE, reto) {
    var grupo = new THREE.Group();
    grupo.position.set(0, 0.48, 0);

    var copa = new THREE.Mesh(
      new THREE.SphereGeometry(0.47, 24, 16, 0, Math.PI * 2, 0, Math.PI * 0.44),
      mat(COR.botao, 0.42, 0.1)
    );
    copa.scale.set(1.02, 0.85, 0.9);
    copa.position.y = 0.01;
    grupo.add(copa);

    var faixa = new THREE.Mesh(new THREE.TorusGeometry(0.455, 0.026, 16, 32), mat(COR.coral, 0.32, 0.14));
    faixa.rotation.x = Math.PI / 2;
    faixa.scale.set(1.02, 0.9, 1.0);
    faixa.position.y = 0.03;
    grupo.add(faixa);

    var aba = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.32, reto ? 0.06 : 0.032, 24), mat(reto ? COR.pessego : COR.rosa, 0.34, 0.12));
    aba.scale.set(1.08, 1.0, reto ? 0.95 : 0.68);
    aba.position.set(0, reto ? 0.0 : 0.025, reto ? 0.5 : 0.35);
    aba.rotation.x = reto ? 0.18 : -0.14;
    grupo.add(aba);

    return grupo;
  }

  function criarPanama(THREE) {
    var grupo = new THREE.Group();
    grupo.position.set(0, 0.5, 0);

    var aba = new THREE.Mesh(new THREE.CylinderGeometry(0.64, 0.64, 0.03, 36), mat(COR.pessego, 0.55, 0.05));
    aba.scale.set(1.0, 1.0, 0.92);
    aba.rotation.z = 0.06;
    grupo.add(aba);

    var copa = new THREE.Mesh(new THREE.CylinderGeometry(0.30, 0.38, 0.27, 30), mat(COR.pessego, 0.55, 0.05));
    copa.position.y = 0.14;
    grupo.add(copa);

    var topo = new THREE.Mesh(new THREE.CircleGeometry(0.30, 30), mat(COR.pessego, 0.55, 0.05));
    topo.rotation.x = -Math.PI / 2;
    topo.position.y = 0.276;
    grupo.add(topo);

    var faixa = new THREE.Mesh(new THREE.CylinderGeometry(0.385, 0.395, 0.07, 30, 1, true), mat(COR.rosa, 0.4, 0.1));
    faixa.position.y = 0.045;
    grupo.add(faixa);

    return grupo;
  }

  function criarCharuto(THREE) {
    var grupo = new THREE.Group();
    grupo.position.set(-0.2, -0.47, 0.6);
    grupo.rotation.set(0.05, -0.5, 0.18);

    var corpo = new THREE.Mesh(new THREE.CylinderGeometry(0.042, 0.046, 0.4, 16), mat(COR.charuto, 0.7, 0.02));
    corpo.rotation.x = Math.PI / 2;
    corpo.position.z = 0.16;
    grupo.add(corpo);

    var anel = new THREE.Mesh(new THREE.TorusGeometry(0.047, 0.011, 12, 20), mat(COR.ouro, 0.3, 0.5));
    anel.position.z = 0.08;
    grupo.add(anel);

    var brasa = new THREE.Mesh(
      new THREE.SphereGeometry(0.046, 14, 12),
      new THREE.MeshStandardMaterial({ color: COR.brasa, emissive: COR.brasa, emissiveIntensity: 0.9, roughness: 0.6 })
    );
    brasa.position.z = 0.37;
    grupo.add(brasa);

    return grupo;
  }

  function criarCorrente(THREE) {
    var grupo = new THREE.Group();
    grupo.position.set(0, -0.72, 0);

    var elo = mat(COR.ouro, 0.22, 0.75);
    var colar = new THREE.Mesh(new THREE.TorusGeometry(0.36, 0.03, 14, 40), elo);
    colar.rotation.x = Math.PI / 2;
    grupo.add(colar);

    var haste = new THREE.Mesh(new THREE.CylinderGeometry(0.012, 0.012, 0.1, 10), elo);
    haste.position.set(0, -0.05, 0.36);
    grupo.add(haste);

    var pingente = new THREE.Mesh(new THREE.OctahedronGeometry(0.075, 0), elo);
    pingente.scale.set(0.85, 1.25, 0.5);
    pingente.position.set(0, -0.14, 0.36);
    grupo.add(pingente);

    return grupo;
  }

  function criarBrinco(THREE) {
    var grupo = new THREE.Group();
    grupo.position.set(-0.8, -0.2, 0.02);

    var argola = new THREE.Mesh(new THREE.TorusGeometry(0.07, 0.014, 12, 24), mat(COR.ouro, 0.22, 0.75));
    argola.rotation.y = Math.PI / 2;
    grupo.add(argola);

    return grupo;
  }

  function criarColarinho(THREE) {
    var grupo = new THREE.Group();
    grupo.position.set(0, -0.72, 0);

    var gola = new THREE.Mesh(new THREE.TorusGeometry(0.37, 0.07, 14, 40), mat(COR.coral, 0.5, 0.05));
    gola.rotation.x = Math.PI / 2;
    grupo.add(gola);

    var cores = [COR.aviso, COR.rosa, COR.pessego, COR.violeta];
    for (var i = 0; i < 10; i++) {
      var a = (i / 10) * Math.PI * 2;
      var flor = new THREE.Mesh(new THREE.SphereGeometry(0.036, 12, 10), mat(cores[i % cores.length], 0.4, 0.05));
      flor.position.set(Math.cos(a) * 0.37, 0.05, Math.sin(a) * 0.37);
      grupo.add(flor);
    }

    return grupo;
  }

  function criarArminha(THREE, marca) {
    var grupo = new THREE.Group();
    grupo.position.set(0.64, -0.02, 0.38);
    grupo.rotation.set(0.14, 0.74, 0.28);

    var corCorpo = marca === "ring" ? COR.coral : marca === "diamond" ? COR.pessego : COR.rosa;
    var corDetalhe = marca === "wave" ? COR.pessego : COR.indigo;

    var corpo = new THREE.Mesh(new THREE.CylinderGeometry(0.058, 0.070, 0.24, 20), mat(corCorpo, 0.26, 0.16));
    corpo.rotation.x = Math.PI / 2;
    corpo.position.set(0, 0.06, 0.04);
    grupo.add(corpo);

    var tampaTras = new THREE.Mesh(new THREE.SphereGeometry(0.070, 18, 16), mat(corDetalhe, 0.3, 0.14));
    tampaTras.position.set(0, 0.06, -0.08);
    grupo.add(tampaTras);

    var ponta = new THREE.Mesh(new THREE.SphereGeometry(0.062, 18, 16), mat(COR.aviso, 0.24, 0.15));
    ponta.position.set(0, 0.06, 0.17);
    grupo.add(ponta);

    var reservatorio = new THREE.Mesh(new THREE.SphereGeometry(0.058, 18, 16), mat(corDetalhe, 0.22, 0.18));
    reservatorio.scale.set(0.95, 1.15, 1.15);
    reservatorio.position.set(0, 0.14, 0.01);
    grupo.add(reservatorio);

    var cabo = new THREE.Mesh(new THREE.CylinderGeometry(0.034, 0.042, 0.17, 16), mat(COR.botao, 0.45, 0.1));
    cabo.position.set(0, -0.05, -0.03);
    cabo.rotation.x = 0.28;
    grupo.add(cabo);

    return grupo;
  }

  function criarPorrete(THREE, marca) {
    var grupo = new THREE.Group();
    grupo.position.set(0.71, 0.03, 0.26);
    grupo.rotation.set(0.20, -0.14, -0.25);

    var corPrincipal = marca === "bar" ? COR.ambar : COR.foco;
    var corFaixa = marca === "bar" ? COR.aviso : COR.pessego;

    var cabo = new THREE.Mesh(new THREE.CylinderGeometry(0.030, 0.036, 0.18, 18), mat(COR.botao, 0.44, 0.1));
    cabo.position.y = -0.14;
    grupo.add(cabo);

    var pomo = new THREE.Mesh(new THREE.SphereGeometry(0.042, 16, 16), mat(corFaixa, 0.3, 0.14));
    pomo.position.y = -0.23;
    grupo.add(pomo);

    var guarda = new THREE.Mesh(new THREE.TorusGeometry(0.052, 0.015, 16, 20), mat(corFaixa, 0.3, 0.15));
    guarda.rotation.x = Math.PI / 2;
    guarda.position.y = -0.05;
    grupo.add(guarda);

    var bastao = new THREE.Mesh(new THREE.CylinderGeometry(0.082, 0.044, 0.30, 22), mat(corPrincipal, 0.3, 0.12));
    bastao.position.y = 0.10;
    grupo.add(bastao);

    var topo = new THREE.Mesh(new THREE.SphereGeometry(0.082, 20, 16), mat(corPrincipal, 0.3, 0.12));
    topo.position.y = 0.25;
    grupo.add(topo);

    var anel = new THREE.Mesh(new THREE.TorusGeometry(0.072, 0.013, 16, 24), mat(corFaixa, 0.28, 0.16));
    anel.rotation.x = Math.PI / 2;
    anel.position.y = 0.18;
    grupo.add(anel);

    return grupo;
  }

  edicao.vestirMascote = function (botGroup, markType) {
    if (typeof document === "undefined" || !window.THREE) return;
    if (document.documentElement.getAttribute("data-edicao") !== "missoes") return;
    if (!botGroup || botGroup.userData.edicaoVestido) return;

    var itens = edicao.acessoriosPorMarca[markType];
    if (!itens || !itens.length) return;

    var THREE = window.THREE;
    var acessorios = new THREE.Group();
    acessorios.name = "edicao-acessorios";

    for (var i = 0; i < itens.length; i++) {
      var tipo = itens[i];
      if (tipo === "bandana") acessorios.add(criarBandana(THREE, markType));
      else if (tipo === "oculos") acessorios.add(criarOculos(THREE, markType));
      else if (tipo === "bone") acessorios.add(criarBone(THREE, false));
      else if (tipo === "bonereto") acessorios.add(criarBone(THREE, true));
      else if (tipo === "panama") acessorios.add(criarPanama(THREE));
      else if (tipo === "charuto") acessorios.add(criarCharuto(THREE));
      else if (tipo === "corrente") acessorios.add(criarCorrente(THREE));
      else if (tipo === "brinco") acessorios.add(criarBrinco(THREE));
      else if (tipo === "colarinho") acessorios.add(criarColarinho(THREE));
      else if (tipo === "arminha") acessorios.add(criarArminha(THREE, markType));
      else if (tipo === "porrete") acessorios.add(criarPorrete(THREE, markType));
    }

    botGroup.add(acessorios);
    botGroup.userData.edicaoVestido = true;
  };
})();
