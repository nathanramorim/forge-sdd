// Assets embarcados do Forge-SDD Studio
// Motor 3D Three.js completo dos 5 Mascotes + Fallback SVG 2D,
// Sincronização SSE em tempo real, Visualização do que cada agente está executando e Modo DEMO.

const STATE_DEFS = {
  idle:      { orb: '#4f8ef7', orbHex: 0x4f8ef7, eyes: 'cap' },
  working:   { orb: '#f2a65a', orbHex: 0xf2a65a, eyes: 'round' },
  thinking:  { orb: '#c9a7ff', orbHex: 0xc9a7ff, eyes: 'round' },
  searching: { orb: '#38c6e8', orbHex: 0x38c6e8, eyes: 'round' },
  approval:  { orb: '#ff8a5c', orbHex: 0xff8a5c, eyes: 'cap' },
  finished:  { orb: '#3ecf8e', orbHex: 0x3ecf8e, eyes: 'happy' },
  error:     { orb: '#f2677a', orbHex: 0xf2677a, eyes: 'x' }
};

let currentMode = 'LIVE'; // LIVE | DEMO
let demoInterval = null;
let currentDemoIdx = 0;
let lastSnapshot = null;
const mascotControllers = {}; // id -> MascotController

let mouseLookX = 0;
let mouseLookY = 0;

window.addEventListener('mousemove', (e) => {
  const cx = window.innerWidth / 2;
  const cy = window.innerHeight / 2;
  mouseLookX = (e.clientX - cx) / cx;
  mouseLookY = (e.clientY - cy) / cy;
});

// ============================================================================
// MOTOR 3D THREE.JS (Idêntico ao do site oficial / showcase)
// ============================================================================
function parseColorHex(c) {
  if (typeof c === 'number') return c;
  if (typeof c === 'string') {
    return parseInt(c.replace('#', ''), 16);
  }
  return 0x4f8ef7;
}

