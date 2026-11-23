import "@/app/globals.css";

import { roboto } from "@/config/fonts";
import { ReactNode, ReactElement } from "react";

interface MdxLayoutProps {
  children: ReactNode;
}

// note: this can be removed when MDX rendering is moved to @/app/ (NAUR-84)
export default function MdxLayout({ children }: MdxLayoutProps): ReactElement {
  return <div className={roboto.className}>{children}</div>;
}
