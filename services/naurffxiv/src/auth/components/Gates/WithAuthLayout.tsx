import AuthGate from "@auth/components/Gates/AuthGate";
import type { FCWithChildren } from "@naur/shared/types";

const WithAuthLayout: FCWithChildren = ({ children }) => {
  return <AuthGate>{children}</AuthGate>;
};

export default WithAuthLayout;
