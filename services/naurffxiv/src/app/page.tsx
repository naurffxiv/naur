"use client";

import { Box } from "@mui/material";
import Hero from "@/components/Homepage/Sections/Hero";
import HomepageContent from "@/components/Homepage/Sections/HomepageContent";
import type { ReactElement } from "react";

export default function Home(): ReactElement {
  return (
    <Box>
      <Box
        sx={{
          minHeight: "100vh",
          width: "100%",
        }}
      >
        <Hero />
        <HomepageContent />
      </Box>
    </Box>
  );
}
