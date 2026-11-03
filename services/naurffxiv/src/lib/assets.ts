import type { StaticImageData } from "next/image";
import type { SvgIconComponent } from "@mui/icons-material";

// icons
import Calendar from "@/assets/icons/calendar.png";
import Chat from "@/assets/icons/chat.png";
import Contact from "@/assets/icons/contact.png";
import Discord from "@/assets/icons/discord.png";
import Github from "@/assets/icons/github-logo.png";
import Lightbulb from "@/assets/icons/lightbulb.png";
import Naur from "@/assets/icons/naur-icon.png";
import Patreon from "@/assets/icons/patreon.png";
// MUI icons
import PersonSearchIcon from "@mui/icons-material/PersonSearch";

export const icons = {
  Chat,
  Lightbulb,
  Contact,
  Discord,
  Calendar,
  Github,
  Naur,
  Patreon,
} satisfies Record<string, StaticImageData>;

export const images = {
  Pandora: "/images/ultimate/fru/banner.avif",
  m9s: "/images/savage/dawntrail/m9s/banner.avif",
  m10s: "/images/savage/dawntrail/m10s/banner.avif",
  m11s: "/images/savage/dawntrail/m11s/banner.avif",
  m12s: "/images/savage/dawntrail/m12s/banner.avif",
  recollection: "/images/extreme/dawntrail/recollection/banner.avif",
} satisfies Record<string, string>;

export const iconsMui = { PersonSearchIcon } satisfies Record<
  string,
  SvgIconComponent
>;
