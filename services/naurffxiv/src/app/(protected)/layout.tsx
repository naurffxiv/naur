"use client";

import type { FCWithChildren } from "@naur/shared/types";
import WithAuthLayout from "@auth/components/Gates/WithAuthLayout";

const ProtectedLayout: FCWithChildren = ({ children }) => {
  return <WithAuthLayout>{children}</WithAuthLayout>;
};

export default ProtectedLayout;
