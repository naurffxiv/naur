import localFont from "next/font/local";

export const roboto = localFont({
  src: [
    {
      path: "../fonts/roboto/roboto-latin-wght-normal.woff2",
      weight: "300",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-wght-normal.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-wght-normal.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-wght-normal.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-wght-italic.woff2",
      weight: "300",
      style: "italic",
    },
    {
      path: "../fonts/roboto/roboto-latin-wght-italic.woff2",
      weight: "400",
      style: "italic",
    },
    {
      path: "../fonts/roboto/roboto-latin-wght-italic.woff2",
      weight: "500",
      style: "italic",
    },
    {
      path: "../fonts/roboto/roboto-latin-wght-italic.woff2",
      weight: "700",
      style: "italic",
    },
  ],
  fallback: ["system-ui", "arial"],
  display: "swap",
});
