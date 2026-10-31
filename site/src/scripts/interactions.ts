import { STATIONS, STATE_DEFS, type MascotState } from '../data/stations';
import { EMU_STEPS } from '../data/emulator';
import { GOV_SCENARIOS } from '../data/governance';
import { createForgeMascot3D, type MascotController } from './mascot-3d';

// Global pointer coords for all mascots
let pointerX = 0;
let pointerY = 0;
window.addEventListener('pointermove', (e) => {
  pointerX = (e.clientX / window.innerWidth) * 2 - 1;
  pointerY = (e.clientY / window.innerHeight) * 2 - 1;
}, { passive: true });

/* ==========================================================================
   1. HERO COPY BUTTON
   ========================================================================== */
const copyHeroBtn = document.getElementById('copyHeroBtn');
if (copyHeroBtn) {
  copyHeroBtn.addEventListener('click', (e) => {
    navigator.clipboard?.writeText('npx @nathanramorim/forge-sdd@latest init');
    const target = e.target as HTMLElement;
    target.textContent = 'Copiado!';
    setTimeout(() => { target.textContent = 'Copiar'; }, 1800);
  });
}

/* ==========================================================================
   2. PROCESS SECTION CONTROLLER (ESTAÇÕES)
   ========================================================================== */
let activeStationIdx = 0;
let autoCycleEnabled = true;
let cycleProgress = 0;
const mascotControllers: (MascotController | null)[] = [null, null, null, null, null];

function updateHandoffBar(fromIdx: number) {
  const st = STATIONS[fromIdx];
  const toIdx = (fromIdx + 1) % STATIONS.length;
  const nextSt = STATIONS[toIdx];

  const hoFrom = document.getElementById('hoFrom');
  const hoTo = document.getElementById('hoTo');
  const hoPayload = document.getElementById('hoPayload');
  const hoGateChip = document.getElementById('hoGateChip');
  const hoLoopNote = document.getElementById('hoLoopNote');
  const hoPacket = document.getElementById('hoPacket');

  if (hoFrom) hoFrom.textContent = st.shortName;
  if (hoTo) hoTo.textContent = nextSt.shortName;
  if (hoPayload) hoPayload.textContent = st.handoffPayload;
  if (hoGateChip) {
    hoGateChip.innerHTML = `<i style="--o:${st.color}"></i>${st.handoffGate}`;
  }
  if (hoLoopNote) hoLoopNote.textContent = st.handoffNote;
  if (hoPacket) hoPacket.style.background = st.color;
}

function setActiveProcessStep(idx: number, isManual = false) {
  if (idx < 0 || idx >= STATIONS.length) return;
  activeStationIdx = idx;
  cycleProgress = 0;

  if (isManual) {
    autoCycleEnabled = false;
    const toggleBtn = document.getElementById('toggleCycleBtn');
    if (toggleBtn) toggleBtn.textContent = 'Retomar fluxo';
  }

  // Update track nodes
  document.querySelectorAll('.step-node').forEach((node, i) => {
    node.classList.toggle('active', i === idx);
    node.classList.toggle('done', i < idx);
  });

  // Update cards
  document.querySelectorAll('.agent-card').forEach((card, i) => {
    card.classList.toggle('active', i === idx);
    const prog = card.querySelector('.agent-progress') as HTMLElement | null;
    if (prog) prog.style.width = '0%';
  });

  updateHandoffBar(idx);

  const ctrl = mascotControllers[idx];
  if (ctrl) ctrl.triggerHop();
}

function setAgentState(agentIdx: number, stateKey: MascotState) {
  const card = document.getElementById(`agent-card-${agentIdx}`);
  if (!card) return;
  const st = STATIONS[agentIdx];
  const stDef = STATE_DEFS[stateKey];
  if (!stDef) return;

  const phrase = st.states[stateKey] || '';
  const phraseEl = card.querySelector('.agent-phrase');
  if (phraseEl) phraseEl.textContent = phrase;

  const chipEl = card.querySelector('.agent-chip');
  if (chipEl) chipEl.innerHTML = `<i style="--o:${stDef.cssOrb}"></i>${stDef.label}`;

  card.querySelectorAll('.state-pill').forEach((p) => {
    p.classList.toggle('on', (p as HTMLElement).dataset.state === stateKey);
  });

  const ctrl = mascotControllers[agentIdx];
  if (ctrl) {
    ctrl.setState(stateKey);
    ctrl.triggerHop();
  }
}

