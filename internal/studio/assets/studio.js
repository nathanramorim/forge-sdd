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
    domElement: renderer.domElement,
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

function escapeHtml(str) {
  if (typeof str !== 'string') return '';
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

let currentActivities = [];
let consoleFilter = 'all';

window.filterConsole = function(station) {
  consoleFilter = station;
  document.querySelectorAll('#consoleFilterTabs .console-tab').forEach(tab => {
    tab.classList.toggle('active', tab.dataset.filter === station);
  });
  renderConsoleLogs(currentActivities);
};

function renderConsoleLogs(activities) {
  const container = document.getElementById('consoleLogs');
  if (!container) return;

  const filtered = (activities || []).filter(act => {
    if (consoleFilter === 'all') return true;
    return act.station === consoleFilter;
  });

  if (filtered.length === 0) {
    container.innerHTML = `<div style="color:var(--muted);font-style:italic;padding:8px 0;">Nenhum evento registrado para esta estação.</div>`;
    return;
  }

  container.innerHTML = filtered.slice(0, 50).map(act => {
    const stMeta = {
      'lider': { name: '01. LÍDER', color: '#f2a65a' },
      'pm-architect': { name: '02. PM/ARQ', color: '#8aa4ff' },
      'planner': { name: '03. PLAN', color: '#c9a7ff' },
      'builder': { name: '04. BUILDER', color: '#4fd1b5' },
      'governance': { name: '05. GOV', color: '#ff8a5c' }
    }[act.station] || { name: act.station, color: '#4f8ef7' };

    let timeStr = '--:--:--';
    if (act.timestamp) {
      try {
        const d = new Date(act.timestamp);
        timeStr = d.toTimeString().split(' ')[0] || act.timestamp.slice(11, 19);
      } catch (e) {
        timeStr = act.timestamp.slice(11, 19);
      }
    }
    const statusClass = (act.status === 'running' || act.status === 'working') ? 'running' : (act.status === 'error' ? 'error' : 'done');

    return `
      <div class="log-row">
        <span class="log-time">${escapeHtml(timeStr)}</span>
        <span class="log-station" style="color:${stMeta.color};">
          <span class="tool-chip-dot" style="--c:${stMeta.color};"></span>${escapeHtml(stMeta.name)}
        </span>
        <span class="log-tool">${escapeHtml(act.tool || 'tool')}</span>
        <span class="log-action">
          <b>${escapeHtml(act.action || '')}</b>
          ${act.details ? `<span class="detail-hint">${escapeHtml(act.details)}</span>` : ''}
        </span>
        <span class="log-status ${statusClass}">${escapeHtml(act.status || 'done')}</span>
      </div>
    `;
  }).join('');
}

// ============================================================================
// STATION DRAWER & DEEP-LINKS (feat-63-02, CA-01, CA-02, CA-04, CA-10)
// ============================================================================

let activeDrawerStationId = null;
let activeDrawerTab = 'tasks'; // 'tasks' | 'contracts' | 'telemetry'
const selectedTaskIdByStation = {};

// Pilares Constitucionais com descrições em linguagem clara
const CONSTITUTIONAL_PILLARS = [
  { key: 'ARQ', title: 'Arquitetura Limpa', desc: 'Go stdlib pura, zero npm runtime em produção', icon: '🏗️' },
  { key: 'SEG', title: 'Segurança & Escopo', desc: 'Sanitização de path traversal e proteção read-only', icon: '🛡️' },
  { key: 'DUP', title: 'Fonte Única da Verdade', desc: 'Zero duplicação de spec/código (DUP-01/02)', icon: '🧬' },
  { key: 'COD', title: 'Qualidade de Código & DOM', desc: 'Sem handlers órfãos nem ReferenceError (COD-01)', icon: '🧹' },
  { key: 'TST', title: 'Testes & Concorrência', desc: '100% PASS determinístico com detecção de data race (-race)', icon: '🧪' },
  { key: 'FLX', title: 'Portões & Validação Humana', desc: 'Portão FLX-01 aprovado pelo humano e transições fluidas', icon: '🚪' },
  { key: 'GOV', title: 'Governança Estrita', desc: 'Auditoria independente da Camada B em modo Read-Only', icon: '⚖️' }
];

// Mapeamento de Regras Constitucionais para Badges Humanizados
const HUMAN_RULE_BADGES = {
  'ARQ-01': '🏛️ Arquitetura Canônica',
  'ARQ-02': '📐 C4 & Interfaces Limpas',
  'ARQ-03': '📦 Stdlib Go Pura',
  'ARQ-04': '⚡ Zero NPM Runtime',
  'SEG-01': '🛡️ Sem Traversal',
  'SEG-02': '🔒 Sandbox Estrito',
  'SEG-06': '🔒 Restrito Read-Only',
  'DUP-01': '🧬 Fonte Única',
  'DUP-02': '🔍 Reuse Check',
  'DUP-05': '📑 Template Padronizado',
  'COD-01': '🧹 DOM & Handlers OK',
  'COD-05': '✨ Clean Code & Linters',
  'TST-01': '🧪 Testes com -race',
  'TST-04': '🎯 Cobertura Mínima',
  'TST-05': '⏱️ Determinismo & Timeout',
  'FLX-01': '🚪 Portão FLX-01 Aprovado',
  'FLX-02': '📋 Decomposição em Fatias',
  'FLX-03': '🔄 Handoffs Auditados',
  'FLX-04': '🏁 Encerramento Controlado',
  'GOV-01': '⚖️ Governança Camada B',
  'GOV-02': '🛡️ Read-Only Enforced',
  'GOV-03': '📝 Emissão de Laudo JSON',
  'GOV-04': '🚦 Zero Falsos-Positivos',
  'GOV-05': '🔗 Rastreabilidade Total'
};

function formatRuleBadge(rule) {
  if (!rule) return '';
  const trimmed = String(rule).trim();
  if (HUMAN_RULE_BADGES[trimmed]) {
    return HUMAN_RULE_BADGES[trimmed];
  }
  return trimmed;
}

function formatTaskStatus(status) {
  if (!status) return { text: 'INDEFINIDO', color: 'var(--muted)', bg: 'transparent', border: 'var(--line)' };
  const s = String(status).toLowerCase().trim();
  switch (s) {
    case 'approved':
    case 'done':
      return {
        text: '✓ CONCLUÍDO & APROVADO',
        color: 'var(--ok)',
        bg: 'rgba(62, 207, 142, 0.08)',
        border: 'rgba(62, 207, 142, 0.3)'
      };
    case 'queued':
      return {
        text: '⏳ NA FILA (AGUARDANDO INÍCIO)',
        color: '#8aa4ff',
        bg: 'rgba(138, 164, 255, 0.08)',
        border: 'rgba(138, 164, 255, 0.3)'
      };
    case 'building':
    case 'working':
    case 'in_progress':
      return {
        text: '⚡ EM ANDAMENTO',
        color: 'var(--warn)',
        bg: 'rgba(242, 166, 90, 0.08)',
        border: 'rgba(242, 166, 90, 0.3)'
      };
    case 'review_pending':
    case 'pending_review':
    case 'approval':
      return {
        text: '🔍 AGUARDANDO AUDITORIA',
        color: '#ff8a5c',
        bg: 'rgba(255, 138, 92, 0.08)',
        border: 'rgba(255, 138, 92, 0.3)'
      };
    case 'changes_requested':
      return {
        text: '⚠️ RETRABALHO SOLICITADO',
        color: 'var(--err)',
        bg: 'rgba(255, 92, 92, 0.08)',
        border: 'rgba(255, 92, 92, 0.3)'
      };
    default:
      return {
        text: s.toUpperCase(),
        color: 'var(--fg)',
        bg: 'transparent',
        border: 'var(--line)'
      };
  }
}

window.CONSTITUTIONAL_PILLARS = CONSTITUTIONAL_PILLARS;
window.HUMAN_RULE_BADGES = HUMAN_RULE_BADGES;
window.formatRuleBadge = formatRuleBadge;
window.formatTaskStatus = formatTaskStatus;

const CANONICAL_RULES_BY_STATION = {
  'lider': ['FLX-03', 'FLX-04', 'GOV-01', 'ARQ-01'],
  'pm-architect': ['ARQ-01', 'ARQ-02', 'DUP-01', 'DUP-02', 'FLX-01'],
  'planner': ['ARQ-03', 'FLX-02', 'DUP-05', 'COD-05'],
  'builder': ['ARQ-03', 'ARQ-04', 'DUP-02', 'COD-01', 'COD-05', 'TST-01', 'TST-04', 'TST-05'],
  'governance': ['GOV-01', 'GOV-02', 'GOV-03', 'GOV-04', 'GOV-05', 'SEG-01', 'SEG-02']
};

function getIdeFilePath(artifactPath) {
  if (!artifactPath) return '';
  let projectDir = (lastSnapshot && lastSnapshot.project_dir)
    ? lastSnapshot.project_dir
    : '';
  if (!projectDir) {
    try {
      projectDir = localStorage.getItem('studio_project_dir') || '';
    } catch (e) {}
  }
  projectDir = (projectDir || '').replace(/\/+$/, '');

  // Se já for caminho absoluto no sistema operacional e coincidir com projectDir
  if (projectDir && artifactPath.startsWith(projectDir)) {
    return artifactPath;
  }

  // Se for um caminho absoluto real do sistema operacional (ex: /Users/..., /home/..., /private/...)
  // mas que não comece com /sdd (que seria um caminho relativo com leading slash indevido)
  if (artifactPath.startsWith('/') && !artifactPath.startsWith('/sdd')) {
    return artifactPath;
  }

  // Remove barra inicial para normalizar caminho relativo ao projeto (ex: '/sdd/...' -> 'sdd/...')
  const cleanPath = artifactPath.replace(/^\/+/, '');

  if (projectDir) {
    const absBase = projectDir.startsWith('/') ? projectDir : '/' + projectDir;
    return absBase + '/' + cleanPath;
  }

  // Fallback: se projectDir ainda não estiver setado, nunca retorne '/sdd/...' sem prefixo absoluto!
  return cleanPath;
}

function sortItemsNewestFirst(list) {
  if (!list || !Array.isArray(list) || list.length === 0) return [];
  return [...list].sort((a, b) => {
    const timeA = new Date(a.updated_at || a.created_at || 0).getTime();
    const timeB = new Date(b.updated_at || b.created_at || 0).getTime();
    if (timeB !== timeA) return timeB - timeA;
    return (b.id || '').localeCompare(a.id || '');
  });
}

function fallbackCopy(text) {
  const ta = document.createElement('textarea');
  ta.value = text;
  ta.style.position = 'fixed';
  ta.style.opacity = '0';
  document.body.appendChild(ta);
  ta.select();
  try {
    document.execCommand('copy');
  } catch (e) {
    console.warn('Falha no fallback de cópia:', e);
  }
  document.body.removeChild(ta);
}

function copyCommand(cmd, btn) {
  if (!cmd) return;
  const setSuccess = () => {
    const btnEl = typeof btn === 'string' ? document.getElementById(btn) : btn;
    if (btnEl) {
      const orig = btnEl.innerText;
      btnEl.innerText = 'Copiado!';
      btnEl.classList.add('copied');
      setTimeout(() => {
        btnEl.innerText = orig;
        btnEl.classList.remove('copied');
      }, 1500);
    }
  };

  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(cmd).then(setSuccess).catch(() => {
      fallbackCopy(cmd);
      setSuccess();
    });
  } else {
    fallbackCopy(cmd);
    setSuccess();
  }
}