function createForgeMascot3D(mountEl, colorVal, initialState, markType) {
  if (typeof THREE === 'undefined') {
    return null;
  }

  const w = 78;
  const h = 86;
  const agentHex = parseColorHex(colorVal);

  let renderer;
  try {
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'low-power' });
  } catch (err) {
    console.warn('WebGL indisponível, usando fallback SVG:', err);
    return null;
  }

  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
  renderer.setSize(w, h, false);
  mountEl.innerHTML = '';
  mountEl.appendChild(renderer.domElement);

  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(34, w / h, 0.1, 50);
  camera.position.set(0, 0.15, 3.55);
  camera.lookAt(0, 0.08, 0);

  scene.add(new THREE.HemisphereLight(0xffffff, 0x10131a, 0.95));
  const keyLight = new THREE.DirectionalLight(0xffffff, 0.9);
  keyLight.position.set(2, 4, 4);
  scene.add(keyLight);

  const botGroup = new THREE.Group();
  scene.add(botGroup);

  const shadowMesh = new THREE.Mesh(
    new THREE.CircleGeometry(0.55, 24),
    new THREE.MeshBasicMaterial({ color: 0x000000, transparent: true, opacity: 0.35 })
  );
  shadowMesh.rotation.x = -Math.PI / 2;
  shadowMesh.scale.set(1, 0.35, 1);
  shadowMesh.position.y = -0.82;
  scene.add(shadowMesh);

  const bodyMat = new THREE.MeshStandardMaterial({
    color: agentHex,
    roughness: 0.28,
    metalness: 0.12
  });
  const body = new THREE.Mesh(new THREE.SphereGeometry(0.68, 32, 32), bodyMat);
  body.scale.set(1.02, 1.08, 0.88);
  body.position.y = -0.04;
  botGroup.add(body);

  const spec = new THREE.Mesh(
    new THREE.CircleGeometry(0.16, 20),
    new THREE.MeshBasicMaterial({ color: 0xffffff, transparent: true, opacity: 0.32 })
  );
  spec.scale.set(1.4, 0.7, 1);
  spec.position.set(-0.26, 0.36, 0.56);
  spec.rotation.z = 0.45;
  botGroup.add(spec);

  const armGeo = new THREE.SphereGeometry(0.11, 16, 16);
  const armL = new THREE.Mesh(armGeo, bodyMat);
  armL.scale.set(0.75, 1.35, 0.8);
  armL.position.set(-0.78, -0.1, 0);
  botGroup.add(armL);

  const armR = new THREE.Mesh(armGeo, bodyMat);
  armR.scale.set(0.75, 1.35, 0.8);
  armR.position.set(0.78, -0.1, 0);
  botGroup.add(armR);

  const markMat = new THREE.MeshBasicMaterial({ color: 0x10131a, transparent: true, opacity: 0.55 });
  let markMesh;
  if (markType === 'ring') {
    markMesh = new THREE.Mesh(new THREE.RingGeometry(0.055, 0.095, 24), markMat);
  } else if (markType === 'diamond') {
    markMesh = new THREE.Mesh(new THREE.PlaneGeometry(0.14, 0.14), markMat);
    markMesh.rotation.z = Math.PI / 4;
  } else if (markType === 'bar') {
    markMesh = new THREE.Mesh(new THREE.PlaneGeometry(0.24, 0.065), markMat);
  } else if (markType === 'tri') {
    markMesh = new THREE.Mesh(new THREE.CircleGeometry(0.095, 3), markMat);
    markMesh.rotation.z = Math.PI / 2;
  } else {
    markMesh = new THREE.Mesh(new THREE.CircleGeometry(0.07, 20), markMat);
  }
  markMesh.position.set(0, 0.31, 0.59);
  botGroup.add(markMesh);

  const antGroup = new THREE.Group();
  antGroup.position.set(0, 0.68, 0);
  botGroup.add(antGroup);

  const stem = new THREE.Mesh(
    new THREE.CylinderGeometry(0.028, 0.028, 0.22, 12),
    new THREE.MeshStandardMaterial({ color: 0x99a2b8, roughness: 0.4 })
  );
  stem.position.y = 0.09;
  antGroup.add(stem);

  const orbMat = new THREE.MeshBasicMaterial({ color: 0x4f8ef7 });
  const orb = new THREE.Mesh(new THREE.SphereGeometry(0.125, 20, 20), orbMat);
  orb.position.y = 0.24;
  antGroup.add(orb);

  const visor = new THREE.Mesh(
    new THREE.SphereGeometry(0.52, 28, 28),
    new THREE.MeshStandardMaterial({ color: 0x10131a, roughness: 0.15, metalness: 0.3 })
  );
  visor.scale.set(1.02, 0.48, 0.42);
  visor.position.set(0, -0.1, 0.46);
  botGroup.add(visor);

  const eyesRoot = new THREE.Group();
  eyesRoot.position.set(0, -0.1, 0.66);
  botGroup.add(eyesRoot);

  const eyeColorMat = new THREE.MeshBasicMaterial({ color: 0xcfe3ff });

  const capGroup = new THREE.Group();
  const capGeo = new THREE.BoxGeometry(0.095, 0.21, 0.02);
  const capL = new THREE.Mesh(capGeo, eyeColorMat);
  capL.position.x = -0.21;
  const capR = new THREE.Mesh(capGeo, eyeColorMat);
  capR.position.x = 0.21;
  capGroup.add(capL, capR);
  eyesRoot.add(capGroup);

  const roundGroup = new THREE.Group();
  const roundGeo = new THREE.CircleGeometry(0.09, 20);
  const roundL = new THREE.Mesh(roundGeo, eyeColorMat);
  roundL.position.x = -0.21;
  const roundR = new THREE.Mesh(roundGeo, eyeColorMat);
  roundR.position.x = 0.21;
  roundGroup.add(roundL, roundR);
  eyesRoot.add(roundGroup);

  const happyGroup = new THREE.Group();
  const arcGeo = new THREE.TorusGeometry(0.08, 0.024, 10, 20, Math.PI);
  const arcL = new THREE.Mesh(arcGeo, eyeColorMat);
  arcL.position.set(-0.21, -0.02, 0);
  const arcR = new THREE.Mesh(arcGeo, eyeColorMat);
  arcR.position.set(0.21, -0.02, 0);
  happyGroup.add(arcL, arcR);
  eyesRoot.add(happyGroup);

  const xGroup = new THREE.Group();
  const barGeo = new THREE.BoxGeometry(0.16, 0.032, 0.02);
  [-0.21, 0.21].forEach((xPos) => {
    const b1 = new THREE.Mesh(barGeo, eyeColorMat);
    b1.position.x = xPos;
    b1.rotation.z = Math.PI / 4;
    const b2 = new THREE.Mesh(barGeo, eyeColorMat);
    b2.position.x = xPos;
    b2.rotation.z = -Math.PI / 4;
    xGroup.add(b1, b2);
  });
  eyesRoot.add(xGroup);

  let currentState = initialState || 'working';
  let hopTimer = 0;

  function setState(nextState, nextColorHex) {
    currentState = STATE_DEFS[nextState] ? nextState : 'working';
    if (nextColorHex !== undefined) bodyMat.color.setHex(parseColorHex(nextColorHex));
    const def = STATE_DEFS[currentState];
    orbMat.color.setHex(def.orbHex);
    capGroup.visible = def.eyes === 'cap';
    roundGroup.visible = def.eyes === 'round';
    happyGroup.visible = def.eyes === 'happy';
    xGroup.visible = def.eyes === 'x';
    hopTimer = 0.45;
  }

  setState(currentState, agentHex);

  return {
    setState,
    triggerHop: () => { hopTimer = 0.55; },
    update: (t, lookX, lookY) => {
      let bobY = Math.sin(t * 2.4) * 0.055;
      if (hopTimer > 0) {
        hopTimer = Math.max(0, hopTimer - 0.016);
        bobY += Math.sin((hopTimer / 0.55) * Math.PI) * 0.18;
      }
      botGroup.position.y = bobY;
      shadowMesh.scale.x = 1 - bobY * 0.6;

      let scanX = 0;
      if (currentState === 'working' || currentState === 'searching') {
        scanX = Math.sin(t * 2.6) * 0.045;
      }
      eyesRoot.position.x = Math.max(-0.09, Math.min(0.09, lookX * 0.08 + scanX));
      eyesRoot.position.y = -0.1 + Math.max(-0.06, Math.min(0.06, -lookY * 0.06));
      botGroup.rotation.y = Math.max(-0.25, Math.min(0.25, lookX * 0.18));

      const blink = (t % 4.5) > 4.36 ? 0.08 : 1;
      eyesRoot.scale.y = blink;

      if (currentState === 'approval' || currentState === 'working') {
        antGroup.rotation.z = Math.sin(t * 4.2) * 0.12;
      } else {
        antGroup.rotation.z = Math.sin(t * 1.5) * 0.04;
      }

      if (currentState === 'error') {
        botGroup.position.x = Math.sin(t * 28) * 0.025;
      } else {
        botGroup.position.x = 0;
      }

      renderer.render(scene, camera);
    }
  };
}

