// Job Icons
const pld = "/images/ffxiv/jobs/tank/Paladin.png";
const war = "/images/ffxiv/jobs/tank/Warrior.png";
const drk = "/images/ffxiv/jobs/tank/DarkKnight.png";
const gnb = "/images/ffxiv/jobs/tank/Gunbreaker.png";

const whm = "/images/ffxiv/jobs/healer/WhiteMage.png";
const sch = "/images/ffxiv/jobs/healer/Scholar.png";
const ast = "/images/ffxiv/jobs/healer/Astrologian.png";
const sge = "/images/ffxiv/jobs/healer/Sage.png";

const mnk = "/images/ffxiv/jobs/dps/Monk.png";
const drg = "/images/ffxiv/jobs/dps/Dragoon.png";
const nin = "/images/ffxiv/jobs/dps/Ninja.png";
const sam = "/images/ffxiv/jobs/dps/Samurai.png";
const rpr = "/images/ffxiv/jobs/dps/Reaper.png";
const vpr = "/images/ffxiv/jobs/dps/Viper.png";

const brd = "/images/ffxiv/jobs/dps/Bard.png";
const mch = "/images/ffxiv/jobs/dps/Machinist.png";
const dnc = "/images/ffxiv/jobs/dps/Dancer.png";

const blm = "/images/ffxiv/jobs/dps/BlackMage.png";
const smn = "/images/ffxiv/jobs/dps/Summoner.png";
const rdm = "/images/ffxiv/jobs/dps/RedMage.png";
const pct = "/images/ffxiv/jobs/dps/Pictomancer.png";

// Role Icons
const tankRole = "/images/ffxiv/roles/tank.png";
const healerRole = "/images/ffxiv/roles/healer.png";
const meleeRole = "/images/ffxiv/roles/melee.png";
const rangedRole = "/images/ffxiv/roles/ranged.png";
const casterRole = "/images/ffxiv/roles/caster.png";

export const JOB_ICONS: Record<string, string> = {
  // Tanks
  PLD: pld,
  WAR: war,
  DRK: drk,
  GNB: gnb,
  // Healers
  WHM: whm,
  SCH: sch,
  AST: ast,
  SGE: sge,
  // Melee
  MNK: mnk,
  DRG: drg,
  NIN: nin,
  SAM: sam,
  RPR: rpr,
  VPR: vpr,
  // Ranged
  BRD: brd,
  MCH: mch,
  DNC: dnc,
  // Casters
  BLM: blm,
  SMN: smn,
  RDM: rdm,
  PCT: pct,
};

export const ROLE_ICONS = {
  tank: tankRole,
  healer: healerRole,
  melee: meleeRole,
  ranged: rangedRole,
  caster: casterRole,
} as const;

// Standard FFXIV Job Priority Order
export const JOB_PRIORITY = [
  // Tank
  "PLD",
  "WAR",
  "DRK",
  "GNB",
  // Healer
  "WHM",
  "SCH",
  "AST",
  "SGE",
  // Melee
  "MNK",
  "DRG",
  "NIN",
  "SAM",
  "RPR",
  "VPR",
  // Physical Ranged
  "BRD",
  "MCH",
  "DNC",
  // Magical Ranged
  "BLM",
  "SMN",
  "RDM",
  "PCT",
];
