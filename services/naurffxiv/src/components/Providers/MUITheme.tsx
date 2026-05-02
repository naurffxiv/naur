"use client";

import type { WithChildren } from "@naur/shared/types";
import type { ReactNode } from "react";
import { ThemeProvider } from "@emotion/react";
import { theme } from "@/config/theme";

export default function MUITheme({ children }: WithChildren): ReactNode {
  return <ThemeProvider theme={theme}>{children}</ThemeProvider>;
}
