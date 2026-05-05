"use client";

import { type ReactElement } from "react";
import { MenuItem, Typography, Box } from "@mui/material";
import { ContentDropdown } from "./ContentDropdown";
import { MobileContentAccordion } from "./MobileContentAccordion";
import {
  ultimateList,
  savageList,
  criterionList,
  extremeList,
} from "@/config/constants";

interface NavItemBase {
  label: string;
}

interface NavLink extends NavItemBase {
  type: "link";
  href: string;
  external?: boolean;
}

interface NavDropdown extends NavItemBase {
  type: "dropdown";
  data: Array<{ title: string; url: string }>;
}

type NavItem = NavLink | NavDropdown;

interface SharedNavLinksProps {
  isMobile: boolean;
  onClick?: () => void;
}

export default function SharedNavLinks({
  isMobile,
  onClick,
}: SharedNavLinksProps): ReactElement {
  const navItems: NavItem[] = [
    { type: "link", label: "Home", href: "/" },
    {
      type: "link",
      label: "WTFDIG",
      href: "https://wtfdig.info",
      external: true,
    },
    { type: "dropdown", label: "Ultimate", data: ultimateList },
    { type: "dropdown", label: "Savage", data: savageList },
    { type: "dropdown", label: "Criterion", data: criterionList },
    { type: "dropdown", label: "Extreme", data: extremeList },
  ];

  return (
    <>
      {navItems.map((item, idx) => {
        if (item.type === "link") {
          return (
            <li key={idx}>
              <MenuItem
                component="a"
                href={item.href}
                onClick={onClick}
                target={item.external ? "_blank" : undefined}
                rel={item.external ? "noopener noreferrer" : undefined}
              >
                <Typography
                  sx={{
                    width: "100%",
                    textAlign: isMobile ? "left" : "center",
                  }}
                >
                  {item.label}
                </Typography>
              </MenuItem>
            </li>
          );
        }

        if (item.type === "dropdown") {
          return (
            <Box key={idx} sx={{ width: "100%" }}>
              {isMobile ? (
                <MobileContentAccordion name={item.label} data={item.data} />
              ) : (
                <ContentDropdown name={item.label} data={item.data} />
              )}
            </Box>
          );
        }

        return null;
      })}
    </>
  );
}
