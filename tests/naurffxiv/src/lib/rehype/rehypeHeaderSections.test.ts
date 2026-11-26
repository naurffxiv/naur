import { expect, test } from "vitest";
import rehypeHeaderSections from "@/lib/rehype/rehypeHeaderSections";
import type { Root, Element } from "hast";

test("Section with h2 first child", () => {
  const tree: Root = {
    type: "root",
    children: [
      {
        type: "element",
        tagName: "section",
        properties: {},
        children: [
          {
            type: "element",
            tagName: "h2",
            properties: { id: "test-id" },
            children: [],
          },
        ],
      },
    ],
  };

  rehypeHeaderSections()(tree);

  const section = tree.children[0] as Element;
  expect(section.properties?.["id"]).toBe("test-id");
  expect(section.properties?.["className"]).toEqual(
    expect.arrayContaining(["scroll-mt-[5.5rem]"]),
  );

  const h2 = section.children[0] as Element;
  expect(h2.properties?.["id"]).toBeUndefined();
});

test("Section with h3-h6 first child", () => {
  const headings = ["h3", "h4", "h5", "h6"];

  for (const h of headings) {
    const tree: Root = {
      type: "root",
      children: [
        {
          type: "element",
          tagName: "section",
          properties: {},
          children: [
            {
              type: "element",
              tagName: h,
              properties: { id: `${h}-id` },
              children: [],
            },
          ],
        },
      ],
    };

    rehypeHeaderSections()(tree);
    const section = tree.children[0] as Element;
    expect(section.properties?.["id"]).toBe(`${h}-id`);
    expect(section.properties?.["className"]).toEqual(
      expect.arrayContaining(["scroll-mt-[5.5rem]"]),
    );
  }
});

test("Section with p first child", () => {
  const tree: Root = {
    type: "root",
    children: [
      {
        type: "element",
        tagName: "section",
        properties: {},
        children: [
          {
            type: "element",
            tagName: "p",
            properties: { id: "p-id" },
            children: [],
          },
        ],
      },
    ],
  };

  rehypeHeaderSections()(tree);
  const section = tree.children[0] as Element;
  expect(section.properties?.["id"]).toBeUndefined();
  const p = section.children[0] as Element;
  expect(p.properties?.["id"]).toBe("p-id");
});

test("Section with no children", () => {
  const tree: Root = {
    type: "root",
    children: [
      {
        type: "element",
        tagName: "section",
        properties: {},
        children: [],
      },
    ],
  };

  expect(() => rehypeHeaderSections()(tree)).not.toThrow();
  const section = tree.children[0] as Element;
  expect(section.properties?.["id"]).toBeUndefined();
});

test("Non-section element with heading", () => {
  const tree: Root = {
    type: "root",
    children: [
      {
        type: "element",
        tagName: "div",
        properties: {},
        children: [
          {
            type: "element",
            tagName: "h2",
            properties: { id: "h2-id" },
            children: [],
          },
        ],
      },
    ],
  };

  rehypeHeaderSections()(tree);
  const div = tree.children[0] as Element;
  expect(div.properties?.["id"]).toBeUndefined();
});

test("Section keeps its existing properties and classes", () => {
  const tree: Root = {
    type: "root",
    children: [
      {
        type: "element",
        tagName: "section",
        properties: { className: ["prose"], dataLevel: "2" },
        children: [
          {
            type: "element",
            tagName: "h2",
            properties: { id: "h2-id" },
            children: [],
          },
        ],
      },
    ],
  };

  rehypeHeaderSections()(tree);
  const section = tree.children[0] as Element;
  expect(section.properties).toEqual({
    id: "h2-id",
    className: ["prose", "scroll-mt-[5.5rem]"],
    dataLevel: "2",
  });
});

test("Section without a className gets only the scroll margin class", () => {
  const tree: Root = {
    type: "root",
    children: [
      {
        type: "element",
        tagName: "section",
        properties: {},
        children: [
          {
            type: "element",
            tagName: "h2",
            properties: { id: "h2-id" },
            children: [],
          },
        ],
      },
    ],
  };

  rehypeHeaderSections()(tree);
  const section = tree.children[0] as Element;
  expect(section.properties?.["className"]).toEqual(["scroll-mt-[5.5rem]"]);
});

test("Section and heading without properties do not throw", () => {
  const tree = {
    type: "root",
    children: [
      {
        type: "element",
        tagName: "section",
        children: [
          {
            type: "element",
            tagName: "h2",
            children: [],
          },
        ],
      },
    ],
  } as unknown as Root;

  expect(() => rehypeHeaderSections()(tree)).not.toThrow();
  const section = tree.children[0] as Element;
  expect(section.properties?.["className"]).toEqual(["scroll-mt-[5.5rem]"]);
  expect(section.properties?.["id"]).toBeUndefined();
});
