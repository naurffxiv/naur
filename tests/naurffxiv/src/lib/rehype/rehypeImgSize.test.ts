import { afterAll, beforeAll, expect, test, vi } from 'vitest';
import { mkdtempSync, rmSync, writeFileSync, mkdirSync } from 'fs';
import { tmpdir } from 'os';
import path from 'path';
import rehypeImgSize from '@/lib/rehype/rehypeImgSize';
import type { Root, Element } from 'hast';

let root: string;

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

function run(src: string): Element {
  const node: Element = { type: 'element', tagName: 'img', properties: { src }, children: [] };
  const tree: Root = { type: 'root', children: [node] };
  rehypeImgSize()(tree);
  return node;
}

beforeAll(() => {
  root = mkdtempSync(path.join(tmpdir(), 'rehype-img-size-'));
  mkdirSync(path.join(root, 'public', 'images'), { recursive: true });
  writeFileSync(path.join(root, 'public', 'images', 'wide.png'), png(640, 360));
  vi.spyOn(process, 'cwd').mockReturnValue(root);
});

afterAll(() => {
  vi.restoreAllMocks();
  rmSync(root, { recursive: true, force: true });
});

test('Root-relative src resolves against public', () => {
  const node = run('/images/wide.png');
  expect(node.properties.width).toBe(640);
  expect(node.properties.height).toBe(360);
});

test('Relative src resolves against public', () => {
  const node = run('images/wide.png');
  expect(node.properties.width).toBe(640);
  expect(node.properties.height).toBe(360);
});

test('Remote sources are left alone', () => {
  for (const src of ['https://xivapi.com/i/a.png', 'http://example.com/a.png', '//cdn.example.com/a.png']) {
    const node = run(src);
    expect(node.properties.width).toBeUndefined();
    expect(node.properties.height).toBeUndefined();
  }
});

test('Missing file throws', () => {
  expect(() => run('/images/missing.png')).toThrow();
});
