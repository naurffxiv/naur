"use client";

import type { FCWithChildren } from "@naur/shared/types";
import { SessionProvider } from "next-auth/react";

/**
 * Provides NextAuth session context to all client components.
 * Enables `useSession()` access across the app.
 * Must wrap the root layout.
 * https://next-auth.js.org/getting-started/client#sessionprovider
 */
const AuthSessionProvider: FCWithChildren = ({ children }) => {
  return <SessionProvider>{children}</SessionProvider>;
};

export default AuthSessionProvider;
