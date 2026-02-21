/**
 * BiS (Best-in-Slot) Configuration
 * Type definitions and job metadata for Google Sheets integration
 */

// Type Definitions
export type JobRole = "tank" | "healer" | "melee" | "ranged" | "caster";

export interface JobInfo {
  name: string;
  role: JobRole;
  xivGearUrl: string;
  color?: string; // Derived from JOB_NAMES
}

export interface UltimateConfig {
  name: string;
  abbreviation: string;
  jobs: Record<string, JobInfo>;
}

export type BisConfig = Record<string, UltimateConfig>;

export const ULTIMATE_INFO = {
  FRU: { name: "Futures Rewritten", abbreviation: "FRU" },
  TOP: { name: "The Omega Protocol", abbreviation: "TOP" },
  DSR: { name: "Dragonsong's Reprise", abbreviation: "DSR" },
  TEA: { name: "The Epic of Alexander", abbreviation: "TEA" },
  UWU: { name: "The Weapon's Refrain", abbreviation: "UWU" },
  UCOB: { name: "The Unending Coil of Bahamut", abbreviation: "UCOB" },
} as const;

// Job Definitions
export const JOB_NAMES: Record<
  string,
  { name: string; role: JobRole; color: string }
> = {
  // Tanks
  PLD: { name: "Paladin", role: "tank", color: "#A8D2E6" },
  WAR: { name: "Warrior", role: "tank", color: "#CF2621" },
  DRK: { name: "Dark Knight", role: "tank", color: "#D126CC" },
  GNB: { name: "Gunbreaker", role: "tank", color: "#796D30" },
  // Healers
  WHM: { name: "White Mage", role: "healer", color: "#FFF0F5" },
  SCH: { name: "Scholar", role: "healer", color: "#8657FF" },
  AST: { name: "Astrologian", role: "healer", color: "#FFE74A" },
  SGE: { name: "Sage", role: "healer", color: "#80A0F0" },
  // Melee DPS
  MNK: { name: "Monk", role: "melee", color: "#d69c00" },
  DRG: { name: "Dragoon", role: "melee", color: "#4164CD" },
  NIN: { name: "Ninja", role: "melee", color: "#AF1964" },
  SAM: { name: "Samurai", role: "melee", color: "#E46D04" },
  RPR: { name: "Reaper", role: "melee", color: "#965a90" },
  VPR: { name: "Viper", role: "melee", color: "#108210" },
  // Ranged DPS
  BRD: { name: "Bard", role: "ranged", color: "#91BA5E" },
  MCH: { name: "Machinist", role: "ranged", color: "#6EE1D6" },
  DNC: { name: "Dancer", role: "ranged", color: "#E2B0AF" },
  // Casters
  BLM: { name: "Black Mage", role: "caster", color: "#A579D6" },
  SMN: { name: "Summoner", role: "caster", color: "#2D9B78" },
  RDM: { name: "Red Mage", role: "caster", color: "#E87B7B" },
  PCT: { name: "Pictomancer", role: "caster", color: "#e36494" },
};

// Helper to get BiS configuration
export { getBisConfig } from "@/lib/sheets/bis-fetcher";
