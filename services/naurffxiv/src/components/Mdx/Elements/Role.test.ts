import { describe, expect, test } from "vitest";

import { parseRole, type RoleKey } from "./Role";

type Case = { input: string; expected: RoleKey | null };

describe("parseRole", () => {
  describe("matches each role", () => {
    test.for([
      { input: "t1", expected: "T1" },
      { input: "mt", expected: "T1" },
      { input: "tank1", expected: "T1" },
      { input: "Tank 1", expected: "T1" },
      { input: "Main Tank", expected: "T1" },
      { input: "main tank", expected: "T1" },
      { input: "t2", expected: "T2" },
      { input: "ot", expected: "T2" },
      { input: "st", expected: "T2" },
      { input: "tank2", expected: "T2" },
      { input: "Tank 2", expected: "T2" },
      { input: "off tank", expected: "T2" },
      { input: "off-tank", expected: "T2" },
      { input: "Off-Tank", expected: "T2" },
      { input: "side tank", expected: "T2" },
      // Tank
      { input: "tank", expected: "Tank" },
      { input: "tanks", expected: "Tank" },
      { input: "TANK", expected: "Tank" },
      // Pure Healer / Barrier Healer — must precede Healer
      { input: "pure", expected: "PureHealer" },
      { input: "regen", expected: "PureHealer" },
      { input: "Pure Healer", expected: "PureHealer" },
      { input: "regen healer", expected: "PureHealer" },
      { input: "pure heal", expected: "PureHealer" }, // without "er" suffix
      { input: "ph", expected: "PureHealer" },
      { input: "rh", expected: "PureHealer" },
      { input: "barrier", expected: "BarrierHealer" },
      { input: "shield", expected: "BarrierHealer" },
      { input: "Barrier Healer", expected: "BarrierHealer" },
      { input: "shield healer", expected: "BarrierHealer" },
      { input: "bh", expected: "BarrierHealer" },
      { input: "sh", expected: "BarrierHealer" },
      // H1 / H2 — must precede Healer
      { input: "h1", expected: "H1" },
      { input: "healer 1", expected: "H1" },
      { input: "h2", expected: "H2" },
      { input: "Healer 2", expected: "H2" },
      // Healer
      { input: "heals", expected: "Healer" },
      { input: "healer", expected: "Healer" },
      { input: "Healers", expected: "Healer" },
      // Support
      { input: "supp", expected: "Support" },
      { input: "Supps", expected: "Support" },
      { input: "support", expected: "Support" },
      { input: "supports", expected: "Support" },
      // M1 / M2 — must precede MeleeDps
      { input: "m1", expected: "M1" },
      { input: "Melee 1", expected: "M1" },
      { input: "m2", expected: "M2" },
      { input: "melee 2", expected: "M2" },
      // Melee DPS
      { input: "melee", expected: "MeleeDps" },
      { input: "melees", expected: "MeleeDps" },
      { input: "Melee DPS", expected: "MeleeDps" },
      { input: "melee dps", expected: "MeleeDps" },
      // Magic DPS
      { input: "magic dps", expected: "MagicDps" },
      { input: "Magic DPS", expected: "MagicDps" },
      { input: "ranged magic", expected: "MagicDps" },
      { input: "range magic", expected: "MagicDps" },
      { input: "magic ranged", expected: "MagicDps" },
      { input: "Magic Range", expected: "MagicDps" },
      { input: "magic ranged dps", expected: "MagicDps" },
      // Physical Ranged DPS
      { input: "physrange", expected: "PhysRangedDps" },
      { input: "Physranged", expected: "PhysRangedDps" },
      { input: "physical ranged", expected: "PhysRangedDps" },
      { input: "Ranged Physical", expected: "PhysRangedDps" },
      { input: "physical ranged dps", expected: "PhysRangedDps" },
      // R1 / R2 — must precede RangedDps
      { input: "r1", expected: "R1" },
      { input: "ranged 1", expected: "R1" },
      { input: "r2", expected: "R2" },
      { input: "Ranged 2", expected: "R2" },
      // Ranged DPS
      { input: "range", expected: "RangedDps" },
      { input: "ranged", expected: "RangedDps" },
      { input: "Ranged DPS", expected: "RangedDps" },
      { input: "ranged dps", expected: "RangedDps" },
      // D1 / D2 / D3 / D4 — must precede Dps
      { input: "d1", expected: "D1" },
      { input: "dps 1", expected: "D1" },
      { input: "d2", expected: "D2" },
      { input: "DPS 2", expected: "D2" },
      { input: "d3", expected: "D3" },
      { input: "d4", expected: "D4" },
      // DPS
      { input: "dps", expected: "Dps" },
      { input: "DPS", expected: "Dps" },
    ] satisfies Case[])(
      'parseRole("$input") === "$expected"',
      ({ input, expected }) => {
        expect(parseRole(input)).toBe(expected);
      },
    );
  });

  describe("does not match substrings or invalid inputs", () => {
    test.each([
      { input: "stank", expected: null }, // "tank" embedded in word
      { input: "outranged", expected: null }, // "ranged" embedded in word
      { input: "unsupported", expected: null }, // "support" embedded in word
      // Two-letter abbreviations need word boundaries on both sides
      { input: "graph", expected: null }, // ends with "ph" but no leading \b
      { input: "bash", expected: null }, // ends with "sh" but no leading \b
      { input: "she", expected: null }, // starts with "sh" but no trailing \b
      { input: "phase", expected: null }, // starts with "ph" but no trailing \b
      // Ambiguous or unsupported
      { input: "magic", expected: null }, // requires "dps" or ranged qualifier
      { input: "rdps", expected: null }, // no whole-word "r", "dps" match
      { input: "mdps", expected: null }, // no whole-word "m", "dps" match
      { input: "caster", expected: null }, // removed from patterns; too ambiguous
      { input: "paladin", expected: null }, // job, not role
      { input: "orange", expected: null }, // no overlap
      { input: "", expected: null }, // empty string
    ] satisfies Case[])(
      'parseRole("$input") === null',
      ({ input, expected }) => {
        expect(parseRole(input)).toBe(expected);
      },
    );
  });
});
