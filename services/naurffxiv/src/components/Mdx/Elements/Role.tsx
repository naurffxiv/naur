import Image, { type StaticImageData } from "next/image";
import type { ReactNode } from "react";

import tankIcon from "@/assets/Icons/Roles/tank.png";
import tank1Icon from "@/assets/Icons/Roles/tank_1.png";
import tank2Icon from "@/assets/Icons/Roles/tank_2.png";
import healerIcon from "@/assets/Icons/Roles/healer.png";
import healer1Icon from "@/assets/Icons/Roles/healer_1.png";
import healer2Icon from "@/assets/Icons/Roles/healer_2.png";
import pureHealerIcon from "@/assets/Icons/Roles/pure_healer.png";
import barrierHealerIcon from "@/assets/Icons/Roles/barrier_healer.png";
import supportIcon from "@/assets/Icons/Roles/support.png";
import dpsIcon from "@/assets/Icons/Roles/dps.png";
import dps1Icon from "@/assets/Icons/Roles/dps_1.png";
import dps2Icon from "@/assets/Icons/Roles/dps_2.png";
import dps3Icon from "@/assets/Icons/Roles/dps_3.png";
import dps4Icon from "@/assets/Icons/Roles/dps_4.png";
import meleeIcon from "@/assets/Icons/Roles/melee_dps.png";
import melee1Icon from "@/assets/Icons/Roles/melee_1.png";
import melee2Icon from "@/assets/Icons/Roles/melee_2.png";
import rangedIcon from "@/assets/Icons/Roles/ranged_dps.png";
import ranged1Icon from "@/assets/Icons/Roles/ranged_dps_1.png";
import ranged2Icon from "@/assets/Icons/Roles/ranged_dps_2.png";
import magicRangedIcon from "@/assets/Icons/Roles/magical_ranged_dps.png";
import physicalRangedIcon from "@/assets/Icons/Roles/physical_ranged_dps.png";

// Order matters: more specific patterns must precede general ones so that e.g.
// "tank 1" resolves to T1 rather than Tank, and "pure healer" to PureHealer
// rather than Healer.
const roleConfig = [
  {
    key: "T1",
    regex: /\b(t(ank)? ?1|mt|main ?tank)\b/i,
    className: "hl-tank",
    icon: tank1Icon,
  },
  {
    key: "T2",
    regex: /\b(t(ank)? ?2|[so]t|off[- ]?tank|side ?tank)\b/i,
    className: "hl-tank",
    icon: tank2Icon,
  },
  { key: "Tank", regex: /\btanks?\b/i, className: "hl-tank", icon: tankIcon },
  {
    key: "PureHealer",
    regex: /\b((pure|regen)( ?heal(er)?)?|[pr]h)\b/i,
    className: "hl-healer",
    icon: pureHealerIcon,
  },
  {
    key: "BarrierHealer",
    regex: /\b((barrier|shield)( ?heal(er)?)?|[bs]h)\b/i,
    className: "hl-healer",
    icon: barrierHealerIcon,
  },
  {
    key: "H1",
    regex: /\bh(ealer)? ?1\b/i,
    className: "hl-healer",
    icon: healer1Icon,
  },
  {
    key: "H2",
    regex: /\bh(ealer)? ?2\b/i,
    className: "hl-healer",
    icon: healer2Icon,
  },
  {
    key: "Healer",
    regex: /\bheal(er)?s?\b/i,
    className: "hl-healer",
    icon: healerIcon,
  },
  {
    key: "Support",
    regex: /\bsupp(ort)?s?\b/i,
    className: "hl-support",
    icon: supportIcon,
  },
  {
    key: "M1",
    regex: /\bm(elee( ?dps)?)? ?1\b/i,
    className: "hl-dps",
    icon: melee1Icon,
  },
  {
    key: "M2",
    regex: /\bm(elee( ?dps)?)? ?2\b/i,
    className: "hl-dps",
    icon: melee2Icon,
  },
  {
    key: "MeleeDps",
    regex: /\bmelee( ?dps|s)?\b/i,
    className: "hl-dps",
    icon: meleeIcon,
  },
  {
    key: "MagicDps",
    regex:
      /\b(magic(al)? ?ranged?|ranged? ?magic(al)?)( ?dps)?\b|\bmagic(al)? ?dps\b/i,
    className: "hl-dps",
    icon: magicRangedIcon,
  },
  {
    key: "PhysRangedDps",
    regex: /\b(phys(ical)? ?ranged?|ranged? ?phys(ical)?)( ?dps)?\b/i,
    className: "hl-dps",
    icon: physicalRangedIcon,
  },
  {
    key: "R1",
    regex: /\br(anged?( ?dps)?)? ?1\b/i,
    className: "hl-dps",
    icon: ranged1Icon,
  },
  {
    key: "R2",
    regex: /\br(anged?( ?dps)?)? ?2\b/i,
    className: "hl-dps",
    icon: ranged2Icon,
  },
  {
    key: "RangedDps",
    regex: /\branged?( ?dps)?\b/i,
    className: "hl-dps",
    icon: rangedIcon,
  },
  { key: "D1", regex: /\bd(ps)? ?1\b/i, className: "hl-dps", icon: dps1Icon },
  { key: "D2", regex: /\bd(ps)? ?2\b/i, className: "hl-dps", icon: dps2Icon },
  { key: "D3", regex: /\bd(ps)? ?3\b/i, className: "hl-dps", icon: dps3Icon },
  { key: "D4", regex: /\bd(ps)? ?4\b/i, className: "hl-dps", icon: dps4Icon },
  { key: "Dps", regex: /\bdps\b/i, className: "hl-dps", icon: dpsIcon },
] as const satisfies ReadonlyArray<{
  key: string;
  regex: RegExp;
  className: string;
  icon: StaticImageData;
}>;

