import type { FCWithChildren } from "@naur/shared/types";
import AuthSessionProvider from "@auth/components/Gates/AuthSessionProvider";

const AuthLayout: FCWithChildren = ({ children }) => {
  return <AuthSessionProvider>{children}</AuthSessionProvider>;
};

export default AuthLayout;