// Hook step clicks
document.querySelectorAll('.step-node').forEach((node) => {
  node.addEventListener('click', () => {
    const idx = parseInt((node as HTMLElement).dataset.idx || '0', 10);
    setActiveProcessStep(idx, true);
  });
});

// Hook card clicks & pills
document.querySelectorAll('.agent-card').forEach((card) => {
  card.addEventListener('click', (e) => {
    const target = e.target as HTMLElement;
    if (target.classList.contains('state-pill')) {
      const stateKey = target.dataset.state as MascotState;
      const agentIdx = parseInt(target.dataset.agent || '0', 10);
      setAgentState(agentIdx, stateKey);
      return;
    }
    const idx = parseInt((card as HTMLElement).dataset.idx || '0', 10);
    setActiveProcessStep(idx, true);
  });
});

const prevStepBtn = document.getElementById('prevStepBtn');
if (prevStepBtn) {
  prevStepBtn.addEventListener('click', () => {
    setActiveProcessStep((activeStationIdx - 1 + STATIONS.length) % STATIONS.length, true);
  });
}

const nextStepBtn = document.getElementById('nextStepBtn');
if (nextStepBtn) {
  nextStepBtn.addEventListener('click', () => {
    setActiveProcessStep((activeStationIdx + 1) % STATIONS.length, true);
  });
}

const toggleCycleBtn = document.getElementById('toggleCycleBtn');
if (toggleCycleBtn) {
  toggleCycleBtn.addEventListener('click', () => {
    autoCycleEnabled = !autoCycleEnabled;
    toggleCycleBtn.textContent = autoCycleEnabled ? 'Pausar fluxo' : 'Retomar fluxo';
  });
}

/* ==========================================================================
   3. EMULATOR SECTION CONTROLLER
   ========================================================================== */
let currentEmuIdx = 0;
let emuMascotCtrl: MascotController | null = null;

function renderEmuStep(idx: number) {
  currentEmuIdx = idx;
  const st = EMU_STEPS[idx];

  document.querySelectorAll('.emu-cmd-btn').forEach((b, i) => {
    b.classList.toggle('active', i === idx);
  });

  const emuLog = document.getElementById('emuLog');
  if (emuLog) emuLog.innerHTML = st.logHtml;

  const emuStepCounter = document.getElementById('emuStepCounter');
  if (emuStepCounter) emuStepCounter.textContent = `Passo ${idx + 1} de ${EMU_STEPS.length}`;

  const emuStationLabel = document.getElementById('emuStationLabel');
  if (emuStationLabel) emuStationLabel.textContent = st.stationLabel;

  const emuAgentName = document.getElementById('emuAgentName');
  if (emuAgentName) emuAgentName.textContent = st.agentName;

  const emuAgentSpeech = document.getElementById('emuAgentSpeech');
  if (emuAgentSpeech) emuAgentSpeech.textContent = st.speech;

  const emuStateChip = document.getElementById('emuStateChip');
  if (emuStateChip) emuStateChip.innerHTML = `<i style="--o:${st.chipColor}"></i>${st.chipText}`;

  const fileListEl = document.getElementById('emuFileList');
  if (fileListEl) {
    fileListEl.innerHTML = st.files.map(([name, badge, color]) => `
      <div class="emu-file-item">
        <span>${name}</span>
        <span class="chip" style="padding:2px 8px;font-size:10px"><i style="--o:${color}"></i>${badge}</span>
      </div>
    `).join('');
  }

  const slot = document.getElementById('emuActionSlot');
  if (slot) {
    const btnClass = st.actionBtn.style ? `btn ${st.actionBtn.style} sm` : 'btn ok sm';
    slot.innerHTML = `<button type="button" class="${btnClass}" id="emuDynamicBtn">${st.actionBtn.label}</button>`;
    document.getElementById('emuDynamicBtn')?.addEventListener('click', () => {
      renderEmuStep(st.actionBtn.nextStep);
    });
  }

  if (emuMascotCtrl) {
    emuMascotCtrl.setState(st.agentState, st.agentHex);
    emuMascotCtrl.triggerHop();
  }
}

