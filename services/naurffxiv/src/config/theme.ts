import { createTheme } from "@mui/material/styles";
import { roboto } from "./fonts";

export const theme = createTheme({
  typography: {
    fontFamily: roboto.style.fontFamily,
  },
});
