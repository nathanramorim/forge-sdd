import fs from "node:fs";
import path from "node:path";

const SAMPLE_RATE = 48000;
const DURATION = 24.0;
const NUM_SAMPLES = Math.floor(SAMPLE_RATE * DURATION);

const left = new Float32Array(NUM_SAMPLES);
const right = new Float32Array(NUM_SAMPLES);

const midiToFreq = (m) => 440 * Math.pow(2, (m - 69) / 12);

function envADSR(t, dur, a = 0.02, d = 0.15, s = 0.5, r = 0.2) {
  if (t < 0 || t >= dur) return 0;
  if (t < a) return t / a;
  if (t < a + d) return 1 - (1 - s) * ((t - a) / d);
  if (t < dur - r) return s;
  return s * Math.max(0, 1 - (t - (dur - r)) / r);
}

function addWarmPadNote(startS, durS, midi, amp, pan = 0) {
  const f = midiToFreq(midi);
  const i0 = Math.max(0, Math.floor(startS * SAMPLE_RATE));
  const i1 = Math.min(NUM_SAMPLES, Math.floor((startS + durS) * SAMPLE_RATE));
  const lGain = Math.cos(((pan + 1) * Math.PI) / 4);
  const rGain = Math.sin(((pan + 1) * Math.PI) / 4);
  for (let i = i0; i < i1; i++) {
    const t = i / SAMPLE_RATE - startS;
    const e = envADSR(t, durS, Math.min(0.45, durS * 0.25), 0.3, 0.75, Math.min(0.6, durS * 0.3));
    const sigL =
      Math.sin(2 * Math.PI * (f * 0.999) * t) * 0.7 +
      Math.sin(2 * Math.PI * (f * 2.001) * t) * 0.2 +
      Math.sin(2 * Math.PI * (f * 3.0) * t) * 0.06;
    const sigR =
      Math.sin(2 * Math.PI * (f * 1.001) * t) * 0.7 +
      Math.sin(2 * Math.PI * (f * 1.999) * t) * 0.2 +
      Math.sin(2 * Math.PI * (f * 3.0) * t) * 0.06;
    left[i] += sigL * e * amp * lGain;
    right[i] += sigR * e * amp * rGain;
  }
}

function addPluck(startS, durS, midi, amp, pan = 0) {
  const f = midiToFreq(midi);
  const i0 = Math.max(0, Math.floor(startS * SAMPLE_RATE));
  const i1 = Math.min(NUM_SAMPLES, Math.floor((startS + durS) * SAMPLE_RATE));
  const lGain = Math.cos(((pan + 1) * Math.PI) / 4);
  const rGain = Math.sin(((pan + 1) * Math.PI) / 4);
  for (let i = i0; i < i1; i++) {
    const t = i / SAMPLE_RATE - startS;
    const attack = Math.min(1, t / 0.008);
    const decay = Math.exp(-t * 5.5);
    const overtoneDecay = Math.exp(-t * 16.0);
    const sig =
      Math.sin(2 * Math.PI * f * t) * decay +
      Math.sin(2 * Math.PI * f * 2 * t) * 0.28 * overtoneDecay +
      Math.sin(2 * Math.PI * f * 3 * t) * 0.08 * overtoneDecay;
    left[i] += sig * attack * amp * lGain;
    right[i] += sig * attack * amp * rGain;
  }
}

function addBass(startS, durS, midi, amp) {
  const f = midiToFreq(midi);
  const i0 = Math.max(0, Math.floor(startS * SAMPLE_RATE));
  const i1 = Math.min(NUM_SAMPLES, Math.floor((startS + durS) * SAMPLE_RATE));
  for (let i = i0; i < i1; i++) {
    const t = i / SAMPLE_RATE - startS;
    const e = envADSR(t, durS, 0.025, 0.15, 0.7, Math.min(0.25, durS * 0.25));
    const sig =
      Math.sin(2 * Math.PI * f * t) * 0.8 +
      Math.sin(2 * Math.PI * f * 2 * t) * 0.2 * Math.exp(-t * 3.0);
    left[i] += sig * e * amp;
    right[i] += sig * e * amp;
  }
}

function addSoftKick(startS, amp = 0.21) {
  const durS = 0.22;
  const i0 = Math.max(0, Math.floor(startS * SAMPLE_RATE));
  const i1 = Math.min(NUM_SAMPLES, Math.floor((startS + durS) * SAMPLE_RATE));
  let phase = 0;
  for (let i = i0; i < i1; i++) {
    const t = i / SAMPLE_RATE - startS;
    const f = 48 + 95 * Math.exp(-t * 28);
    phase += (2 * Math.PI * f) / SAMPLE_RATE;
    const env = Math.min(1, t / 0.004) * Math.exp(-t * 14);
    const s = Math.sin(phase) * env * amp;
    left[i] += s;
    right[i] += s;
  }
}