document.querySelectorAll('.emu-cmd-btn').forEach((btn) => {
  btn.addEventListener('click', () => {
    const emuIdx = parseInt((btn as HTMLElement).dataset.emu || '0', 10);
    renderEmuStep(emuIdx);
  });
});

const emuNextStepBtn = document.getElementById('emuNextStepBtn');
if (emuNextStepBtn) {
  emuNextStepBtn.addEventListener('click', () => {
    renderEmuStep((currentEmuIdx + 1) % EMU_STEPS.length);
  });
}

/* ==========================================================================
   4. GOVERNANCE SECTION CONTROLLER
   ========================================================================== */
let currentGovIdx = 0;
let govAutoCycle = true;
let govTimer = 0;
let govMascotCtrl: MascotController | null = null;

function selectGov(idx: number) {
  currentGovIdx = idx;
  govTimer = 0;
  const sc = GOV_SCENARIOS[idx];

  document.querySelectorAll('.gov-tab').forEach((t, i) => {
    t.classList.toggle('active', i === idx);
  });

  const insp = document.getElementById('govInspector');
  if (insp) insp.style.setProperty('--k', sc.k);

  const govStateChip = document.getElementById('govStateChip');
  if (govStateChip) govStateChip.innerHTML = `<i style="--o:${sc.chipColor}"></i>${sc.chipLabel}`;

  const govRuleCode = document.getElementById('govRuleCode');
  if (govRuleCode) govRuleCode.textContent = sc.ruleCode;

  const govInspectorTitle = document.getElementById('govInspectorTitle');
  if (govInspectorTitle) govInspectorTitle.textContent = sc.title;

  const govInspectorSub = document.getElementById('govInspectorSub');
  if (govInspectorSub) govInspectorSub.textContent = sc.sub;

  const govTerminal = document.getElementById('govTerminal');
  if (govTerminal) govTerminal.innerHTML = sc.termHtml;

  if (govMascotCtrl) {
    govMascotCtrl.setState(sc.state);
    govMascotCtrl.triggerHop();
  }
}

document.querySelectorAll('.gov-tab').forEach((btn) => {
  btn.addEventListener('click', () => {
    govAutoCycle = false;
    const toggleGovCycleBtn = document.getElementById('toggleGovCycleBtn');
    if (toggleGovCycleBtn) toggleGovCycleBtn.textContent = 'Retomar simulação';
    const govIdx = parseInt((btn as HTMLElement).dataset.gov || '0', 10);
    selectGov(govIdx);
  });
});

const govBtnFix = document.getElementById('govBtnFix');
if (govBtnFix) {
  govBtnFix.addEventListener('click', () => {
    govAutoCycle = false;
    const toggleGovCycleBtn = document.getElementById('toggleGovCycleBtn');
    if (toggleGovCycleBtn) toggleGovCycleBtn.textContent = 'Retomar simulação';
    const sc = GOV_SCENARIOS[currentGovIdx];

    const insp = document.getElementById('govInspector');
    if (insp) insp.style.setProperty('--k', 'var(--ok)');

    const govStateChip = document.getElementById('govStateChip');
    if (govStateChip) govStateChip.innerHTML = `<i style="--o:var(--ok)"></i>Corrigido &amp; Aprovado`;

    const govInspectorTitle = document.getElementById('govInspectorTitle');
    if (govInspectorTitle) govInspectorTitle.textContent = sc.fixTitle;

    const govTerminal = document.getElementById('govTerminal');
    if (govTerminal) govTerminal.innerHTML = sc.fixTerm;

    if (govMascotCtrl) {
      govMascotCtrl.setState('finished');
      govMascotCtrl.triggerHop();
    }
  });
}