export type RoleKey = (typeof roleConfig)[number]["key"];

export function parseRole(text: string): RoleKey | null {
  const trimmed = text.trim();
  for (const entry of roleConfig) {
    if (entry.regex.test(trimmed)) {
      return entry.key;
    }
  }
  return null;
}

type RoleProps = {
  role?: string;
  icon?: boolean | "false";
  children?: ReactNode;
};

function hasChildren(children: ReactNode): boolean {
  return (
    Boolean(children) &&
    (Array.isArray(children) ? children.length > 0 : children !== "")
  );
}

function Role({
  role: roleProp,
  icon: iconProp,
  children,
}: RoleProps): React.JSX.Element {
  const roleText = roleProp || children?.toString() || "";
  const role = parseRole(roleText);

  if (!role) {
    throw new Error(`Could not infer role from: "${roleText}"`);
  }

  const config = roleConfig.find((e) => e.key === role);
  if (!config) {
    throw new Error(`No icon configured for role: "${role}"`);
  }

  const { className, icon } = config;
  const showIcon = iconProp !== false && iconProp !== "false";
  const showText = hasChildren(children);
  // If we are showing both an icon and text, then put a space between.
  // This is U+202F NARROW NO-BREAK SPACE, narrower than a normal space,
  // and it does not allow breaking so that the icon can't be split off
  // from the text at the end of a line.
  const space = showIcon && showText ? "\u202F" : "";
  // If we have text, then treat the icon as just decorative.
  // If not, treat the icon as equivalent to the text in the 'role' argument.
  const iconAlt = showText ? "" : roleText;

  return (
    <mark className={`${className} select-text`}>
      <span className="preserve-line-height">
        {showIcon && (
          <Image
            src={icon}
            alt={iconAlt}
            title={`${roleText} icon`}
            width={64}
            height={64}
            className={`${showText && "select-none"} self-center m-0 w-[0.95lh] h-[0.95lh]`}
          />
        )}
      </span>
      <span className="select-none">{space}</span>
      {children}
    </mark>
  );
}

export default Role;
