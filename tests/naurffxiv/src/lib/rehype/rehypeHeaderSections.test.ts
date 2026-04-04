import { expect, test } from 'vitest';
import rehypeHeaderSections from '@/lib/rehype/rehypeHeaderSections';
import type { Root, Element } from 'hast';

test('Section with h2 first child', () => {
  const tree: Root = {
    type: 'root',
    children: [
      {
        type: 'element',
        tagName: 'section',
        properties: {},
        children: [
          {
            type: 'element',
            tagName: 'h2',
            properties: { id: 'test-id' },
            children: []
          }
        ]
      }
    ]
  };

  rehypeHeaderSections()(tree);

  const section = tree.children[0] as Element;
  expect(section.properties?.['id']).toBe('test-id');
  expect(section.properties?.['className']).toEqual(expect.arrayContaining(['scroll-mt-[5.5rem]']));

  const h2 = section.children[0] as Element;
  expect(h2.properties?.['id']).toBeUndefined();
});

test('Section with h3-h6 first child', () => {
  const headings = ['h3', 'h4', 'h5', 'h6'];

  for (const h of headings) {
    const tree: Root = {
      type: 'root',
      children: [{
        type: 'element',
        tagName: 'section',
        properties: {},
        children: [{
          type: 'element',
          tagName: h,
          properties: { id: `${h}-id` },
          children: []
        }]
      }]
    };

    rehypeHeaderSections()(tree);
    const section = tree.children[0] as Element;
    expect(section.properties?.['id']).toBe(`${h}-id`);
    expect(section.properties?.['className']).toEqual(expect.arrayContaining(['scroll-mt-[5.5rem]']));
  }
});

test('Section with p first child', () => {
  const tree: Root = {
    type: 'root',
    children: [{
      type: 'element',
      tagName: 'section',
      properties: {},
      children: [{
        type: 'element',
        tagName: 'p',
        properties: { id: 'p-id' },
        children: []
      }]
    }]
  };

  rehypeHeaderSections()(tree);
  const section = tree.children[0] as Element;
  expect(section.properties?.['id']).toBeUndefined();
  const p = section.children[0] as Element;
  expect(p.properties?.['id']).toBe('p-id');
});

test('Section with no children', () => {
  const tree: Root = {
    type: 'root',
    children: [{
      type: 'element',
      tagName: 'section',
      properties: {},
      children: []
    }]
  };

  expect(() => rehypeHeaderSections()(tree)).not.toThrow();
  const section = tree.children[0] as Element;
  expect(section.properties?.['id']).toBeUndefined();
});

test('Non-section element with heading', () => {
  const tree: Root = {
    type: 'root',
    children: [{
      type: 'element',
      tagName: 'div',
      properties: {},
      children: [{
        type: 'element',
        tagName: 'h2',
        properties: { id: 'h2-id' },
        children: []
      }]
    }]
  };

  rehypeHeaderSections()(tree);
  const div = tree.children[0] as Element;
  expect(div.properties?.['id']).toBeUndefined();
});
