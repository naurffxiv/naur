import {
  Button,
  Menu,
  MenuItem,
  Typography,
  Box,
  type SxProps,
  type Theme,
} from "@mui/material";
import { useState, type ReactElement, type MouseEvent } from "react";
import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";

interface ContentDropdownProps {
  data: Array<{ title: string; url: string }>;
  name: string;
}

/**
 * Desktop version of content dropdown menu
 * For desktop dropdown outside hamburger menu
 * */
export function ContentDropdown({
  data,
  name,
}: ContentDropdownProps): ReactElement {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: MouseEvent<HTMLButtonElement>): void => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = (): void => {
    setAnchorEl(null);
  };

  return (
    <>
      <Button
        id="basic-button"
        aria-controls={open ? "basic-menu" : undefined}
        aria-haspopup="true"
        aria-expanded={open ? "true" : undefined}
        sx={sx.button}
        onClick={handleClick}
      >
        <Box sx={sx.nameContainer}>
          <Typography sx={sx.nameText}>{name}</Typography>
          <ArrowDropDownIcon viewBox="2 4 15 15" sx={sx.dropdownIcon} />
        </Box>
      </Button>
      <Menu
        id={"basic-menu"}
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        MenuListProps={{ "aria-labelledby": "basic-button" }}
      >
        {data.map((fight, i) => (
          <li key={i}>
            <MenuItem
              onClick={handleClose}
              sx={sx.dropdownItemContainer}
              component="a"
              href={fight.url}
            >
              <Typography sx={sx.dropdownItemText}>{fight.title}</Typography>
            </MenuItem>
          </li>
        ))}
      </Menu>
    </>
  );
}

const sx = {
  button: {
    color: "white",
    textTransform: "none",
  },
  nameContainer: {
    flexGrow: 1,
    display: { xs: "flex" },
  },
  nameText: {
    textAlign: "center",
  },
  dropdownIcon: {
    fontSize: "15px",
    marginY: "auto",
  },
  dropdownItemContainer: {
    justifyContent: "flex-start",
  },
  dropdownItemText: {
    textAlign: "left",
  },
} satisfies Record<string, SxProps<Theme>>;
