import "@/app/globals.css";

import { Roboto } from "next/font/google";
import { ReactNode, ReactElement } from "react";

interface MdxLayoutProps {
  children: ReactNode;
}

// note: this can be removed when MDX rendering is moved to @/app/ (NAUR-84)
const roboto = Roboto({
  weight: ["300", "400", "500", "700"],
  style: ["normal", "italic"],
  subsets: ["latin"],
  fallback: ["system-ui", "arial"],
  display: "swap",
});

export default function MdxLayout({ children }: MdxLayoutProps): ReactElement {
  return <div className={roboto.className}>{children}</div>;
}