function inspectSpec(pathOrId) {
  if (!pathOrId) return;
  openSpecViewer(pathOrId);
}

let currentOpenSpecPath = null;

const SPEC_MODAL_FALLBACK_HTML = `
  <div class="spec-modal-header">
    <div class="spec-header-info">
      <div class="spec-header-top-row">
        <span class="chip mono" id="specStationBadge">02. PM & Arquiteto</span>
        <span class="chip mono" id="specStatusBadge">Aguardando Revisão (FLX-01)</span>
      </div>
      <h2 id="specModalTitle">Especificação de Feature</h2>
      <div class="spec-meta-row">
        <span class="spec-path-label mono" id="specPathLabel">sdd/features/...</span>
        <button id="specCopyPathBtn" class="copy-btn mono" onclick="copySpecPath()">Copiar Caminho</button>
      </div>
    </div>
    <div class="spec-header-actions">
      <a id="specVsCodeLink" href="#" class="btn-ide mono" title="Abrir no VS Code">VS Code</a>
      <a id="specCursorLink" href="#" class="btn-ide cursor-btn mono" title="Abrir no Cursor">Cursor</a>
      <button class="spec-close-btn" onclick="closeSpecViewer()" aria-label="Fechar Visualizador">&times;</button>
    </div>
  </div>
  <div class="spec-modal-body" id="specModalBody">
    <div class="markdown-body" id="specMarkdownContent">
      <!-- Conteúdo Markdown formatado -->
    </div>
  </div>
  <div class="spec-modal-footer">
    <div class="spec-footer-info">
      <span class="mono" id="specFileSize">0 bytes</span>
      <span class="spec-gate-hint">Portão FLX-01 — Revisão Técnica & Conformidade Humana</span>
    </div>
    <div class="spec-footer-actions">
      <button class="btn-secondary" onclick="closeSpecViewer()">Fechar</button>
      <button id="btnApproveSpec" class="approve-spec-btn" onclick="handleApproveSpecClick()">
        <span class="approve-icon">✓</span> Aprovar Spec (FLX-01)
      </button>
    </div>
  </div>
`;

function copySpecPath() {
  const path = currentOpenSpecPath || (document.getElementById('specPathLabel') ? document.getElementById('specPathLabel').textContent.trim() : '');
  if (!path) return;
  const btn = document.getElementById('specCopyPathBtn');
  const original = btn ? btn.textContent : 'Copiar Caminho';
  const setSuccess = () => {
    if (btn) {
      btn.textContent = 'Copiado!';
      btn.classList.add('copied');
      setTimeout(() => {
        btn.textContent = original;
        btn.classList.remove('copied');
      }, 1500);
    }
    showToast(`Caminho copiado: ${path}`);
  };

  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(path).then(setSuccess).catch(() => {
      fallbackCopy(path);
      setSuccess();
    });
  } else {
    fallbackCopy(path);
    setSuccess();
  }
}

function handleApproveSpecClick() {
  if (currentOpenSpecPath) {
    approveSpec(currentOpenSpecPath);
  }
}

async function openSpecViewer(specPath) {
  const modal = document.getElementById('specViewerModal');
  const backdrop = document.getElementById('specModalBackdrop');
  if (!modal || !backdrop) return;

  const targetPath = specPath || 'sdd/discovery/discovery-63-studio-navegavel-aprovacao-spec.md';
  currentOpenSpecPath = targetPath;

  if (!document.getElementById('specMarkdownContent')) {
    modal.innerHTML = SPEC_MODAL_FALLBACK_HTML;
  }

  const titleEl = document.getElementById('specModalTitle');
  const pathLabelEl = document.getElementById('specPathLabel');
  const markdownEl = document.getElementById('specMarkdownContent');
  const statusBadge = document.getElementById('specStatusBadge');
  const stationBadge = document.getElementById('specStationBadge');
  const vsCodeLink = document.getElementById('specVsCodeLink');
  const cursorLink = document.getElementById('specCursorLink');
  const fileSizeEl = document.getElementById('specFileSize');
  const btnApprove = document.getElementById('btnApproveSpec');

  if (titleEl) titleEl.textContent = targetPath.split('/').pop() || 'Especificação de Feature';
  if (pathLabelEl) pathLabelEl.textContent = targetPath;
  if (stationBadge) stationBadge.textContent = '02. PM & Arquiteto';
  if (statusBadge) {
    statusBadge.textContent = 'Carregando especificação...';
    statusBadge.style.color = 'var(--muted)';
    statusBadge.style.borderColor = 'currentColor';
  }
  if (markdownEl) {
    markdownEl.innerHTML = '<div style="text-align:center;padding:40px;color:var(--muted);">Carregando especificação...</div>';
  }

  modal.classList.add('open');
  backdrop.classList.add('open');

  try {
    const res = await fetch(`/api/spec?path=${encodeURIComponent(targetPath)}`);
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      if (markdownEl) {
        markdownEl.innerHTML = `
          <div style="background:rgba(235,94,85,0.1);border:1px solid var(--err);border-radius:8px;padding:20px;color:var(--err);">
            <b>Erro ao carregar especificação</b>
            <p style="margin-top:6px;font-size:13px;">${escapeHtml(errData.error || res.statusText)}</p>
          </div>
        `;
      }
      return;
    }

    const data = await res.json();
    const absPath = data.absolute_path || getIdeFilePath(data.path);
    const vsCodeUri = absPath ? (absPath.startsWith('/') ? `vscode://file${absPath}` : `vscode://file/${absPath}`) : 'vscode://';
    const cursorUri = absPath ? (absPath.startsWith('/') ? `cursor://file${absPath}` : `cursor://file/${absPath}`) : 'cursor://';
    const isApproved = data.status === 'approved';

    if (stationBadge) {
      stationBadge.textContent = data.station ? (data.station === 'pm-architect' ? '02. PM & Arquiteto' : data.station) : '02. PM & Arquiteto';
    }
    if (statusBadge) {
      statusBadge.textContent = isApproved ? '✓ Aprovada (FLX-01)' : 'Aguardando Revisão (FLX-01)';
      statusBadge.style.color = isApproved ? 'var(--ok)' : 'var(--warn)';
      statusBadge.style.borderColor = 'currentColor';
    }
    if (pathLabelEl) pathLabelEl.textContent = data.path;
    if (vsCodeLink) vsCodeLink.href = vsCodeUri;
    if (cursorLink) cursorLink.href = cursorUri;
    if (fileSizeEl) fileSizeEl.textContent = `${data.size || (data.content ? data.content.length : 0)} bytes`;
    if (markdownEl) markdownEl.innerHTML = renderMarkdown(data.content);

    if (btnApprove) {
      btnApprove.disabled = isApproved;
      if (isApproved) {
        btnApprove.classList.add('approved');
        btnApprove.innerHTML = '<span class="approve-icon">✓</span> Spec Aprovada (FLX-01)';
      } else {
        btnApprove.classList.remove('approved');
        btnApprove.innerHTML = '<span class="approve-icon">✓</span> Aprovar Spec (FLX-01)';
      }
    }
  } catch (err) {
    if (markdownEl) {
      markdownEl.innerHTML = `
        <div style="background:rgba(235,94,85,0.1);border:1px solid var(--err);border-radius:8px;padding:20px;color:var(--err);">
          <b>Falha na comunicação com o servidor</b>
          <p style="margin-top:6px;font-size:13px;">${escapeHtml(err.message)}</p>
        </div>
      `;
    }
  }
}