// Loop contínuo de renderização RAF
function startAnimationLoop() {
  function loop(time) {
    const t = time * 0.001;
    for (const id in mascotControllers) {
      const ctrl = mascotControllers[id];
      if (ctrl && ctrl.update) {
        ctrl.update(t, mouseLookX, mouseLookY);
      }
    }
    requestAnimationFrame(loop);
  }
  requestAnimationFrame(loop);
}
startAnimationLoop();

// Fallback SVG 2D de alta fidelidade
function renderMascotFallbackSvg(mountEl, color, mark, state) {
  const sDef = STATE_DEFS[state] || STATE_DEFS.idle;
  let markSvg = '';
  if (mark === 'ring') {
    markSvg = `<circle cx="60" cy="51" r="7" fill="none" stroke="rgba(18,24,48,0.7)" stroke-width="3"/>`;
  } else if (mark === 'diamond') {
    markSvg = `<rect x="54" y="45" width="12" height="12" fill="rgba(18,24,48,0.7)" transform="rotate(45 60 51)"/>`;
  } else if (mark === 'bar') {
    markSvg = `<rect x="50" y="48" width="20" height="6" rx="3" fill="rgba(18,24,48,0.7)"/>`;
  } else if (mark === 'tri') {
    markSvg = `<polygon points="60,45 67,56 53,56" fill="rgba(18,24,48,0.7)"/>`;
  } else {
    markSvg = `<path d="M52 51 Q56 47 60 51 T68 51" fill="none" stroke="rgba(18,24,48,0.7)" stroke-width="3" stroke-linecap="round"/>`;
  }

  let eyeSvg = '';
  if (state === 'happy' || state === 'finished') {
    eyeSvg = `<path d="M40 88 Q45 80 50 88" stroke="${sDef.orb}" stroke-width="3.5" fill="none" stroke-linecap="round"/>
              <path d="M70 88 Q75 80 80 88" stroke="${sDef.orb}" stroke-width="3.5" fill="none" stroke-linecap="round"/>`;
  } else if (state === 'error') {
    eyeSvg = `<g stroke="${sDef.orb}" stroke-width="3" stroke-linecap="round">
                <line x1="42" y1="82" x2="48" y2="90"/><line x1="48" y1="82" x2="42" y2="90"/>
                <line x1="72" y1="82" x2="78" y2="90"/><line x1="78" y1="82" x2="72" y2="90"/>
              </g>`;
  } else {
    eyeSvg = `<circle cx="45" cy="86" r="6.5" fill="${sDef.orb}"/>
              <circle cx="75" cy="86" r="6.5" fill="${sDef.orb}"/>`;
  }

  mountEl.innerHTML = `
    <svg class="mascot-fallback-svg" viewBox="0 0 120 134" width="78" height="86" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <ellipse cx="60" cy="127" rx="32" ry="4.5" fill="rgba(0,0,0,0.38)"/>
      <ellipse cx="13" cy="90" rx="7" ry="11" fill="${color}"/>
      <ellipse cx="107" cy="90" rx="7" ry="11" fill="${color}"/>
      <line x1="60" y1="30" x2="60" y2="17" stroke="${color}" stroke-width="3.2" stroke-linecap="round"/>
      <circle cx="60" cy="11" r="14" fill="${sDef.orb}" opacity="0.3"/>
      <circle cx="60" cy="11" r="7.8" fill="${sDef.orb}"/>
      <path d="M60 28 C86 28 102 52 105 80 C108 106 92 118 60 118 C28 118 12 106 15 80 C18 52 34 28 60 28 Z" fill="${color}"/>
      ${markSvg}
      <rect x="25" y="70" width="70" height="32" rx="16" fill="#10131a" stroke="rgba(255,255,255,0.1)" stroke-width="1"/>
      ${eyeSvg}
    </svg>
  `;
}

