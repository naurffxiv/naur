import { afterAll, beforeAll, expect, test } from 'vitest';
import { mkdtempSync, rmSync, writeFileSync, mkdirSync } from 'fs';
import { tmpdir } from 'os';
import path from 'path';
import rehypeImgSize from '@/lib/rehype/rehypeImgSize';
import type { Root, Element } from 'hast';

let dir: string;

function png(width: number, height: number): Buffer {
  const buffer = Buffer.alloc(33);
  Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]).copy(buffer, 0);
  buffer.writeUInt32BE(13, 8);
  buffer.write('IHDR', 12, 'ascii');
  buffer.writeUInt32BE(width, 16);
  buffer.writeUInt32BE(height, 20);
  buffer[24] = 8;
  buffer[25] = 6;
  return buffer;
}

function img(src: string): Element {
  return { type: 'element', tagName: 'img', properties: { src }, children: [] };
}

function run(src: string, options?: { dir?: string }): Element {
  const node = img(src);
  const tree: Root = { type: 'root', children: [node] };
  rehypeImgSize(options)(tree);
  return node;
}

beforeAll(() => {
  dir = mkdtempSync(path.join(tmpdir(), 'rehype-img-size-'));
  mkdirSync(path.join(dir, 'images'));
  writeFileSync(path.join(dir, 'images', 'wide.png'), png(640, 360));
});

afterAll(() => {
  rmSync(dir, { recursive: true, force: true });
});

test('Root-relative src resolves against dir', () => {
  const node = run('/images/wide.png', { dir });
  expect(node.properties.width).toBe(640);
  expect(node.properties.height).toBe(360);
});

test('Relative src resolves against dir', () => {
  const node = run('images/wide.png', { dir });
  expect(node.properties.width).toBe(640);
  expect(node.properties.height).toBe(360);
});

test('Without dir, src is read as given', () => {
  const node = run(path.join(dir, 'images', 'wide.png'));
  expect(node.properties.width).toBe(640);
  expect(node.properties.height).toBe(360);
});

test('Remote sources are left alone', () => {
  for (const src of ['https://xivapi.com/i/a.png', 'http://example.com/a.png', '//cdn.example.com/a.png']) {
    const node = run(src, { dir });
    expect(node.properties.width).toBeUndefined();
    expect(node.properties.height).toBeUndefined();
  }
});

test('Missing file throws', () => {
  expect(() => run('/images/missing.png', { dir })).toThrow();
});
