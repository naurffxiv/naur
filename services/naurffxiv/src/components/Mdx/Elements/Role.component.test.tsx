// @vitest-environment jsdom
import React from "react";
import { render } from "@testing-library/react";
import { describe, expect, test } from "vitest";
import "@testing-library/jest-dom/vitest";

import Role from "./Role";

describe("Role component", () => {
  describe("image accessibility", () => {
    test("image is decorative (alt='') when text is present", () => {
      const { container } = render(<Role>tank</Role>);
      expect(container.querySelector("img")).toHaveAttribute("alt", "");
    });

    test("image conveys role (alt=roleText) when no text is present", () => {
      const { container } = render(<Role role="tank" />);
      expect(container.querySelector("img")).toHaveAttribute("alt", "tank");
    });

    test("image has title for hover when text is present", () => {
      const { container: withText } = render(<Role>healer</Role>);
      expect(withText.querySelector("img")).toHaveAttribute(
        "title",
        "healer icon",
      );
    });

    test("image has title for hover when no text is present", () => {
      const { container: iconOnly } = render(<Role role="mt" />);
      expect(iconOnly.querySelector("img")).toHaveAttribute("title", "mt icon");
    });
  });

  describe("icon visibility", () => {
    test("image is present when text is present", () => {
      const { container } = render(<Role>tank</Role>);
      expect(container.querySelector("img")).not.toBeNull();
    });

    test('image is not present when icon="false"', () => {
      const { container } = render(<Role icon="false">tank</Role>);
      expect(container.querySelector("img")).toBeNull();
    });
  });

  describe("space insertion", () => {
    test("NNBSP inserted when text and icon present", () => {
      const { container } = render(<Role>healer</Role>);
      expect(container.querySelector("mark")?.textContent).toContain("\u202F");
    });

    test("NNBSP not inserted when text not present", () => {
      const { container } = render(<Role role="d1" />);
      expect(container.querySelector("mark")?.textContent).not.toContain(
        "\u202F",
      );
    });

    test("NNBSP not inserted when icon not present", () => {
      const { container } = render(<Role icon="false">shield healer</Role>);
      expect(container.querySelector("mark")?.textContent).not.toContain(
        "\u202F",
      );
    });
  });

  describe("CSS class per role", () => {
    test.for([
      { input: "tank", expected: "hl-tank" },
      { input: "healer", expected: "hl-healer" },
      { input: "support", expected: "hl-support" },
      { input: "melee dps", expected: "hl-dps" },
    ])('Role("$input") applies class "$expected"', ({ input, expected }) => {
      const { container } = render(<Role>{input}</Role>);
      expect(container.querySelector("mark")).toHaveClass(expected);
    });
  });

  describe("invalid role rejection", () => {
    test("throws an error when the text is not a valid role", () => {
      expect(() => render(<Role>yoship sampo</Role>)).toThrow(
        'Could not infer role from: "yoship sampo"',
      );
    });

    test("throws an error when the role attribute is not a valid role", () => {
      expect(() => render(<Role role="9th">support</Role>)).toThrow(
        'Could not infer role from: "9th"',
      );
    });

    test("role attribute allows for text content that would otherwise be rejected", () => {
      expect(() =>
        render(<Role role="magic dps">yoship sampo</Role>),
      ).not.toThrow();
    });
  });
});