// Renderiza a grade de estações e atualiza os DOMs
function applySnapshot(snap) {
  if (!snap) return;
  lastSnapshot = snap;

  document.getElementById('projectName').textContent = snap.project_name || 'forge-sdd';
  document.getElementById('projectVersion').textContent = 'v' + (snap.version || '3.2.1-beta');
  document.getElementById('lastUpdate').textContent = new Date(snap.updated_at).toLocaleTimeString();

  // Handoff Bar
  const ho = snap.last_handoff;
  if (ho) {
    document.getElementById('hoFrom').textContent = ho.from_station;
    document.getElementById('hoTo').textContent = ho.to_station;
    document.getElementById('hoPayload').textContent = ho.payload;
    document.getElementById('hoGate').innerHTML = `<i></i>${ho.gate}`;
  }

  // Grid das 5 estações
  const grid = document.getElementById('stationsGrid');
  grid.innerHTML = '';

  snap.stations.forEach((st, idx) => {
    const isActive = st.id === snap.active_station;
    const card = document.createElement('div');
    card.className = `station-card ${isActive ? 'active' : ''}`;
    card.style.setProperty('--c', st.color);

    // O QUE ESTÁ EXECUTANDO POR AGENTE (Destaque Principal)
    const execTask = st.executing_task;
    let execHtml = '';
    if (execTask && (execTask.status === 'building' || execTask.status === 'in_progress' || execTask.status === 'review_pending' || execTask.status === 'changes_requested')) {
      const isRetrying = execTask.status === 'changes_requested';
      const label = isRetrying ? 'RETRABALHO SOLICITADO' : (execTask.status === 'review_pending' ? 'AUDITANDO' : 'EM EXECUÇÃO');
      execHtml = `
        <div class="executing-box active" style="--c:${isRetrying ? 'var(--err)' : st.color}">
          <div class="executing-header">
            <span class="pulse-dot"></span>
            <span class="executing-label">${label}</span>
            <span class="executing-id">${execTask.id}</span>
          </div>
          <div class="executing-title">${execTask.title || st.current_focus}</div>
        </div>
      `;
    } else if (st.mascot_state === 'working' || st.mascot_state === 'thinking' || st.mascot_state === 'searching') {
      execHtml = `
        <div class="executing-box active" style="--c:${st.color}">
          <div class="executing-header">
            <span class="pulse-dot"></span>
            <span class="executing-label">ATIVIDADE VIVA</span>
          </div>
          <div class="executing-title">${st.current_focus || 'Processando etapa do ciclo'}</div>
        </div>
      `;
    } else {
      execHtml = `
        <div class="executing-box idle">
          <div class="executing-header">
            <span class="idle-dot"></span>
            <span class="executing-label">STATUS</span>
            <span class="executing-id" style="color:var(--muted);font-size:9.5px;">STANDBY</span>
          </div>
          <div class="executing-title muted">${st.current_focus || 'Fila ociosa — Pronto para novo handoff'}</div>
        </div>
      `;
    }

    // Itens na Fila
    let itemsHtml = '';
    if (st.items && st.items.length > 0) {
      itemsHtml = st.items.map(it => `
        <div class="queue-task-row">
          <span style="color:var(--fg);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:140px;" title="${it.title || it.id}">${it.id}</span>
          <span class="chip" style="font-size:9.5px;padding:1px 6px;">${it.status}</span>
        </div>
      `).join('');
    } else {
      itemsHtml = `<div style="color:var(--muted);font-style:italic;">Nenhum item pendente</div>`;
    }

    card.innerHTML = `
      <div class="station-top">
        <div class="mascot-stage" id="stage-${st.id}"></div>
        <div class="station-info">
          <h3>${st.name}</h3>
          <p class="role">${st.role}</p>
          <span class="chip" style="margin-top:4px;"><i style="--o:${st.color}"></i>${st.mascot_state}</span>
        </div>
      </div>

      ${execHtml}

      <div class="io-strip">
        <div class="io-item"><span>Recebe</span><b>${st.input_contract}</b></div>
        <div class="io-item"><span>Entrega</span><b>${st.output_contract}</b></div>
      </div>

      <div class="queue-items-box">
        <div class="queue-header-row">
          <span>Fila (${st.id}.json)</span>
          <span>v${st.version}</span>
        </div>
        ${itemsHtml}
      </div>
    `;

    grid.appendChild(card);

    // Inicialização ou atualização do Mascote (Three.js 3D com fallback SVG)
    const stage = card.querySelector(`#stage-${st.id}`);
    let ctrl = mascotControllers[st.id];
    if (!ctrl) {
      ctrl = createForgeMascot3D(stage, st.color, st.mascot_state, st.mark);
      if (ctrl) {
        mascotControllers[st.id] = ctrl;
      } else {
        renderMascotFallbackSvg(stage, st.color, st.mark, st.mascot_state);
      }
    } else {
      // Se o elemento DOM foi recriado, remonta o renderer do Three.js
      if (ctrl.setState) {
        ctrl.setState(st.mascot_state, st.color);
        ctrl.triggerHop();
      }
    }
  });

  // Governança
  const gVerdict = snap.governance_verdict;
  const govCard = document.getElementById('govFindingCard');
  if (gVerdict) {
    govCard.className = `finding-card ${gVerdict.veredito}`;
    govCard.innerHTML = `
      <b>Parecer da Camada B: ${gVerdict.veredito.toUpperCase()}</b>
      <p style="margin-top:4px;color:var(--fg);">${gVerdict.resumo || 'Em conformidade com a Constituição'}</p>
    `;
  } else {
    govCard.className = 'finding-card conforme';
    govCard.innerHTML = `
      <b>Portão de Governança Camada A & B</b>
      <p style="margin-top:4px;color:var(--muted);">Aguardando submissão ou inspeção ativa</p>
    `;
  }
}