function closeSpecViewer() {
  currentOpenSpecPath = null;
  const modal = document.getElementById('specViewerModal');
  const backdrop = document.getElementById('specModalBackdrop');
  if (modal) modal.classList.remove('open');
  if (backdrop) backdrop.classList.remove('open');
}

async function approveSpec(specPath) {
  const targetPath = specPath || currentOpenSpecPath;
  if (!targetPath) return;

  try {
    const res = await fetch('/api/spec/approve', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ spec_path: targetPath })
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      alert(`Erro ao aprovar spec: ${errData.error || res.statusText}`);
      return;
    }

    const data = await res.json();
    showToast(`✓ Especificação aprovada no Portão FLX-01 e enviada para o Planejador (03): ${data.spec_path}`);

    const btnApprove = document.getElementById('btnApproveSpec');
    if (btnApprove) {
      btnApprove.disabled = true;
      btnApprove.classList.add('approved');
      btnApprove.innerHTML = '<span class="approve-icon">✓</span> Spec Aprovada (FLX-01)';
    }
    const statusBadge = document.getElementById('specStatusBadge');
    if (statusBadge) {
      statusBadge.textContent = '✓ Aprovada (FLX-01)';
      statusBadge.style.color = 'var(--ok)';
      statusBadge.style.borderColor = 'currentColor';
    }

    // Fecha o modal após aprovação com sucesso
    setTimeout(() => {
      closeSpecViewer();
    }, 600);

    // Atualiza estado se estiver em LIVE
    if (currentMode === 'LIVE') {
      fetch('/api/state').then(r => r.json()).then(applySnapshot).catch(console.error);
    }
  } catch (err) {
    alert(`Falha ao conectar ao servidor: ${err.message}`);
  }
}

function showToast(msg) {
  let toast = document.getElementById('studioToast');
  if (!toast) {
    toast = document.createElement('div');
    toast.id = 'studioToast';
    toast.style.cssText = 'position:fixed;bottom:24px;right:24px;background:#1a2332;border:1px solid var(--ok);color:#3ecf8e;padding:12px 20px;border-radius:8px;font:600 13px var(--f-mono);box-shadow:0 8px 30px rgba(0,0,0,0.5);z-index:9999;transition:all 0.3s ease;transform:translateY(100px);opacity:0;';
    document.body.appendChild(toast);
  }
  toast.textContent = msg;
  toast.style.transform = 'translateY(0)';
  toast.style.opacity = '1';
  setTimeout(() => {
    toast.style.transform = 'translateY(100px)';
    toast.style.opacity = '0';
  }, 3500);
}

async function sendHandoff(from, to, taskId, title, artifactPath) {
  try {
    const res = await fetch('/api/handoff', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        from: from,
        to: to,
        task_id: taskId,
        title: title,
        artifact_path: artifactPath
      })
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      alert(`Erro na transferência de handoff: ${errData.error || res.statusText}`);
      return null;
    }

    const data = await res.json();
    showToast(`✓ Artefato transferido: ${from} → ${to}`);

    if (currentMode === 'LIVE') {
      fetch('/api/state')
        .then(r => r.json())
        .then(applySnapshot)
        .catch(console.error);
    }
    return data;
  } catch (err) {
    alert(`Falha na comunicação de handoff: ${err.message}`);
    return null;
  }
}

function handleHandoffClick(btn) {
  if (!btn) return;
  const from = btn.getAttribute('data-from');
  const to = btn.getAttribute('data-to');
  const taskId = btn.getAttribute('data-id');
  const title = btn.getAttribute('data-title');
  const artifactPath = btn.getAttribute('data-artifact');
  window.sendHandoff(from, to, taskId, title, artifactPath);
}

function renderMarkdown(md) {
  if (!md) return '';
  const lines = md.split('\n');
  let html = '';
  let inCode = false;
  let codeBuffer = '';
  let inList = false;
  let inTable = false;
  let tableHeader = true;

  const escapeText = (t) => t
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

  const formatInline = (t) => {
    let res = escapeText(t);
    res = res.replace(/`([^`]+)`/g, '<code>$1</code>');
    res = res.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    res = res.replace(/\*([^*]+)\*/g, '<em>$1</em>');
    return res;
  };

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // Code blocks
    if (line.trim().startsWith('```')) {
      if (inCode) {
        html += `<pre><code>${escapeText(codeBuffer)}</code></pre>\n`;
        codeBuffer = '';
        inCode = false;
      } else {
        if (inList) { html += '</ul>\n'; inList = false; }
        if (inTable) { html += '</tbody></table>\n'; inTable = false; }
        inCode = true;
      }
      continue;
    }
    if (inCode) {
      codeBuffer += (codeBuffer ? '\n' : '') + line;
      continue;
    }

    const trimmed = line.trim();

    // Tables
    if (trimmed.startsWith('|') && trimmed.endsWith('|')) {
      const isDivider = trimmed.split('|').filter(c => c.trim()).every(c => /^[-:]+$/.test(c.trim()));
      if (isDivider) {
        tableHeader = false;
        continue;
      }
      const cells = trimmed.split('|').slice(1, -1).map(c => c.trim());
      if (!inTable) {
        if (inList) { html += '</ul>\n'; inList = false; }
        html += '<table>\n';
        inTable = true;
        tableHeader = true;
      }
      if (tableHeader) {
        html += '<thead><tr>' + cells.map(c => `<th>${formatInline(c)}</th>`).join('') + '</tr></thead>\n<tbody>\n';
      } else {
        html += '<tr>' + cells.map(c => `<td>${formatInline(c)}</td>`).join('') + '</tr>\n';
      }
      continue;
    } else if (inTable) {
      html += '</tbody></table>\n';
      inTable = false;
    }

    // Checklist
    if (/^[-*]\s+\[([ xX])\]\s+(.*)$/.test(trimmed)) {
      if (!inList) { html += '<ul style="list-style:none;padding-left:4px;">\n'; inList = true; }
      const match = trimmed.match(/^[-*]\s+\[([ xX])\]\s+(.*)$/);
      const checked = match[1].toLowerCase() === 'x';
      html += `<li class="checklist-item"><input type="checkbox" ${checked ? 'checked' : ''} disabled> <span>${formatInline(match[2])}</span></li>\n`;
      continue;
    }

    // Bullet lists
    if (/^[-*]\s+(.*)$/.test(trimmed)) {
      if (!inList) { html += '<ul>\n'; inList = true; }
      const itemText = trimmed.replace(/^[-*]\s+/, '');
      html += `<li>${formatInline(itemText)}</li>\n`;
      continue;
    } else if (inList) {
      html += '</ul>\n';
      inList = false;
    }

    // Empty line
    if (!trimmed) {
      continue;
    }

    // Headers
    if (trimmed.startsWith('#### ')) {
      html += `<h4>${formatInline(trimmed.substring(5))}</h4>\n`;
    } else if (trimmed.startsWith('### ')) {
      html += `<h3>${formatInline(trimmed.substring(4))}</h3>\n`;
    } else if (trimmed.startsWith('## ')) {
      html += `<h2>${formatInline(trimmed.substring(3))}</h2>\n`;
    } else if (trimmed.startsWith('# ')) {
      html += `<h1>${formatInline(trimmed.substring(2))}</h1>\n`;
    } else if (trimmed.startsWith('> ')) {
      html += `<blockquote>${formatInline(trimmed.substring(2))}</blockquote>\n`;
    } else if (trimmed.startsWith('---') || trimmed.startsWith('***')) {
      html += '<hr style="border:none;border-top:1px solid var(--line);margin:18px 0;">\n';
    } else {
      html += `<p>${formatInline(trimmed)}</p>\n`;
    }
  }

  if (inCode) {
    html += `<pre><code>${escapeText(codeBuffer)}</code></pre>\n`;
  }
  if (inList) {
    html += '</ul>\n';
  }
  if (inTable) {
    html += '</tbody></table>\n';
  }

  return html;
}

function selectTaskInDrawer(taskId) {
  if (!activeDrawerStationId) return;
  selectedTaskIdByStation[activeDrawerStationId] = taskId;
  if (lastSnapshot && lastSnapshot.stations) {
    const st = lastSnapshot.stations.find(s => s.id === activeDrawerStationId);
    if (st) {
      renderStationDrawerContent(st);
    }
  }
}

function switchDrawerTab(tab) {
  activeDrawerTab = tab;
  if (activeDrawerStationId && lastSnapshot && lastSnapshot.stations) {
    const st = lastSnapshot.stations.find(s => s.id === activeDrawerStationId);
    if (st) {
      renderStationDrawerContent(st);
    }
  }
}

function openStationDrawer(stationId, tab) {
  activeDrawerStationId = stationId;
  if (tab) {
    activeDrawerTab = tab;
  }
  const drawer = document.getElementById('stationDrawer');
  const backdrop = document.getElementById('drawerBackdrop');
  if (!drawer || !backdrop) return;

  drawer.classList.add('open');
  drawer.removeAttribute('aria-hidden');
  backdrop.classList.add('open');

  if (lastSnapshot && lastSnapshot.stations) {
    const st = lastSnapshot.stations.find(s => s.id === stationId);
    if (st) {
      renderStationDrawerContent(st);
    }
  }
}

