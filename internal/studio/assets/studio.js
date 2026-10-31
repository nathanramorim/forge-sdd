// Assets embarcados do Forge-SDD Studio
// Trata o motor dos 5 mascotes, sincronização SSE e chaveamento LIVE/DEMO.

const STATE_DEFS = {
  idle:      { orb: '#4f8ef7', eyes: 'cap' },
  working:   { orb: '#f2a65a', eyes: 'round' },
  thinking:  { orb: '#c9a7ff', eyes: 'round' },
  searching: { orb: '#38c6e8', eyes: 'round' },
  approval:  { orb: '#ff8a5c', eyes: 'cap' },
  finished:  { orb: '#3ecf8e', eyes: 'happy' },
  error:     { orb: '#f2677a', eyes: 'x' }
};

let currentMode = 'LIVE'; // LIVE | DEMO
let demoInterval = null;
let currentDemoIdx = 0;
let lastSnapshot = null;

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
  document.getElementById('projectVersion').textContent = 'v' + (snap.version || '3.2.0-beta');
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

    let itemsHtml = '';
    if (st.items && st.items.length > 0) {
      itemsHtml = st.items.map(it => `
        <div class="queue-task-row">
          <span style="color:var(--fg);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:140px;">${it.id}</span>
          <span class="chip" style="font-size:9.5px;padding:1px 6px;">${it.status}</span>
        </div>
      `).join('');
    } else {
      itemsHtml = `<div style="color:var(--muted);font-style:italic;">Nenhum item na fila</div>`;
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
    const stage = card.querySelector(`#stage-${st.id}`);
    renderMascotFallbackSvg(stage, st.color, st.mark, st.mascot_state);
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
    // Fallback para polling HTTP caso SSE caia
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

function runDemoSimulation() {
  if (demoInterval) clearInterval(demoInterval);
  demoInterval = setInterval(() => {
    if (!lastSnapshot) return;
    const stations = lastSnapshot.stations;
    currentDemoIdx = (currentDemoIdx + 1) % stations.length;
    const cloned = JSON.parse(JSON.stringify(lastSnapshot));
    cloned.active_station = stations[currentDemoIdx].id;
    cloned.active_station_idx = currentDemoIdx;
    cloned.stations.forEach((s, i) => {
      s.mascot_state = (i === currentDemoIdx) ? 'working' : 'finished';
    });
    cloned.last_handoff = {
      from_station: stations[currentDemoIdx].id,
      to_station: stations[(currentDemoIdx + 1) % stations.length].id,
      payload: stations[currentDemoIdx].output_contract,
      gate: 'Portão Validado (Simulação)',
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
