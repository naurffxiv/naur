import type { StaticImageData } from "next/image";
import type { SvgIconComponent } from "@mui/icons-material";

// icons
import Calendar from "@/assets/Icons/Calendar.png";
import Chat from "@/assets/Icons/Chat.png";
import Contact from "@/assets/Icons/Contact.png";
import Discord from "@/assets/Icons/Discord.png";
import Github from "@/assets/Icons/githublogo.png";
import Lightbulb from "@/assets/Icons/Lightbulb.png";
import Naur from "@/assets/Icons/naur_icon.png";
import Pandora from "@/assets/Images/Pandora.avif";
import Patreon from "@/assets/Icons/Patreon.png";
// MUI icons
import PersonSearchIcon from "@mui/icons-material/PersonSearch";
// homepage card images
import m9s from "@/assets/Images/m9s.avif";
import m10s from "@/assets/Images/m10s.avif";
import m11s from "@/assets/Images/m11s.avif";
import m12s from "@/assets/Images/m12s.avif";
import recollection from "@/assets/Images/recollection.avif";

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
  Pandora,
  m9s,
  m10s,
  m11s,
  m12s,
  recollection,
} satisfies Record<string, StaticImageData>;

export const iconsMui = { PersonSearchIcon } satisfies Record<
  string,
  SvgIconComponent
>;
