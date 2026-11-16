#!/usr/bin/env python3
"""Sintetiza a trilha do vídeo (original, sem direitos autorais) em Python puro.

Pad suave, baixo, arpejo em pluck e batida leve a partir do compasso 5. 104 BPM, Dó maior
(C - G - Am - F). Uso: python3 -I musica.py <saida.wav> [duracao=80]
"""
import math
import random
import struct
import sys
import wave

SR = 44100
BPM = 104
BEAT = 60.0 / BPM
out = sys.argv[1] if len(sys.argv) > 1 else "musica.wav"
dur = float(sys.argv[2]) if len(sys.argv) > 2 else 80.0
N = int(SR * dur)
buf = [0.0] * N
random.seed(7)

NOTE = {"C": 0, "D": 2, "E": 4, "F": 5, "G": 7, "A": 9, "B": 11}
def hz(name, octave):
    return 440.0 * 2 ** ((NOTE[name] + 12 * (octave - 4) - 9) / 12)

CHORDS = [("C", ["C", "E", "G"]), ("G", ["G", "B", "D"]), ("A", ["A", "C", "E"]), ("F", ["F", "A", "C"])]

def add(start, samples, gain=1.0):
    i0 = int(start * SR)
    for k, v in enumerate(samples):
        j = i0 + k
        if j >= N:
            break
        buf[j] += v * gain

def tone(freq, length, shape, attack=0.01, release=0.2, decay=0.0):
    n = int(length * SR)
    res = []
    for k in range(n):
        t = k / SR
        env = min(1.0, t / attack) if attack > 0 else 1.0
        env *= min(1.0, (length - t) / release) if release > 0 else 1.0
        if decay:
            env *= math.exp(-decay * t)
        ph = 2 * math.pi * freq * t
        res.append(shape(ph) * env)
    return res

sine = lambda p: math.sin(p)
soft = lambda p: math.sin(p) + 0.35 * math.sin(2 * p) + 0.12 * math.sin(3 * p)
tri = lambda p: (2 / math.pi) * math.asin(math.sin(p))

bars = int(dur / (4 * BEAT)) + 1
for b in range(bars):
    t0 = b * 4 * BEAT
    root, tones = CHORDS[b % 4]
    # pad
    for nm in tones:
        add(t0, tone(hz(nm, 4), 4 * BEAT, sine, attack=0.6, release=0.9), 0.055)
        add(t0, tone(hz(nm, 3), 4 * BEAT, sine, attack=0.8, release=1.0), 0.03)
    # baixo: tempo 1 e "e" do 3
    add(t0, tone(hz(root, 2), 1.6 * BEAT, sine, attack=0.02, release=0.3, decay=1.4), 0.34)
    add(t0 + 2.5 * BEAT, tone(hz(root, 2), 1.3 * BEAT, sine, attack=0.02, release=0.3, decay=1.6), 0.26)
    # arpejo em colcheias
    order = [0, 1, 2, 1, 2, 1, 0, 1] if b % 2 == 0 else [0, 2, 1, 2, 0, 2, 1, 2]
    for s, idx in enumerate(order):
        nm = tones[idx]
        add(t0 + s * BEAT / 2, tone(hz(nm, 5), 0.9, soft, attack=0.004, release=0.2, decay=5.5), 0.075)
    # melodia leve a cada 2 compassos (a partir do 3º)
    if b >= 2 and b % 2 == 0:
        mel = [tones[2], tones[1], tones[2], tones[0]]
        for s, nm in enumerate(mel):
            add(t0 + (s * 2 + 1) * BEAT / 2, tone(hz(nm, 6), 0.7, tri, attack=0.006, release=0.25, decay=4.0), 0.05)
    # batida a partir do compasso 5
    if b >= 4:
        for beat in (0, 2):
            n = int(0.28 * SR)
            kick = [math.sin(2 * math.pi * (50 + 90 * math.exp(-k / SR * 28)) * k / SR) * math.exp(-k / SR * 14) for k in range(n)]
            add(t0 + beat * BEAT, kick, 0.5)
        for beat in range(8):
            if beat % 2 == 1:
                n = int(0.05 * SR)
                hat = [(random.random() * 2 - 1) * math.exp(-k / SR * 90) for k in range(n)]
                add(t0 + beat * BEAT / 2, hat, 0.07)
        # estalo suave no 2 e 4
        for beat in (1, 3):
            n = int(0.12 * SR)
            snap = [(random.random() * 2 - 1) * math.exp(-k / SR * 40) * 0.5 + math.sin(2 * math.pi * 190 * k / SR) * math.exp(-k / SR * 30) * 0.5 for k in range(n)]
            add(t0 + beat * BEAT, snap, 0.16)

# fade in 1 s, fade out nos últimos 4 s, normaliza
peak = max(abs(v) for v in buf) or 1.0
scale = 0.85 / peak
frames = bytearray()
for i, v in enumerate(buf):
    t = i / SR
    g = min(1.0, t / 1.0) * min(1.0, max(0.0, (dur - t) / 4.0))
    s = max(-1.0, min(1.0, v * scale * g))
    frames += struct.pack("<h", int(s * 32767))
with wave.open(out, "wb") as w:
    w.setnchannels(1)
    w.setsampwidth(2)
    w.setframerate(SR)
    w.writeframes(bytes(frames))
print(f"ok: {out} ({dur:.0f}s)")
