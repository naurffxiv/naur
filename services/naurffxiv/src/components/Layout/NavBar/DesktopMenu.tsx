import { MenuList } from "@mui/material";
import { type ReactElement } from "react";
import SharedNavLinks from "./SharedNavLinks";

/**
 * Desktop variant of the header navbar menu
 * */
export function DesktopMenu(): ReactElement {
  return (
    <MenuList sx={{ display: "flex" }}>
      <SharedNavLinks isMobile={false} />
    </MenuList>
  );
}
