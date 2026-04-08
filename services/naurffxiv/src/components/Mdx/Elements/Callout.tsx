import { type ReactNode, type ReactElement } from "react";
import { Box } from "@mui/material";

interface CalloutTypeConfig {
  text: string;
  backgroundColor: string;
  borderColor: string;
  icon: string;
}

const types = {
  warning: {
    text: "Warning",
    backgroundColor: "#000",
    borderColor: "#FFA500",
    icon: "⚠️",
  },
  tip: {
    text: "Tip",
    backgroundColor: "#000",
    borderColor: "#32CD32",
    icon: "💡",
  },
  default: {
    text: "Note",
    backgroundColor: "#000",
    borderColor: "#28506E",
    icon: "ℹ️",
  },
} satisfies Record<string, CalloutTypeConfig>;

interface CalloutProps {
  children: ReactNode;
  type: keyof typeof types;
}

export default function Callout({
  children,
  type,
}: CalloutProps): ReactElement {
  return (
    <Box
      sx={{
        backgroundColor: types[type]
          ? types[type].backgroundColor
          : types.default.backgroundColor,
        borderLeft: `4px solid ${types[type] ? types[type].borderColor : types.default.borderColor}`,
        padding: 2,
        margin: 2,
        borderRadius: 1,
      }}
      className="callout"
    >
      <span className="font-bold">
        {types[type]
          ? `${types[type].icon} ${types[type].text}`
          : `${types.default.icon} ${types.default.text}`}
      </span>
      {children}
    </Box>
  );
}
