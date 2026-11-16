# Vídeo do pipeline (fontes)

Gera `site/public/video/forge-sdd-pipeline.mp4` (80 s, 1920×1080, com música). Tudo é código: nada é gravado de tela.

- `video.html`: todas as cenas, animadas em função do tempo (`render(t)`), na identidade da landing. Textos, tempos, falas dos terminais e legendas ficam nos objetos `L`, `ST`, `ACT`, `PK` e `CAP`.
- `render.mjs`: abre o `video.html` no Chrome headless (protocolo DevTools) e salva um JPEG por quadro.
- `musica.py`: sintetiza a trilha original (Python puro, sem dependências).

## Como regerar
```bash
cd docs/video-pipeline
python3 -I musica.py /tmp/musica.wav 80
node render.mjs /tmp/quadros 30 80              # 2400 quadros; com um 4º argumento (segundos) salva só aquele quadro
ffmpeg -framerate 30 -i /tmp/quadros/%05d.jpg -i /tmp/musica.wav \
  -c:v libx264 -preset slow -crf 22 -pix_fmt yuv420p -movflags +faststart \
  -c:a aac -b:a 160k -ac 2 -shortest ../../site/public/video/forge-sdd-pipeline.mp4
cp /tmp/quadros/00780.jpg ../../site/public/video/forge-sdd-pipeline-poster.jpg
```

Requisitos: Node 22+, Python 3, ffmpeg e Google Chrome (ou `CHROME=/caminho/do/chrome`).
