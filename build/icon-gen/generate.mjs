/**
 * Чёрный гербовый щит + белая «P» из Keania One (контур из woff2 проекта).
 * Запуск: из каталога icon-gen — node generate.mjs
 * Выход: ../appicon.png (для wails build).
 */
import path from 'path';
import { fileURLToPath } from 'url';
import { openSync as openFont } from 'fontkit';
import sharp from 'sharp';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const fontPath = path.join(
    __dirname,
    '../../frontend/node_modules/@fontsource/keania-one/files/keania-one-latin-400-normal.woff2',
);

const font = openFont(fontPath);
const glyph = font.glyphForCodePoint(0x50); // P
const letterD = glyph.path.toSVG();
const bbox = glyph.path.bbox;
const cx = (bbox.minX + bbox.maxX) / 2;
const cy = (bbox.minY + bbox.maxY) / 2;
const gw = bbox.maxX - bbox.minX;

/** Гербовый щит (heater), чёрная заливка, вписан в 1024×1024 */
const shieldD =
    'M 512 76 ' +
    'C 276 76 96 226 96 448 L 96 598 ' +
    'C 96 734 192 864 378 940 L 512 992 ' +
    'L 646 940 ' +
    'C 832 864 928 734 928 598 L 928 448 ' +
    'C 928 226 748 76 512 76 Z';

const targetLetterW = 520;
const scale = targetLetterW / gw;
const ty = 458;

const svg = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <path fill="#000000" d="${shieldD}"/>
  <g transform="translate(512 ${ty}) scale(${scale}) scale(1,-1) translate(${-cx},${-cy})">
    <path fill="#ffffff" d="${letterD}"/>
  </g>
</svg>`;

const outPng = path.join(__dirname, '../appicon.png');
await sharp(Buffer.from(svg)).png().toFile(outPng);

console.log('Wrote', outPng);