function addSoftHat(startS, amp = 0.024, pan = 0.15) {
  const durS = 0.07;
  const i0 = Math.max(0, Math.floor(startS * SAMPLE_RATE));
  const i1 = Math.min(NUM_SAMPLES, Math.floor((startS + durS) * SAMPLE_RATE));
  const lGain = Math.cos(((pan + 1) * Math.PI) / 4);
  const rGain = Math.sin(((pan + 1) * Math.PI) / 4);
  let prev = 0;
  for (let i = i0; i < i1; i++) {
    const t = i / SAMPLE_RATE - startS;
    const env = Math.min(1, t / 0.005) * Math.exp(-t * 42);
    const raw = Math.sin(i * 12.9898) * 43758.5453;
    const n = raw - Math.floor(raw) - 0.5;
    const hp = n - prev;
    prev = n * 0.85;
    left[i] += hp * env * amp * lGain;
    right[i] += hp * env * amp * rGain;
  }
}

function addSoftWhoosh(startS, durS, amp = 0.055) {
  const i0 = Math.max(0, Math.floor(startS * SAMPLE_RATE));
  const i1 = Math.min(NUM_SAMPLES, Math.floor((startS + durS) * SAMPLE_RATE));
  let lpL = 0, lpR = 0;
  for (let i = i0; i < i1; i++) {
    const t = i / SAMPLE_RATE - startS;
    const norm = t / durS;
    const env = Math.sin(norm * Math.PI);
    const cutoff = 0.03 + 0.12 * env;
    const nL = Math.sin(i * 78.233) * 0.5 + Math.cos(i * 19.13) * 0.5;
    const nR = Math.cos(i * 63.719) * 0.5 + Math.sin(i * 27.41) * 0.5;
    lpL += cutoff * (nL - lpL);
    lpR += cutoff * (nR - lpR);
    const sweepF = 180 + 230 * norm;
    const tone = Math.sin(2 * Math.PI * sweepF * t) * 0.35;
    left[i] += (lpL * 0.65 + tone) * env * amp;
    right[i] += (lpR * 0.65 + tone) * env * amp;
  }
}

function addUiClick(startS, amp = 0.14) {
  addPluck(startS, 0.09, 72, amp * 0.65, 0);
  addPluck(startS + 0.012, 0.11, 79, amp * 0.55, 0.1);
}

// 1. CHORD PROGRESSION ACROSS 5 SCENES (24.0s)
const chords = [
  // Scene 1 (0.0 - 4.0s): Fmaj9 -> Cmaj9 (De 1 agente sobrecarregado para 5 especialistas)
  { start: 0.0, dur: 2.1, bass: 41, notes: [57, 60, 64, 67, 71] },
  { start: 2.0, dur: 2.1, bass: 36, notes: [60, 62, 64, 67, 71] },
  // Scene 2 (4.0 - 9.0s): Dm9 -> G13sus4 -> Cmaj9 (Visão Processual & Portão FLX-01)
  { start: 4.0, dur: 2.3, bass: 38, notes: [57, 60, 64, 65, 69] },
  { start: 6.2, dur: 2.9, bass: 43, notes: [59, 62, 65, 69, 72] },
  // Scene 3 (9.0 - 14.5s): Am9 -> Fmaj9 -> G6 (Como Instalar & 4 Comandos Reais)
  { start: 9.0, dur: 2.8, bass: 45, notes: [57, 60, 64, 67, 71] },
  { start: 11.7, dur: 2.9, bass: 41, notes: [60, 64, 65, 69, 72] },
  // Scene 4 (14.5 - 20.0s): Gm9 (tensão do bloqueio SEG-01) -> Fmaj9/C -> Cmaj9 (Gate PASS + Auto-Skill)
  { start: 14.5, dur: 2.7, bass: 43, notes: [58, 62, 65, 69] },
  { start: 17.2, dur: 2.9, bass: 36, notes: [60, 64, 67, 71, 74] },
  // Scene 5 (20.0 - 24.0s): Fmaj9#11 -> Dmaj9 (Celebração final + forge-sdd.org)
  { start: 20.0, dur: 4.0, bass: 38, notes: [57, 61, 64, 66, 69, 73] },
];

for (const ch of chords) {
  addBass(ch.start, ch.dur, ch.bass, 0.16);
  ch.notes.forEach((m, idx) => {
    const pan = ((idx / Math.max(1, ch.notes.length - 1)) - 0.5) * 0.7;
    addWarmPadNote(ch.start, ch.dur, m, 0.046, pan);
  });
}

// 2. RHYTHMIC PULSE & ARPEGGIO (120 BPM)
const arpScale = [60, 64, 67, 71, 72, 71, 67, 64];
for (let step = 0; step < 96; step++) {
  const t = step * 0.25;
  if (t >= 20.2) break;
  const note = arpScale[step % arpScale.length];
  const isGroove = t >= 1.5 && t < 20.0;
  const arpAmp = isGroove ? 0.04 : 0.022;
  const pan = Math.sin(step * 0.85) * 0.45;
  addPluck(t, 0.34, note, arpAmp, pan);

  if (isGroove && step % 2 === 0) {
    addSoftKick(t, 0.165);
  }
  if (isGroove && step % 2 === 1) {
    addSoftHat(t, 0.02, step % 4 === 1 ? -0.25 : 0.25);
  }
}