// Conexão SSE
function initSSE() {
  const es = new EventSource('/api/events');
  es.addEventListener('state', (e) => {
    if (currentMode === 'LIVE') {
      try {
        const snap = JSON.parse(e.data);
        applySnapshot(snap);
      } catch (err) {
        console.error('Falha ao processar SSE:', err);
      }
    }
  });
  es.onerror = () => {
    if (currentMode === 'LIVE') {
      fetch('/api/state')
        .then(r => r.json())
        .then(applySnapshot)
        .catch(console.error);
    }
  };
}

// Chaveador Modo LIVE / DEMO
window.setStudioMode = function(mode) {
  currentMode = mode;
  document.getElementById('btnLive').classList.toggle('active', mode === 'LIVE');
  document.getElementById('btnDemo').classList.toggle('active', mode === 'DEMO');

  const statusChip = document.getElementById('statusChip');
  if (mode === 'LIVE') {
    if (demoInterval) clearInterval(demoInterval);
    statusChip.className = 'chip live';
    statusChip.innerHTML = `<i></i>LIVE SSE`;
    fetch('/api/state').then(r => r.json()).then(applySnapshot);
  } else {
    statusChip.className = 'chip';
    statusChip.innerHTML = `<i style="--o:var(--warn)"></i>DEMO SIM`;
    runDemoSimulation();
  }
};

