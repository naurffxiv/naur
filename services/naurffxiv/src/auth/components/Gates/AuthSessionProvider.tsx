"use client";

import type { FCWithChildren } from "@naur/shared/types";
import { SessionProvider } from "next-auth/react";

/**
 * Provides NextAuth session context to client components.
 * Enables `useSession()` in the protected and /auth layouts that wrap it.
 * Kept off public pages so they don't fetch /api/auth/session on every load.
 * https://next-auth.js.org/getting-started/client#sessionprovider
 */
const AuthSessionProvider: FCWithChildren = ({ children }) => {
  return <SessionProvider>{children}</SessionProvider>;
};

export default AuthSessionProvider;