const govBtnNext = document.getElementById('govBtnNext');
if (govBtnNext) {
  govBtnNext.addEventListener('click', () => {
    selectGov((currentGovIdx + 1) % GOV_SCENARIOS.length);
  });
}

const toggleGovCycleBtn = document.getElementById('toggleGovCycleBtn');
if (toggleGovCycleBtn) {
  toggleGovCycleBtn.addEventListener('click', () => {
    govAutoCycle = !govAutoCycle;
    toggleGovCycleBtn.textContent = govAutoCycle ? 'Pausar simulação' : 'Retomar simulação';
  });
}

/* ==========================================================================
   5. SCROLLSPY
   ========================================================================== */
const sections = ['estacoes', 'emulador', 'instalacao', 'governanca', 'evolucao'];
const navLinks = document.querySelectorAll('nav a.nav-sec');
const spyObserver = new IntersectionObserver((entries) => {
  for (const entry of entries) {
    if (entry.isIntersecting) {
      const id = entry.target.id;
      navLinks.forEach((link) => {
        link.classList.toggle('active', link.getAttribute('href') === `#${id}`);
      });
    }
  }
}, { rootMargin: '-20% 0px -70% 0px', threshold: 0 });

sections.forEach((id) => {
  const el = document.getElementById(id);
  if (el) spyObserver.observe(el);
});

/* ==========================================================================
   6. LAZY INITIALIZATION OF 3D MASCOTS (requestIdleCallback)
   ========================================================================== */
function init3DMascots() {
  // 1. Process section mascots
  document.querySelectorAll('.avatar-stage').forEach((mount) => {
    const idx = parseInt((mount as HTMLElement).dataset.idx || '0', 10);
    const st = STATIONS[idx];
    if (st) {
      mascotControllers[idx] = createForgeMascot3D(
        mount as HTMLElement,
        st.hex,
        st.defaultState,
        st.mark
      );
    }
  });

  // 2. Emulator mascot
  const emuMount = document.getElementById('emuAvatarMount');
  if (emuMount) {
    emuMascotCtrl = createForgeMascot3D(emuMount, 0xf2a65a, 'finished', 'ring');
  }

  // 3. Governance mascot
  const govMount = document.getElementById('govAvatarMount');
  if (govMount) {
    govMascotCtrl = createForgeMascot3D(govMount, 0xff8a5c, 'error', 'tri');
  }

  // Animation Loop
  let lastTime = performance.now();
  function animate(now: number) {
    const dt = (now - lastTime) / 1000;
    lastTime = now;
    const t = now * 0.001;

    // Process auto cycle (4.8s)
    if (autoCycleEnabled && !document.hidden) {
      cycleProgress += dt;
      const card = document.getElementById(`agent-card-${activeStationIdx}`);
      const prog = card?.querySelector('.agent-progress') as HTMLElement | null;
      if (prog) {
        prog.style.width = `${Math.min(100, (cycleProgress / 4.8) * 100)}%`;
      }
      if (cycleProgress >= 4.8) {
        setActiveProcessStep((activeStationIdx + 1) % STATIONS.length, false);
      }
    }

    // Governance auto cycle (6.5s)
    if (govAutoCycle && !document.hidden) {
      govTimer += dt;
      if (govTimer >= 6.5) {
        selectGov((currentGovIdx + 1) % GOV_SCENARIOS.length);
      }
    }

    // Update all 3D mascots
    for (const ctrl of mascotControllers) {
      ctrl?.update(t, pointerX, pointerY);
    }
    emuMascotCtrl?.update(t, pointerX, pointerY);
    govMascotCtrl?.update(t, pointerX, pointerY);

    requestAnimationFrame(animate);
  }

  requestAnimationFrame(animate);
}

// Kick off when browser is idle to ensure zero TBT
if ('requestIdleCallback' in window) {
  (window as any).requestIdleCallback(() => init3DMascots(), { timeout: 1500 });
} else {
  setTimeout(() => init3DMascots(), 200);
}
