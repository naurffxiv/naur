import { afterEach, expect, test, vi } from 'vitest';
import imageLoader from '@/lib/imageLoader';
import { variantUrl } from '@/lib/imageVariants.mjs';

afterEach(() => {
  vi.unstubAllEnvs();
});

test('variantUrl — public image → pre-generated webp at that width', () => {
  expect(variantUrl('/images/ultimate/fru/banner.avif', 828)).toBe(
    '/_variants/images/ultimate/fru/banner.avif.828.webp',
  );
});

test('variantUrl — keeps the source extension so banner.png and banner.avif differ', () => {
  expect(variantUrl('/images/m9s/banner.png', 640)).not.toBe(
    variantUrl('/images/m9s/banner.avif', 640),
  );
});

test('variantUrl — remote, bundled, svg and gif images are not pre-generated', () => {
  expect(variantUrl('https://cdn.discordapp.com/avatars/1/a.png', 64)).toBeNull();
  expect(variantUrl('//xivapi.com/i/000000/000001.png', 64)).toBeNull();
  expect(variantUrl('/_next/static/media/naur-icon.6777cfdf.png', 64)).toBeNull();
  expect(variantUrl('/images/logo.svg', 64)).toBeNull();
  expect(variantUrl('/images/anim.gif', 64)).toBeNull();
});

test('imageLoader — production serves the variant', () => {
  vi.stubEnv('NODE_ENV', 'production');
  expect(imageLoader({ src: '/images/a.png', width: 1200 })).toBe(
    '/_variants/images/a.png.1200.webp',
  );
});

test('imageLoader — production passes through images without variants', () => {
  vi.stubEnv('NODE_ENV', 'production');
  const src = 'https://cdn.discordapp.com/avatars/1/a.png';
  expect(imageLoader({ src, width: 64 })).toBe(src);
});

test('imageLoader — development serves the original with a width query', () => {
  vi.stubEnv('NODE_ENV', 'development');
  expect(imageLoader({ src: '/images/a.png', width: 640 })).toBe('/images/a.png?w=640');
  expect(imageLoader({ src: '/images/a.png?v=2', width: 640 })).toBe(
    '/images/a.png?v=2&w=640',
  );
});
