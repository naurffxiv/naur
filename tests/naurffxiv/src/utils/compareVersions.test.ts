import { expect, test } from 'vitest';
import { parseVersion, compareVersions } from '@/utils/compareVersions';
import type { GitHubRelease } from '@/utils/fetchGithubReleases';

test('parseVersion("v1.2.3") → [1, 2, 3]', () => {
  expect(parseVersion('v1.2.3')).toEqual([1, 2, 3]);
});

test('parseVersion("1.0") → [1, 0]', () => {
  expect(parseVersion('1.0')).toEqual([1, 0]);
});

test('compareVersions — b newer', () => {
  const a = { tag_name: 'v1.0.0' } as GitHubRelease;
  const b = { tag_name: 'v2.0.0' } as GitHubRelease;
  expect(compareVersions(a, b)).toBeGreaterThan(0);
});

test('compareVersions — a newer', () => {
  const a = { tag_name: 'v2.0.0' } as GitHubRelease;
  const b = { tag_name: 'v1.0.0' } as GitHubRelease;
  expect(compareVersions(a, b)).toBeLessThan(0);
});

test('compareVersions — equal', () => {
  const a = { tag_name: 'v1.2.3' } as GitHubRelease;
  const b = { tag_name: 'v1.2.3' } as GitHubRelease;
  expect(compareVersions(a, b)).toBe(0);
});

test('compareVersions — different segment lengths (v1.2 vs v1.2.0)', () => {
  const a = { tag_name: 'v1.2' } as GitHubRelease;
  const b = { tag_name: 'v1.2.0' } as GitHubRelease;
  expect(compareVersions(a, b)).toBe(0);
});