// 3. SYNCHRONIZED EDUCATIONAL CUES & SFX
// Scene 1: 5 stations pop in sequentially (0.8s, 1.05s, 1.3s, 1.55s, 1.8s)
[67, 69, 72, 74, 76].forEach((m, idx) => {
  addPluck(0.8 + idx * 0.24, 0.42, m, 0.088, (idx - 2) * 0.22);
});

// Scene 2 (4.0s): Transition to Procedural Flow (01 -> 02 -> 03 -> 04 -> 05)
addSoftWhoosh(3.88, 0.42, 0.06);
addSoftKick(4.0, 0.22);

// Scene 2 (6.2s): User clicks "Aprovar Spec (Portão FLX-01)" -> baton passes 02 -> 03 -> 04
addUiClick(6.20, 0.14);
addPluck(6.32, 0.5, 76, 0.095, -0.2);
addPluck(6.44, 0.55, 79, 0.10, 0.0);
addPluck(6.58, 0.65, 84, 0.11, 0.25);

// Scene 3 (9.0s): Transition to Practical Emulator (Install + 4 Daily Commands)
addSoftWhoosh(8.88, 0.40, 0.055);
// 4 command steps lighting up at 9.5s, 10.7s, 11.9s, 13.1s
[69, 72, 76, 79].forEach((m, idx) => {
  const stepT = 9.5 + idx * 1.2;
  addUiClick(stepT, 0.11);
  addPluck(stepT + 0.08, 0.45, m, 0.085, (idx - 1.5) * 0.2);
});

// Scene 4 (14.5s): Transition to 3-Layer Governance
addSoftWhoosh(14.38, 0.40, 0.055);
// 15.3s: Layer A/B blocks SEG-01 / P-13 and triggers loop 05 -> 04
addPluck(15.30, 0.45, 62, 0.10, 0.2);
addPluck(15.44, 0.55, 58, 0.105, -0.2);
// 17.3s: Builder fixes, Gate PASS + Auto-Skill DUP-05 generated!
addUiClick(17.25, 0.13);
[76, 79, 83, 86].forEach((m, idx) => {
  addPluck(17.35 + idx * 0.11, 0.65, m, 0.095, (idx - 1.5) * 0.2);
});

// Scene 5 (20.0s): Outro / 5 Mascots Celebrating + forge-sdd.org
addSoftWhoosh(19.86, 0.42, 0.06);
addSoftKick(20.0, 0.22);
[74, 78, 81, 85, 86].forEach((m, idx) => {
  addPluck(20.08 + idx * 0.09, 1.5, m, 0.08, (idx - 2) * 0.2);
});

// 4. STEREO DELAY & MASTER LIMITER
const delaySamplesL = Math.floor(0.1875 * SAMPLE_RATE);
const delaySamplesR = Math.floor(0.25 * SAMPLE_RATE);
for (let i = delaySamplesR; i < NUM_SAMPLES; i++) {
  left[i] += right[i - delaySamplesL] * 0.16;
  right[i] += left[i - delaySamplesR] * 0.16;
}

const pcm = Buffer.alloc(NUM_SAMPLES * 4);
for (let i = 0; i < NUM_SAMPLES; i++) {
  const t = i / SAMPLE_RATE;
  let masterEnv = 1.0;
  if (t < 0.08) masterEnv = t / 0.08;
  if (t > DURATION - 0.75) masterEnv = Math.max(0, (DURATION - t) / 0.75);

  const l = Math.tanh(left[i] * masterEnv * 0.95) * 0.88;
  const r = Math.tanh(right[i] * masterEnv * 0.95) * 0.88;

  const sL = Math.max(-32767, Math.min(32767, Math.round(l * 32767)));
  const sR = Math.max(-32767, Math.min(32767, Math.round(r * 32767)));
  pcm.writeInt16LE(sL, i * 4);
  pcm.writeInt16LE(sR, i * 4 + 2);
}

const wavHeader = Buffer.alloc(44);
wavHeader.write("RIFF", 0);
wavHeader.writeUInt32LE(36 + pcm.length, 4);
wavHeader.write("WAVE", 8);
wavHeader.write("fmt ", 12);
wavHeader.writeUInt32LE(16, 16);
wavHeader.writeUInt16LE(1, 20);
wavHeader.writeUInt16LE(2, 22);
wavHeader.writeUInt32LE(SAMPLE_RATE, 24);
wavHeader.writeUInt32LE(SAMPLE_RATE * 4, 28);
wavHeader.writeUInt16LE(4, 32);
wavHeader.writeUInt16LE(16, 34);
wavHeader.write("data", 36);
wavHeader.writeUInt32LE(pcm.length, 40);

const outPath = path.resolve("/Users/nathanramorim/git/forge-sdd/brag-output/work/soundtrack.wav");
fs.mkdirSync(path.dirname(outPath), { recursive: true });
fs.writeFileSync(outPath, Buffer.concat([wavHeader, pcm]));
console.log("Generated soundtrack.wav:", outPath, (pcm.length + 44) + " bytes");
