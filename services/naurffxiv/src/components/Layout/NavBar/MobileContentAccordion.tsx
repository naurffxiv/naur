import {
  Box,
  Collapse,
  Typography,
  MenuList,
  MenuItem,
  type SxProps,
  type Theme,
} from "@mui/material";
import { useState, type ReactElement } from "react";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";

interface MobileContentAccordionProps {
  data: Array<{ title: string; url: string }>;
  name: string;
}

/**
 * Mobile (thin screen) version of content accordion
 * For mobile accordion inside hamburger menu
 * */
export function MobileContentAccordion({
  data,
  name,
}: MobileContentAccordionProps): ReactElement {
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleClick = (): void => {
    setMobileOpen(!mobileOpen);
  };

  return (
    <Box sx={sx.root}>
      <MenuItem onClick={handleClick}>
        <Box sx={sx.nameContainer}>
          <Typography>{name}</Typography>
          {mobileOpen ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
        </Box>
      </MenuItem>
      <Collapse in={mobileOpen} timeout="auto">
        <MenuList>
          {data.map((fight, i) => (
            <MenuItem
              component="a"
              href={fight.url}
              key={i}
              sx={sx.dropdownItemContainer}
            >
              <Box sx={sx.dropdownItem}>
                <Typography>{fight.title}</Typography>
              </Box>
            </MenuItem>
          ))}
        </MenuList>
      </Collapse>
    </Box>
  );
}

const sx = {
  root: {
    width: "100%",
  },
  nameContainer: {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    width: "100%",
    cursor: "pointer",
  },
  dropdownItemContainer: {
    display: "block",
    "&:hover": {
      backgroundColor: "rgba(0, 0, 0, 0.04)",
    },
  },
  dropdownItem: {
    py: 0.5,
    px: 1,
    display: "block",
    textDecoration: "none",
    color: "inherit",
  },
  dropdownItemText: {
    fontSize: "0.9rem",
  },
} satisfies Record<string, SxProps<Theme>>;
