"use client";

import type { FCWithChildren } from "@naur/shared/types";
import AuthSessionProvider from "@auth/components/Gates/AuthSessionProvider";
import WithAuthLayout from "@auth/components/Gates/WithAuthLayout";

const ProtectedLayout: FCWithChildren = ({ children }) => {
  return (
    <AuthSessionProvider>
      <WithAuthLayout>{children}</WithAuthLayout>
    </AuthSessionProvider>
  );
};

export default ProtectedLayout;
