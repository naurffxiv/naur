import { IconButton } from "@mui/material";
import ClearIcon from "@mui/icons-material/Clear";
import { MouseEventHandler, ReactNode } from "react";

export interface ExitButtonProps {
  onClick?: MouseEventHandler<HTMLButtonElement>;
}

function ExitButton({ onClick }: ExitButtonProps): ReactNode {
  return (
    <IconButton sx={closeDrawer} size="large" onClick={onClick}>
      <ClearIcon sx={{ color: "white" }} />
    </IconButton>
  );
}

const closeDrawer = {
  ":hover": { background: "#00171fa3" },
};

export default ExitButton;