const DEMO_SIMULATION_STEPS = [
  {
    station: 'lider',
    mascot_state: 'working',
    executing: { id: 'sdd-ciclo-01', title: 'Orquestrando transição de onda e validando portões humanos', status: 'building' },
    gate: 'Gate 1 — Alinhamento de Visão Concluído',
    payload: 'Demanda Estruturada → Estação 02 (PM & Arquiteto)'
  },
  {
    station: 'pm-architect',
    mascot_state: 'thinking',
    executing: { id: 'discovery-60', title: 'Elaborando especificação técnica, diagramas C4 e Reuse Check', status: 'building' },
    gate: 'Gate 2 — Revisão Humana de Spec (FLX-01)',
    payload: 'discovery.md + criteria.md → Estação 03'
  },
  {
    station: 'planner',
    mascot_state: 'working',
    executing: { id: 'plan-60', title: 'Fatiando spec em tarefas P/M/G e mapeando rules por task', status: 'building' },
    gate: 'Gate 3 — Plano Validado e Enfileirado',
    payload: 'plan.md + queue.json → Estação 04 (Builders)'
  },
  {
    station: 'builder',
    mascot_state: 'working',
    executing: { id: 'feat-60-03', title: 'Implementando código Go e suite de testes com detecção de corrida (-race)', status: 'building' },
    gate: 'Gate 4 — Testes Locais e Camada A (PASS)',
    payload: 'Código + Testes Conforme → Estação 05 (Governança)'
  },
  {
    station: 'governance',
    mascot_state: 'approval',
    executing: { id: 'audit-60-03', title: 'Auditoria Read-Only contra Constituição e emissão de laudo JSON', status: 'review_pending' },
    gate: 'Gate 5 — Laudo Camada B Aprovado',
    payload: 'Laudo Conforme → Estação 01 (Líder / PR)'
  }
];

function runDemoSimulation() {
  if (demoInterval) clearInterval(demoInterval);
  demoInterval = setInterval(() => {
    if (!lastSnapshot) return;
    const stations = lastSnapshot.stations;
    currentDemoIdx = (currentDemoIdx + 1) % DEMO_SIMULATION_STEPS.length;
    const simStep = DEMO_SIMULATION_STEPS[currentDemoIdx];

    const cloned = JSON.parse(JSON.stringify(lastSnapshot));
    cloned.active_station = simStep.station;
    cloned.active_station_idx = currentDemoIdx;

    cloned.stations.forEach((s, i) => {
      if (s.id === simStep.station) {
        s.mascot_state = simStep.mascot_state;
        s.executing_task = simStep.executing;
        s.current_focus = simStep.executing.title;
      } else {
        s.mascot_state = (i < currentDemoIdx) ? 'finished' : 'idle';
        s.executing_task = null;
      }
    });

    cloned.last_handoff = {
      from_station: simStep.station,
      to_station: stations[(currentDemoIdx + 1) % stations.length].id,
      payload: simStep.payload,
      gate: simStep.gate,
      is_loop: false
    };

    applySnapshot(cloned);
  }, 2200);
}

// Inicialização
document.addEventListener('DOMContentLoaded', () => {
  fetch('/api/state')
    .then(r => r.json())
    .then(applySnapshot)
    .finally(() => {
      initSSE();
    });
});