function closeStationDrawer() {
  activeDrawerStationId = null;
  const drawer = document.getElementById('stationDrawer');
  const backdrop = document.getElementById('drawerBackdrop');
  if (drawer) {
    drawer.classList.remove('open');
    drawer.setAttribute('aria-hidden', 'true');
  }
  if (backdrop) {
    backdrop.classList.remove('open');
  }
}

window.openStationDrawer = openStationDrawer;
window.closeStationDrawer = closeStationDrawer;
window.switchDrawerTab = switchDrawerTab;
window.selectTaskInDrawer = selectTaskInDrawer;
window.copyCommand = copyCommand;
window.copySpecPath = copySpecPath;
window.handleApproveSpecClick = handleApproveSpecClick;
window.inspectSpec = inspectSpec;
window.openSpecViewer = openSpecViewer;
window.closeSpecViewer = closeSpecViewer;
window.approveSpec = approveSpec;
window.sendHandoff = sendHandoff;
window.handleHandoffClick = handleHandoffClick;
window.renderMarkdown = renderMarkdown;
window.renderMarkdownToHtml = renderMarkdown;

window.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    const modal = document.getElementById('specViewerModal');
    if (modal && modal.classList.contains('open')) {
      closeSpecViewer();
    } else if (activeDrawerStationId) {
      closeStationDrawer();
    }
  }
});

function getToolColor(tool) {
  const t = (tool || '').toLowerCase();
  if (t.includes('git')) return '#f05032';
  if (t.includes('go') || t.includes('test')) return '#00add8';
  if (t.includes('command') || t.includes('sh') || t.includes('cmd')) return '#4f8ef7';
  if (t.includes('read')) return '#a78bfa';
  if (t.includes('write')) return '#3ecf8e';
  if (t.includes('edit')) return '#f2a65a';
  if (t.includes('gate') || t.includes('audit')) return '#ff8a5c';
  if (t.includes('subagent')) return '#e879f9';
  if (t.includes('context7') || t.includes('search')) return '#38c6e8';
  return '#8aa4ff';
}

