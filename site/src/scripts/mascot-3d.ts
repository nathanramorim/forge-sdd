import * as THREE from 'three';
import { STATE_DEFS, type MascotMark, type MascotState } from '../data/stations';

export interface MascotController {
  setState: (nextState: MascotState, nextColorHex?: number) => void;
  triggerHop: () => void;
  update: (t: number, lookX: number, lookY: number) => void;
  destroy: () => void;
}

export function createForgeMascot3D(
  mountEl: HTMLElement,
  agentHex: number,
  initialState: MascotState,
  markType: MascotMark
): MascotController {
  const w = mountEl.clientWidth || 78;
  const h = mountEl.clientHeight || 86;

  let isVisible = true;
  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      isVisible = entry.isIntersecting;
    }
  }, { threshold: 0.05 });
  observer.observe(mountEl);

  const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.setSize(w, h, false);

  // Fade out / replace SVG fallback smoothly without layout shift
  const fallbackSvg = mountEl.querySelector('.mascot-fallback-svg');
  if (fallbackSvg) {
    fallbackSvg.remove();
  }
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
  let markMesh: THREE.Mesh;
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

  let currentState: MascotState = initialState || 'working';
  let hopTimer = 0;

  function setState(nextState: MascotState, nextColorHex?: number) {
    currentState = STATE_DEFS[nextState] ? nextState : 'working';
    if (nextColorHex !== undefined) bodyMat.color.setHex(nextColorHex);
    const def = STATE_DEFS[currentState];
    orbMat.color.setHex(def.orb);
    capGroup.visible = def.eyes === 'cap';
    roundGroup.visible = def.eyes === 'round';
    happyGroup.visible = def.eyes === 'happy';
    xGroup.visible = def.eyes === 'x';
    hopTimer = 0.45;
  }

  setState(currentState);

  return {
    setState,
    triggerHop: () => { hopTimer = 0.55; },
    update: (t: number, lookX: number, lookY: number) => {
      // Smart pause when page is hidden or element is out of viewport
      if (document.hidden || !isVisible) {
        return;
      }

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
      eyesRoot.position.x = THREE.MathUtils.clamp(lookX * 0.08 + scanX, -0.09, 0.09);
      eyesRoot.position.y = -0.1 + THREE.MathUtils.clamp(-lookY * 0.06, -0.06, 0.06);
      botGroup.rotation.y = THREE.MathUtils.clamp(lookX * 0.18, -0.25, 0.25);

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
    },
    destroy: () => {
      observer.disconnect();
      renderer.dispose();
    }
  };
}
