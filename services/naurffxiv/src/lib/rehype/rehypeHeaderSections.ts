import { visit } from "unist-util-visit";
import type { Root, Element } from "hast";

/*
    This rehype plugin searches for <section> tags and then checks whether its first child
    is a header tag from h2-h6. If it is, transfer the header's id to the section instead.

    This makes it so that fragments target the section. It also improves IntersectionObserver
    since it is now tracking an entire section instead of a single header.
*/
export default function rehypeHeaderSections() {
  return function (tree: Root): void {
    visit(tree, "element", function (node: Element) {
      if (node.tagName !== "section") return;
      if (!node.children || node.children.length === 0) return;

      const allowedTags = ["h2", "h3", "h4", "h5", "h6"];
      const firstChild = node.children[0];

      if (firstChild.type !== "element") return;
      if (allowedTags.indexOf(firstChild.tagName) === -1) return;

      if (!node.properties) {
        node.properties = {};
      }
      if (!firstChild.properties) {
        firstChild.properties = {};
      }

      node.properties["id"] = firstChild.properties["id"];

      const existingClasses = Array.isArray(node.properties["className"])
        ? (node.properties["className"] as string[])
        : [];
      node.properties["className"] = [...existingClasses, "scroll-mt-[5.5rem]"];

      delete firstChild.properties["id"];
    });
  };
}