function renderStationDrawerContent(st) {
  const drawer = document.getElementById('stationDrawer');
  if (!drawer) return;

  drawer.style.setProperty('--c', st.color);

  const stationActsFromList = (currentActivities || []).filter(a => a.station === st.id);
  const stActs = (st.recent_activity && st.recent_activity.length > 0) ? st.recent_activity : stationActsFromList;
  const items = st.items || [];
  const selectedTaskId = selectedTaskIdByStation[st.id] || (items.length > 0 ? items[0].id : null);

  const activeItems = sortItemsNewestFirst(items.filter(it => it.status !== 'approved' && it.status !== 'done'));
  const completedItems = sortItemsNewestFirst(items.filter(it => it.status === 'approved' || it.status === 'done'));

  // --- Painel Tarefas ---
  let activeCardsHtml = '';
  if (activeItems.length > 0) {
    activeCardsHtml = activeItems.map(it => {
      const isSelected = it.id === selectedTaskId;
      const filePath = getIdeFilePath(it.artifact_path);
      const vsCodeUri = filePath ? (filePath.startsWith('/') ? `vscode://file${filePath}` : `vscode://file/${filePath}`) : 'vscode://';
      const cursorUri = filePath ? (filePath.startsWith('/') ? `cursor://file${filePath}` : `cursor://file/${filePath}`) : 'cursor://';
      const cliCmd = it.artifact_path ? `code ${it.artifact_path}` : (it.id ? `forge-sdd task view ${it.id}` : 'code .');
      const copyBtnId = `copy-btn-${it.id.replace(/[^a-zA-Z0-9_-]/g, '_')}`;

      let rules = [];
      if (it.metadata && it.metadata.rules) {
        rules = it.metadata.rules.split(',').map(r => r.trim()).filter(Boolean);
      }
      if (rules.length === 0) {
        rules = CANONICAL_RULES_BY_STATION[st.id] || ['ARQ-03', 'DUP-02'];
      }

      let handoffTarget = null;
      let handoffBtnLabel = null;
      if (st.id === 'pm-architect') {
        handoffTarget = 'planner';
        handoffBtnLabel = '🚀 Enviar para Planejador (03)';
      } else if (st.id === 'planner') {
        handoffTarget = 'builder';
        handoffBtnLabel = '🚀 Enviar para Builder (04)';
      } else if (st.id === 'builder') {
        handoffTarget = 'governance';
        handoffBtnLabel = '🚀 Enviar para Governança (05)';
      } else if (st.id === 'governance') {
        handoffTarget = 'lider';
        handoffBtnLabel = '🚀 Enviar para Líder (01)';
      }

      const statusInfo = formatTaskStatus(it.status);

      return `
        <div class="drawer-task-card ${isSelected ? 'selected' : ''}" onclick="selectTaskInDrawer('${escapeHtml(it.id)}')">
          <div class="drawer-task-header">
            <span class="drawer-task-id">${escapeHtml(it.id)}</span>
            <div style="display:flex;align-items:center;gap:6px;">
              <span class="chip" style="font-size:10px;padding:2px 8px;">tentativa #${it.attempt || 1}</span>
              <span class="chip" style="font-size:10px;padding:3px 9px;font-weight:700;color:${statusInfo.color};border-color:${statusInfo.border};background:${statusInfo.bg};">${escapeHtml(statusInfo.text)}</span>
            </div>
          </div>

          <div class="drawer-task-title">${escapeHtml(it.title || it.id)}</div>

          ${it.artifact_path ? `
            <div class="drawer-task-path" title="${escapeHtml(it.artifact_path)}">
              📄 ${escapeHtml(it.artifact_path)}
            </div>
          ` : ''}

          <div class="rules-strip" title="Pilares Constitucionais Mapeados">
            <span style="font-size:10px;color:var(--muted);font-weight:600;">Regras:</span>
            ${rules.map(r => `<span class="rule-chip" title="${escapeHtml(r)}">${escapeHtml(formatRuleBadge(r))}</span>`).join('')}
          </div>

          <div class="ide-buttons-row">
            <a href="${vsCodeUri}" class="btn-ide ide-link-btn" title="Abrir arquivo no editor VS Code nativo">
              Abrir no VS Code
            </a>
            <a href="${cursorUri}" class="btn-ide ide-link-btn cursor-btn" title="Abrir arquivo no editor Cursor">
              Cursor
            </a>
            <button id="${copyBtnId}" class="copy-btn ide-link-btn" onclick="event.stopPropagation(); copyCommand('${escapeHtml(cliCmd)}', '${copyBtnId}')" title="Copiar comando CLI para terminal local">
              Copiar CLI
            </button>
            <button class="inspect-spec-btn ide-link-btn" onclick="event.stopPropagation(); openSpecViewer('${escapeHtml(it.artifact_path || it.id)}')" title="Ver especificação técnica detalhada">
              Ver Spec
            </button>
            ${handoffTarget ? `
              <button class="btn-handoff ide-link-btn" data-from="${escapeHtml(st.id)}" data-to="${escapeHtml(handoffTarget)}" data-id="${escapeHtml(it.id)}" data-title="${escapeHtml(it.title || it.id)}" data-artifact="${escapeHtml(it.artifact_path || '')}" onclick="event.stopPropagation(); handleHandoffClick(this)" title="Transferir tarefa e artefato para a próxima estação">
                ${escapeHtml(handoffBtnLabel)}
              </button>
            ` : ''}
          </div>
        </div>
      `;
    }).join('');
  } else {
    activeCardsHtml = `
      <div style="background:var(--bg3);border:1px dashed var(--line);border-radius:8px;padding:16px;text-align:center;color:var(--muted);">
        <p style="font-size:12px;font-weight:600;color:var(--fg);margin-bottom:4px;">Nenhuma tarefa ativa ou pendente na fila</p>
        <p style="font-size:11px;">Todas as tarefas recebidas foram executadas ou aguardam novo handoff.</p>
      </div>
    `;
  }

  let completedCardsHtml = '';
  if (completedItems.length > 0) {
    completedCardsHtml = completedItems.map(it => {
      const isSelected = it.id === selectedTaskId;
      const filePath = getIdeFilePath(it.artifact_path);
      const vsCodeUri = filePath ? (filePath.startsWith('/') ? `vscode://file${filePath}` : `vscode://file/${filePath}`) : 'vscode://';
      const cursorUri = filePath ? (filePath.startsWith('/') ? `cursor://file${filePath}` : `cursor://file/${filePath}`) : 'cursor://';
      const cliCmd = it.artifact_path ? `code ${it.artifact_path}` : (it.id ? `forge-sdd task view ${it.id}` : 'code .');
      const copyBtnId = `copy-comp-btn-${it.id.replace(/[^a-zA-Z0-9_-]/g, '_')}`;
      const completedTime = it.updated_at ? new Date(it.updated_at).toLocaleString() : (it.created_at ? new Date(it.created_at).toLocaleString() : 'Concluído');
      const statusInfo = formatTaskStatus(it.status || 'approved');

      let compRules = [];
      if (it.metadata && it.metadata.rules) {
        compRules = it.metadata.rules.split(',').map(r => r.trim()).filter(Boolean);
      }
      if (compRules.length === 0) {
        compRules = CANONICAL_RULES_BY_STATION[st.id] || ['ARQ-03', 'DUP-02'];
      }

      let verdictBoxHtml = '';
      if (it.governance_verdict) {
        const gv = it.governance_verdict;
        const isAprovado = (gv.veredito || '').toLowerCase() === 'conforme' && (!gv.achados || gv.achados.length === 0);
        if (isAprovado) {
          verdictBoxHtml = `
            <div class="verdict-box conforme" style="border-left:3px solid var(--ok);background:rgba(62,207,142,0.06);padding:8px 12px;border-radius:6px;margin-top:8px;">
              <div style="display:flex;align-items:center;gap:6px;margin-bottom:4px;">
                <span class="chip" style="font-size:10px;font-weight:700;color:var(--ok);border-color:var(--ok);background:rgba(62,207,142,0.12);">✓ APROVADO PELA GOVERNANÇA</span>
              </div>
              <span style="font-size:11px;color:var(--fg);">${escapeHtml(gv.resumo || 'Auditado em conformidade com as regras constitucionais.')}</span>
            </div>
          `;
        } else {
          const achadosCount = (gv.achados && gv.achados.length) ? gv.achados.length : 1;
          const achadosDetails = (gv.achados && gv.achados.length > 0)
            ? gv.achados.map(a => `<div style="font-size:10.5px;color:var(--err);margin-top:2px;">• <b>${escapeHtml(formatRuleBadge(a.regra || ''))}:</b> ${escapeHtml(a.explicacao || a.verificacao || '')}${a.arquivo ? ` <i>(${escapeHtml(a.arquivo)}${a.linha ? `:${a.linha}` : ''})</i>` : ''}</div>`).join('')
            : '';
          verdictBoxHtml = `
            <div class="verdict-box falha" style="border-left:3px solid var(--err);background:rgba(255,92,92,0.06);padding:8px 12px;border-radius:6px;margin-top:8px;">
              <div style="display:flex;align-items:center;gap:6px;margin-bottom:4px;">
                <span class="chip" style="font-size:10px;font-weight:700;color:var(--err);border-color:var(--err);background:rgba(255,92,92,0.12);">✕ FALHA APONTADA</span>
                <span style="font-size:10px;color:var(--muted);">${achadosCount} achado(s)</span>
              </div>
              <div style="font-size:11px;color:var(--fg);font-weight:600;">${escapeHtml(gv.resumo || 'Auditoria detectou violação de regras constitucionais.')}</div>
              ${achadosDetails}
            </div>
          `;
        }
      } else {
        verdictBoxHtml = `
          <div class="verdict-box conforme" style="border-left:3px solid var(--ok);background:rgba(62,207,142,0.06);padding:8px 12px;border-radius:6px;margin-top:8px;">
            <div style="display:flex;align-items:center;gap:6px;margin-bottom:4px;">
              <span class="chip" style="font-size:10px;font-weight:700;color:var(--ok);border-color:var(--ok);background:rgba(62,207,142,0.12);">✓ APROVADO PELA GOVERNANÇA</span>
            </div>
            <span style="font-size:11px;color:var(--muted);">Executado e finalizado com sucesso pela estação ${escapeHtml(st.name)}.</span>
          </div>
        `;
      }

      return `
        <div class="drawer-task-card completed ${isSelected ? 'selected' : ''}" onclick="selectTaskInDrawer('${escapeHtml(it.id)}')">
          <div class="drawer-task-header">
            <span class="drawer-task-id">${escapeHtml(it.id)}</span>
            <div style="display:flex;align-items:center;gap:6px;">
              <span class="chip" style="font-size:10px;padding:2px 8px;color:var(--muted);">${escapeHtml(completedTime)}</span>
              <span class="chip" style="font-size:10px;padding:3px 9px;font-weight:700;color:${statusInfo.color};border-color:${statusInfo.border};background:${statusInfo.bg};">${escapeHtml(statusInfo.text)}</span>
            </div>
          </div>

          <div class="drawer-task-title">${escapeHtml(it.title || it.id)}</div>

          ${it.artifact_path ? `
            <div class="drawer-task-path" title="${escapeHtml(it.artifact_path)}">
              📄 ${escapeHtml(it.artifact_path)}
            </div>
          ` : ''}

          <div class="rules-strip" title="Pilares Constitucionais Mapeados">
            <span style="font-size:10px;color:var(--muted);font-weight:600;">Regras:</span>
            ${compRules.map(r => `<span class="rule-chip" title="${escapeHtml(r)}">${escapeHtml(formatRuleBadge(r))}</span>`).join('')}
          </div>

          ${verdictBoxHtml}

          <div class="ide-buttons-row">
            <a href="${vsCodeUri}" class="btn-ide ide-link-btn" title="Abrir arquivo no editor VS Code nativo">
              Abrir no VS Code
            </a>
            <a href="${cursorUri}" class="btn-ide ide-link-btn cursor-btn" title="Abrir arquivo no editor Cursor">
              Cursor
            </a>
            <button id="${copyBtnId}" class="copy-btn ide-link-btn" onclick="event.stopPropagation(); copyCommand('${escapeHtml(cliCmd)}', '${copyBtnId}')" title="Copiar comando CLI para terminal local">
              Copiar CLI
            </button>
            ${it.artifact_path ? `
              <button class="inspect-spec-btn ide-link-btn" onclick="event.stopPropagation(); openSpecViewer('${escapeHtml(it.artifact_path || it.id)}')" title="Ver especificação técnica detalhada">
                Ver Spec
              </button>
            ` : ''}
          </div>
        </div>
      `;
    }).join('');
  } else {
    completedCardsHtml = `
      <div style="background:var(--bg3);border:1px dashed var(--line);border-radius:8px;padding:16px;text-align:center;color:var(--muted);">
        <p style="font-size:12px;font-weight:500;color:var(--muted);">Nenhuma tarefa concluída no histórico desta estação.</p>
      </div>
    `;
  }

  const tasksHtml = `
    <div class="drawer-section-title">
      <span>⚡ Tarefas Ativas / Na Fila</span>
      <span class="chip mono">${activeItems.length}</span>
    </div>
    <div style="display:flex;flex-direction:column;gap:12px;margin-bottom:20px;">
      ${activeCardsHtml}
    </div>

    <div class="drawer-section-title">
      <span>📜 Histórico de Conclusões desta Estação</span>
      <span class="chip mono">${completedItems.length}</span>
    </div>
    <div style="display:flex;flex-direction:column;gap:12px;">
      ${completedCardsHtml}
    </div>
  `;

  // --- Painel Contratos ---
  const contractsHtml = `
    <div class="contract-section">
      <div class="contract-card">
        <span class="contract-label">Papel Canônico (.agents/roles/${escapeHtml(st.id)}.md)</span>
        <div class="contract-val">${escapeHtml(st.role)}</div>
        <p style="font-size:12px;color:var(--muted);margin-top:4px;">
          Responsabilidade canônica definida na esteira SDD das 5 estações.
        </p>
      </div>

      <div class="contract-card">
        <span class="contract-label">Recebe (Entrada)</span>
        <div class="contract-val" style="color:#79d4fd;">${escapeHtml(st.input_contract || 'Demanda inicial')}</div>
      </div>

      <div class="contract-card">
        <span class="contract-label">Entrega (Saída)</span>
        <div class="contract-val" style="color:var(--ok);">${escapeHtml(st.output_contract || 'Artefato validado')}</div>
      </div>

      <div class="contract-card">
        <span class="contract-label">Metadados de Fila</span>
        <div style="display:flex;gap:10px;margin-top:6px;flex-wrap:wrap;">
          <span class="chip mono">Fila: ${escapeHtml(st.id)}.json</span>
          <span class="chip mono">Versão: v${st.version}</span>
          <span class="chip mono">Itens: ${items.length} (${st.active_items || 0} ativos)</span>
        </div>
      </div>
    </div>
  `;

  // --- Painel Histórico de Ações (Telemetria) ---
  let telemetryHtml = '';
  if (stActs.length > 0) {
    telemetryHtml = `
      <div class="drawer-timeline">
        ${stActs.map(act => {
          const tColor = getToolColor(act.tool);
          const actTime = act.timestamp ? new Date(act.timestamp).toLocaleTimeString() : '';
          let statusBadge = '';
          if (act.status) {
            const isErr = act.status === 'error';
            const isRun = act.status === 'running' || act.status === 'working';
            const badgeColor = isErr ? 'var(--err)' : (isRun ? 'var(--warn)' : 'var(--ok)');
            statusBadge = `<span class="chip" style="font-size:9.5px;padding:1px 6px;color:${badgeColor};border-color:${badgeColor};">${escapeHtml(act.status)}</span>`;
          }

          return `
            <div class="drawer-timeline-item">
              <span class="drawer-timeline-dot" style="--tool-color:${tColor};"></span>
              <div class="drawer-timeline-card">
                <div class="drawer-timeline-header">
                  <span class="tool-chip" style="--c:${tColor};font-size:10px;padding:2px 8px;">
                    <span class="tool-chip-dot" style="background:${tColor};"></span>${escapeHtml(act.tool || 'cmd')}
                  </span>
                  <div style="display:flex;align-items:center;gap:6px;">
                    ${statusBadge}
                    <span class="drawer-timeline-time">${escapeHtml(actTime)}</span>
                  </div>
                </div>
                <div class="drawer-timeline-action">${escapeHtml(act.action || act.details || '')}</div>
                ${act.details && act.details !== act.action ? `
                  <div class="drawer-timeline-detail" title="${escapeHtml(act.details)}">${escapeHtml(act.details)}</div>
                ` : ''}
              </div>
            </div>
          `;
        }).join('')}
      </div>
    `;
  } else {
    const stationGuidance = {
      'lider': 'O Líder orquestra os ciclos de desenvolvimento, monitora handoffs, delega tarefas aos subagentes e aprova avanços de estágio.',
      'pm-architect': 'O PM & Arquiteto concebe especificações técnicas (Discovery, Criteria, C4), gerencia backlog e avalia reuso de código.',
      'planner': 'O Planejador decompõe especificações aprovadas em tarefas atômicas sequenciais (Plano) com critérios de aceite estritos.',
      'builder': 'O Builder implementa código e testes automatizados em contexto isolado, respeitando as regras constitucionais SDD.',
      'governance': 'A Governança audita em modo estrito somente-leitura (Camada B) a conformidade arquitetural, de segurança e integridade.'
    };
    telemetryHtml = `
      <div style="background:var(--bg3);border:1px dashed var(--line);border-radius:10px;padding:24px;text-align:center;color:var(--muted);">
        <p style="font-size:13.5px;font-weight:700;color:var(--fg);margin-bottom:8px;">Nenhuma ação registrada para ${escapeHtml(st.name)}</p>
        <div style="display:inline-block;text-align:left;background:rgba(0,0,0,0.25);border:1px solid var(--line);border-radius:8px;padding:12px 16px;margin:10px 0 16px;">
          <div style="font-size:11px;font-weight:700;color:var(--c, var(--accent));text-transform:uppercase;letter-spacing:0.05em;margin-bottom:4px;">🎯 Papel da Estação</div>
          <div style="font-size:12px;color:var(--fg);font-weight:600;margin-bottom:6px;">${escapeHtml(st.role)}</div>
          <div style="font-size:11.5px;color:var(--muted);line-height:1.45;">${escapeHtml(stationGuidance[st.id] || 'As atividades serão registradas automaticamente conforme os comandos forem emitidos.')}</div>
        </div>
        <p style="font-size:11px;color:var(--muted);">Aguardando execução de tarefas ou novo handoff em <code>sdd/.queue/stations/${escapeHtml(st.id)}.json</code>.</p>
      </div>
    `;
  }

  drawer.innerHTML = `
    <div class="drawer-header">
      <div class="drawer-title-group">
        <div class="drawer-badge-row">
          <span class="chip" style="color:${st.color};border-color:${st.color};padding:2px 8px;font-size:10px;">
            <i style="--o:${st.color}"></i>${escapeHtml(st.mascot_state.toUpperCase())}
          </span>
          <span class="chip mono" style="font-size:10px;padding:2px 8px;">Marca: ${escapeHtml(st.mark)}</span>
        </div>
        <div class="drawer-station-name">
          <span>${escapeHtml(st.name)}</span>
        </div>
        <div class="drawer-station-role">${escapeHtml(st.role)}</div>
      </div>
      <button class="drawer-close-btn" onclick="closeStationDrawer()" aria-label="Fechar painel da estação" title="Fechar (ESC)">✕</button>
    </div>

    <div class="drawer-tabs">
      <button class="drawer-tab ${activeDrawerTab === 'tasks' ? 'active' : ''}" onclick="switchDrawerTab('tasks')">
        Tarefas (${items.length})
      </button>
      <button class="drawer-tab ${activeDrawerTab === 'contracts' ? 'active' : ''}" onclick="switchDrawerTab('contracts')">
        Contratos &amp; Papel
      </button>
      <button class="drawer-tab ${activeDrawerTab === 'telemetry' ? 'active' : ''}" onclick="switchDrawerTab('telemetry')">
        Histórico de Ações (${stActs.length})
      </button>
    </div>

    <div class="drawer-body">
      <div class="drawer-pane ${activeDrawerTab === 'tasks' ? 'active' : ''}">
        ${tasksHtml}
      </div>
      <div class="drawer-pane ${activeDrawerTab === 'contracts' ? 'active' : ''}">
        ${contractsHtml}
      </div>
      <div class="drawer-pane ${activeDrawerTab === 'telemetry' ? 'active' : ''}">
        ${telemetryHtml}
      </div>
    </div>
  `;
}

