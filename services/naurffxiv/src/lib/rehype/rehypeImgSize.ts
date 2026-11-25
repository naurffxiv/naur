import { readFileSync } from "fs";
import path from "path";
import { imageSize } from "image-size";
import { visit } from "unist-util-visit";
import type { Root, Element } from "hast";

const remoteSource = /^(?:[a-z]+:)?\/\//;

export default function rehypeImgSize(options: { dir?: string } = {}) {
  return function (tree: Root): void {
    visit(tree, "element", function (node: Element) {
      if (node.tagName !== "img") return;

      const src = node.properties["src"];
      if (typeof src !== "string" || remoteSource.test(src)) return;

      const file =
        options.dir && (!path.isAbsolute(src) || src.startsWith("/"))
          ? path.join(options.dir, src)
          : src;
      const { width, height } = imageSize(readFileSync(file));

      node.properties["width"] = width;
      node.properties["height"] = height;
    });
  };
}