// Renderiza a grade de estações e atualiza os DOMs
function applySnapshot(snap) {
  if (!snap) return;
  lastSnapshot = snap;

  if (snap.project_dir) {
    try {
      localStorage.setItem('studio_project_dir', snap.project_dir);
    } catch (e) {
      // quota or private mode
    }
  }

  document.getElementById('projectName').textContent = snap.project_name || 'forge-sdd';
  document.getElementById('projectVersion').textContent = 'v' + (snap.version || '3.2.1-beta');
  document.getElementById('lastUpdate').textContent = new Date(snap.updated_at).toLocaleTimeString();

  // Atualiza telemetria de atividades
  currentActivities = snap.recent_activity || [];
  renderConsoleLogs(currentActivities);

  // Handoff Bar
  const ho = snap.last_handoff;
  if (ho) {
    document.getElementById('hoFrom').textContent = ho.from_station;
    document.getElementById('hoTo').textContent = ho.to_station;
    const isSpecHandoff = ho.from_station === 'pm-architect' || ho.to_station === 'pm-architect' || (ho.gate && ho.gate.includes('FLX-01'));
    if (isSpecHandoff) {
      document.getElementById('hoPayload').innerHTML = `${escapeHtml(ho.payload)} • <a href="javascript:void(0)" onclick="openSpecViewer('sdd/discovery/discovery-63-studio-navegavel-aprovacao-spec.md')" style="color:var(--accent);text-decoration:underline;font-weight:700;">Ver Spec (FLX-01)</a>`;
    } else {
      document.getElementById('hoPayload').textContent = ho.payload;
    }
    document.getElementById('hoGate').innerHTML = `<i></i>${ho.gate}`;
  }

  // Grid das 5 estações
  const grid = document.getElementById('stationsGrid');

  snap.stations.forEach((st, idx) => {
    const isActive = st.id === snap.active_station;

    // Ferramentas ativas ou padrão da estação
    const defaultToolsByStation = {
      'lider': ['run_command', 'git', 'read_file', 'subagent'],
      'pm-architect': ['read_file', 'write_file', 'context7', 'search_web'],
      'planner': ['read_file', 'edit_file', 'gate', 'write_file'],
      'builder': ['edit_file', 'run_command', 'git', 'read_file'],
      'governance': ['gate', 'read_file', 'analyst', 'audit']
    };
    const tools = (st.active_tools && st.active_tools.length > 0) ? st.active_tools : (defaultToolsByStation[st.id] || []);

    let toolStripHtml = '';
    if (tools.length > 0) {
      toolStripHtml = `
        <div class="tool-strip">
          ${tools.map((t, tIdx) => {
            const isToolRunning = tIdx === 0 && (st.mascot_state === 'working' || st.mascot_state === 'thinking' || st.mascot_state === 'searching' || st.mascot_state === 'approval');
            return `
              <span class="tool-chip ${isToolRunning ? 'running' : ''}" style="--c:${st.color}">
                <span class="tool-chip-dot"></span>${escapeHtml(t)}
              </span>
            `;
          }).join('')}
        </div>
      `;
    }

    let actionRowHtml = '';
    if (st.current_action) {
      actionRowHtml = `
        <div class="executing-action-row" title="${escapeHtml(st.current_action)}">
          <span class="act-label">CMD:</span><b>${escapeHtml(st.current_action)}</b>
        </div>
      `;
    }

    // O QUE ESTÁ EXECUTANDO POR AGENTE (Destaque Principal Padronizado em Todos os Agentes)
    const execTask = st.executing_task;
    let execHtml = '';
    if (execTask && (execTask.status === 'building' || execTask.status === 'in_progress' || execTask.status === 'review_pending' || execTask.status === 'changes_requested' || execTask.status === 'queued' || execTask.status === 'pending_review' || execTask.status === 'approval' || execTask.status === 'working')) {
      const isRetrying = execTask.status === 'changes_requested';
      let shortLabel = 'EM EXECUÇÃO';
      if (execTask.status === 'queued') {
        shortLabel = 'NA FILA';
      } else if (execTask.status === 'review_pending' || execTask.status === 'pending_review' || execTask.status === 'approval') {
        shortLabel = 'AUDITORIA';
      } else if (execTask.status === 'changes_requested') {
        shortLabel = 'RETRABALHO';
      }

      execHtml = `
        <div class="executing-box active" style="--c:${isRetrying ? 'var(--err)' : st.color}">
          <div class="executing-header">
            <span class="pulse-dot"></span>
            <span class="executing-label">${shortLabel}</span>
            <span class="executing-id" title="${escapeHtml(execTask.id)}">${escapeHtml(execTask.id)}</span>
          </div>
          <div class="executing-title" title="${escapeHtml(execTask.title || st.current_focus)}">${escapeHtml(execTask.title || st.current_focus)}</div>
          ${actionRowHtml}
          ${toolStripHtml}
        </div>
      `;
    } else if (st.mascot_state === 'working' || st.mascot_state === 'thinking' || st.mascot_state === 'searching') {
      execHtml = `
        <div class="executing-box active" style="--c:${st.color}">
          <div class="executing-header">
            <span class="pulse-dot"></span>
            <span class="executing-label">ATIVIDADE VIVA</span>
          </div>
          <div class="executing-title" title="${escapeHtml(st.current_focus || 'Processando etapa do ciclo')}">${escapeHtml(st.current_focus || 'Processando etapa do ciclo')}</div>
          ${actionRowHtml}
          ${toolStripHtml}
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
          <div class="executing-title muted" title="${escapeHtml(st.current_focus || 'Fila ociosa — Pronto para novo handoff')}">${escapeHtml(st.current_focus || 'Fila ociosa — Pronto para novo handoff')}</div>
          ${actionRowHtml}
          ${toolStripHtml}
        </div>
      `;
    }

    // Itens na Fila (exibe EXCLUSIVAMENTE tarefas ativas/pendentes na frente do card)
    let itemsHtml = '';
    const rawItems = st.items || [];
    const activeTasks = sortItemsNewestFirst(rawItems.filter(it => it.status !== 'approved' && it.status !== 'done'));
    const completedTasks = sortItemsNewestFirst(rawItems.filter(it => it.status === 'approved' || it.status === 'done'));

    if (activeTasks.length > 0) {
      itemsHtml = activeTasks.map(it => {
        let shortStatus = 'NA FILA';
        let statusColor = 'var(--accent)';
        if (it.status === 'building' || it.status === 'working') {
          shortStatus = 'EXECUÇÃO';
          statusColor = 'var(--warn)';
        } else if (it.status === 'review_pending' || it.status === 'approval' || it.status === 'pending_review') {
          shortStatus = 'AUDITORIA';
          statusColor = '#ffd8a8';
        } else if (it.status === 'changes_requested') {
          shortStatus = 'RETRABALHO';
          statusColor = 'var(--err)';
        }

        return `
          <div class="queue-task-row">
            <span class="queue-task-name" title="${escapeHtml(it.title || it.id)}">${escapeHtml(it.id)}</span>
            <span class="chip" style="font-size:8.5px;padding:1px 5px;color:${statusColor};border-color:currentColor;flex-shrink:0;">${shortStatus}</span>
          </div>
        `;
      }).join('');
    } else if (completedTasks.length > 0) {
      itemsHtml = `
        <div style="font-size:11px;color:var(--ok);font-weight:600;padding:6px 0;display:flex;align-items:center;gap:6px;">
          <span>✓</span> <span>${completedTasks.length} concluídas</span>
        </div>
      `;
    } else {
      itemsHtml = `<div style="color:var(--muted);font-style:italic;font-size:11px;padding:4px 0;">Fila ociosa</div>`;
    }

    // Badge de Spec (FLX-01):
    // Na Estação 02 (PM & Arquiteto): exibido APENAS se houver tarefa com status ativo/pendente ('pending_review', 'approval', 'queued', 'in_progress') e artifact_path.
    // Na Estação 03 (Planejador): se houver spec ativa recebida por handoff (status !== 'approved' && status !== 'done' com artifact_path), exibir badge discreto.
    let specBadgeHtml = '';
    if (st.id === 'pm-architect') {
      const pendingStatuses = ['pending_review', 'approval', 'queued', 'in_progress'];
      const activeSpecTask = (st.items || []).find(it => it.artifact_path && pendingStatuses.includes(it.status));
      if (activeSpecTask) {
        specBadgeHtml = `
          <a href="javascript:void(0)" class="spec-link-btn" onclick="event.stopPropagation(); openSpecViewer('${escapeHtml(activeSpecTask.artifact_path)}')" style="color:#ffd8a8;font-size:10px;font-weight:600;text-decoration:none;background:rgba(242,166,90,0.15);padding:2px 6px;border-radius:4px;border:1px solid rgba(242,166,90,0.3);" title="Ver especificação técnica detalhada">📄 Spec (FLX-01)</a>
        `;
      }
    } else if (st.id === 'planner') {
      const activePlannerSpecTask = (st.items || []).find(it => it.artifact_path && it.status !== 'approved' && it.status !== 'done');
      if (activePlannerSpecTask) {
        specBadgeHtml = `
          <a href="javascript:void(0)" class="spec-link-btn" onclick="event.stopPropagation(); openSpecViewer('${escapeHtml(activePlannerSpecTask.artifact_path)}')" style="color:#c9a7ff;font-size:10px;font-weight:600;text-decoration:none;background:rgba(157,110,237,0.15);padding:2px 6px;border-radius:4px;border:1px solid rgba(157,110,237,0.3);" title="Ver especificação técnica recebida da Estação 02">📄 Spec</a>
        `;
      }
    }

    let card = document.getElementById(`station-card-${st.id}`);
    if (!card) {
      card = document.createElement('div');
      card.id = `station-card-${st.id}`;
      card.className = `station-card ${isActive ? 'active' : ''}`;
      card.style.setProperty('--c', st.color);
      card.setAttribute('tabindex', '0');
      card.setAttribute('role', 'button');
      card.setAttribute('aria-label', `Inspecionar Estação ${st.name}`);
      card.setAttribute('onclick', `openStationDrawer('${st.id}')`);
      card.onclick = () => openStationDrawer(st.id);
      card.onkeydown = (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          openStationDrawer(st.id);
        }
      };

      card.innerHTML = `
        <div class="station-top">
          <div class="mascot-stage" id="stage-${st.id}"></div>
          <div class="station-info">
            <h3>${st.name}</h3>
            <p class="role">${st.role}</p>
            <div id="station-badges-${st.id}" style="display:flex;align-items:center;gap:6px;margin-top:4px;flex-wrap:wrap;">
              <span class="chip" id="chip-${st.id}"><i style="--o:${st.color}"></i>${st.mascot_state}</span>
              ${specBadgeHtml}
            </div>
          </div>
        </div>

        <div id="exec-wrap-${st.id}">${execHtml}</div>

        <div class="io-strip">
          <div class="io-item"><span>Recebe</span><b>${st.input_contract}</b></div>
          <div class="io-item"><span>Entrega</span><b>${st.output_contract}</b></div>
        </div>

        <div class="queue-items-box">
          <div class="queue-header-row">
            <span class="mono" style="font-size:10px;color:var(--muted);font-weight:700;">FILA</span>
            <span id="queue-ver-${st.id}" class="chip mono" style="font-size:9px;padding:1px 5px;">v${st.version}</span>
          </div>
          <div id="queue-items-${st.id}">${itemsHtml}</div>
        </div>
      `;
      grid.appendChild(card);
    } else {
      card.className = `station-card ${isActive ? 'active' : ''}`;
      card.style.setProperty('--c', st.color);
      card.setAttribute('tabindex', '0');
      card.setAttribute('role', 'button');
      card.setAttribute('aria-label', `Inspecionar Estação ${st.name}`);
      card.setAttribute('onclick', `openStationDrawer('${st.id}')`);
      card.onclick = () => openStationDrawer(st.id);
      card.onkeydown = (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          openStationDrawer(st.id);
        }
      };
      const badgesContainer = document.getElementById(`station-badges-${st.id}`);
      if (badgesContainer) {
        badgesContainer.innerHTML = `
          <span class="chip" id="chip-${st.id}"><i style="--o:${st.color}"></i>${st.mascot_state}</span>
          ${specBadgeHtml}
        `;
      } else {
        const chip = document.getElementById(`chip-${st.id}`);
        if (chip) chip.innerHTML = `<i style="--o:${st.color}"></i>${st.mascot_state}`;
      }
      const execWrap = document.getElementById(`exec-wrap-${st.id}`);
      if (execWrap) execWrap.innerHTML = execHtml;
      const queueVer = document.getElementById(`queue-ver-${st.id}`);
      if (queueVer) queueVer.textContent = `v${st.version}`;
      const queueItems = document.getElementById(`queue-items-${st.id}`);
      if (queueItems) queueItems.innerHTML = itemsHtml;
    }

    // Inicialização ou atualização do Mascote (Three.js 3D com fallback SVG)
    const stage = document.getElementById(`stage-${st.id}`);
    if (stage) {
      let ctrl = mascotControllers[st.id];
      if (!ctrl) {
        try {
          ctrl = createForgeMascot3D(stage, st.color, st.mascot_state, st.mark);
        } catch (err) {
          console.warn('Erro ao criar Three.js mascot:', err);
        }
        if (ctrl && ctrl.domElement) {
          mascotControllers[st.id] = ctrl;
        } else {
          renderMascotFallbackSvg(stage, st.color, st.mark, st.mascot_state);
        }
      } else {
        if (ctrl.domElement && !stage.contains(ctrl.domElement)) {
          stage.innerHTML = '';
          stage.appendChild(ctrl.domElement);
        }
        if (ctrl.setState) {
          ctrl.setState(st.mascot_state, st.color);
          ctrl.triggerHop();
        }
      }

      if (stage.children.length === 0) {
        renderMascotFallbackSvg(stage, st.color, st.mark, st.mascot_state);
      }
    }
  });

  // Governança e Pilares Constitucionais Dinâmicos
  const gVerdict = snap.governance_verdict;
  const govCard = document.getElementById('govFindingCard');
  const pillarsGrid = document.getElementById('govPillarsGrid');

  const failedPillarKeys = new Set();
  const hasFindings = gVerdict && Array.isArray(gVerdict.achados) && gVerdict.achados.length > 0;
  const isViolacao = gVerdict && ((gVerdict.veredito || '').toLowerCase() === 'violacao' || hasFindings);

  if (isViolacao) {
    if (hasFindings) {
      gVerdict.achados.forEach(a => {
        if (a.regra) {
          const prefix = a.regra.split('-')[0].toUpperCase().trim();
          failedPillarKeys.add(prefix);
        }
      });
    }
    if (failedPillarKeys.size === 0) {
      failedPillarKeys.add('GOV');
    }

    if (govCard) {
      govCard.className = 'finding-card violacao';
      let findingsHtml = '';
      if (hasFindings) {
        findingsHtml = gVerdict.achados.map(a => {
          const ruleLabel = formatRuleBadge(a.regra || 'GOV-01');
          const loc = a.arquivo ? `${escapeHtml(a.arquivo)}${a.linha ? `:${a.linha}` : ''}` : '';
          return `
            <div class="finding-item falha">
              <div class="finding-item-header">
                <span class="finding-item-rule">✕ ${escapeHtml(ruleLabel)} ${a.regra ? `(${escapeHtml(a.regra)})` : ''}</span>
                ${loc ? `<span class="finding-item-loc mono">📄 ${loc}</span>` : ''}
              </div>
              <div class="finding-item-desc">${escapeHtml(a.explicacao || a.verificacao || a.evidencia || 'Violação de conformidade detectada')}</div>
              ${a.correcao ? `<div class="finding-item-fix"><b>Instrução de Correção:</b> ${escapeHtml(a.correcao)}</div>` : ''}
            </div>
          `;
        }).join('');
      } else {
        findingsHtml = `
          <div class="finding-item falha">
            <div class="finding-item-desc">${escapeHtml(gVerdict.resumo || 'Falha detectada durante a auditoria da Camada B.')}</div>
          </div>
        `;
      }

      govCard.innerHTML = `
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px;">
          <b style="color:var(--err);font-size:13px;display:flex;align-items:center;gap:6px;">
            <span>✕</span> FALHA DETECTADA NA AUDITORIA
          </b>
          <span class="chip" style="color:var(--err);border-color:var(--err);background:rgba(255,92,92,0.1);font-size:10px;">
            ${(gVerdict.achados && gVerdict.achados.length) ? `${gVerdict.achados.length} achado(s)` : 'VIOLAÇÃO DETECTADA'}
          </span>
        </div>
        <p style="margin-top:2px;color:var(--fg);font-size:12px;">${escapeHtml(gVerdict.resumo || 'Foram encontradas não-conformidades constitucionais que impedem o avanço.')}</p>
        <div style="margin-top:8px;">${findingsHtml}</div>
      `;
    }
  } else if (gVerdict && (gVerdict.veredito || '').toLowerCase() === 'conforme') {
    if (govCard) {
      govCard.className = 'finding-card conforme';
      govCard.innerHTML = `
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px;">
          <b style="color:var(--ok);font-size:13px;display:flex;align-items:center;gap:6px;">
            <span>✓</span> AUDITORIA APROVADA (CONFORME)
          </b>
          <span class="chip" style="color:var(--ok);border-color:var(--ok);background:rgba(62,207,142,0.1);font-size:10px;">
            100% CONFORME
          </span>
        </div>
        <p style="margin-top:2px;color:var(--fg);font-size:12px;">${escapeHtml(gVerdict.resumo || 'Todos os pilares constitucionais e contratos entre estações foram verificados e aprovados.')}</p>
        <div class="finding-item aprovado" style="margin-top:8px;">
          <div style="font-size:11.5px;color:var(--fg);">
            ✓ Nenhuma violação constitucional detectada. Portão de Governança Camadas A & B liberado com sucesso.
          </div>
        </div>
      `;
    }
  } else {
    // Estado inicial / ocioso / aguardando
    if (govCard) {
      govCard.className = 'finding-card conforme';
      govCard.innerHTML = `
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px;">
          <b style="color:var(--ok);font-size:13px;display:flex;align-items:center;gap:6px;">
            <span>✓</span> PORTÃO DE GOVERNANÇA ATIVO (CAMADAS A & B)
          </b>
          <span class="chip" style="font-size:10px;color:var(--muted);">STANDBY</span>
        </div>
        <p style="margin-top:2px;color:var(--muted);font-size:12px;">Aguardando submissão de artefatos para inspeção e auditoria independente em modo Read-Only.</p>
      `;
    }
  }

  // Renderiza os Pilares Constitucionais Dinâmicos
  if (pillarsGrid) {
    pillarsGrid.innerHTML = CONSTITUTIONAL_PILLARS.map(p => {
      const isFailed = failedPillarKeys.has(p.key);
      const chipClass = isFailed ? 'falha' : 'aprovado';
      const chipText = isFailed ? '✕ FALHOU' : '✓ APROVADO';
      return `
        <div class="gov-pillar-card ${isFailed ? 'falha' : 'aprovado'}">
          <div class="gov-pillar-header">
            <div class="gov-pillar-title-wrap">
              <span class="gov-pillar-icon">${p.icon}</span>
              <div>
                <div class="gov-pillar-name">${escapeHtml(p.title)}</div>
                <div class="gov-pillar-key mono">${escapeHtml(p.key)}</div>
              </div>
            </div>
            <span class="gov-pillar-chip ${chipClass}">${chipText}</span>
          </div>
          <p class="gov-pillar-desc">${escapeHtml(p.desc)}</p>
        </div>
      `;
    }).join('');
  }

  // Atualização reativa suave do Drawer de Inspeção se estiver aberto (feat-63-02)
  if (activeDrawerStationId && snap.stations) {
    const activeSt = snap.stations.find(s => s.id === activeDrawerStationId);
    if (activeSt) {
      renderStationDrawerContent(activeSt);
    }
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
    payload: 'Demanda Estruturada → Estação 02 (PM & Arquiteto)',
    tools: ['run_command', 'git', 'read_file', 'subagent'],
    action: 'git status --porcelain',
    details: 'Verificando integridade da árvore git e sincronizando ponteiros'
  },
  {
    station: 'pm-architect',
    mascot_state: 'thinking',
    executing: { id: 'discovery-60', title: 'Elaborando especificação técnica, diagramas C4 e Reuse Check', status: 'building' },
    gate: 'Gate 2 — Revisão Humana de Spec (FLX-01)',
    payload: 'discovery.md + criteria.md → Estação 03',
    tools: ['read_file', 'write_file', 'context7', 'search_web'],
    action: 'read_file sdd/features/feat-60-01-scaffold.md',
    details: 'Auditoria de critérios de aceitação e diagramas arquiteturais C4'
  },
  {
    station: 'planner',
    mascot_state: 'working',
    executing: { id: 'plan-60', title: 'Fatiando spec em tarefas P/M/G e mapeando rules por task', status: 'building' },
    gate: 'Gate 3 — Plano Validado e Enfileirado',
    payload: 'plan.md + queue.json → Estação 04 (Builders)',
    tools: ['read_file', 'edit_file', 'gate', 'write_file'],
    action: 'edit_file sdd/features/plan.md',
    details: 'Associando pilares ARQ-03, ARQ-04 e DUP-05 às sub-tarefas'
  },
  {
    station: 'builder',
    mascot_state: 'working',
    executing: { id: 'feat-60-03', title: 'Implementando código Go e suite de testes com detecção de corrida (-race)', status: 'building' },
    gate: 'Gate 4 — Testes Locais e Camada A (PASS)',
    payload: 'Código + Testes Conforme → Estação 05 (Governança)',
    tools: ['edit_file', 'run_command', 'git', 'read_file'],
    action: 'run_command go test -race -v ./internal/...',
    details: 'Executando testes unitários determinísticos com detecção de concorrência'
  },
  {
    station: 'governance',
    mascot_state: 'approval',
    executing: { id: 'audit-60-03', title: 'Auditoria Read-Only contra Constituição e emissão de laudo JSON', status: 'review_pending' },
    gate: 'Gate 5 — Laudo Camada B Aprovado',
    payload: 'Laudo Conforme → Estação 01 (Líder / PR)',
    tools: ['gate', 'read_file', 'analyst', 'audit'],
    action: 'gate audit-diff --strict',
    details: 'Camada B Read-Only: 7 pilares verificados, 0 violações'
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
        s.active_tools = simStep.tools;
        s.current_action = `${simStep.tools[0]}: ${simStep.action}`;
      } else {
        s.mascot_state = (i < currentDemoIdx) ? 'finished' : 'idle';
        s.executing_task = null;
        s.current_action = '';
      }
    });

    cloned.last_handoff = {
      from_station: simStep.station,
      to_station: stations[(currentDemoIdx + 1) % stations.length].id,
      payload: simStep.payload,
      gate: simStep.gate,
      is_loop: false
    };

    // Cria evento no histórico recente de telemetria
    const nowIso = new Date().toISOString();
    const newAct = {
      timestamp: nowIso,
      station: simStep.station,
      tool: simStep.tools[0],
      action: simStep.action,
      details: simStep.details,
      status: (simStep.mascot_state === 'approval' || simStep.mascot_state === 'finished') ? 'done' : 'running'
    };
    cloned.recent_activity = [newAct, ...(cloned.recent_activity || [])].slice(0, 50);

    if (simStep.station === 'governance') {
      cloned.governance_verdict = {
        veredito: 'conforme',
        resumo: 'Auditoria estrita concluída com 100% de conformidade com os 7 pilares constitucionais.',
        achados: []
      };
    }

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
