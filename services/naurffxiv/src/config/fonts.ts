import localFont from "next/font/local";

export const roboto = localFont({
  src: [
    {
      path: "../fonts/roboto/roboto-latin-300-normal.woff2",
      weight: "300",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-300-italic.woff2",
      weight: "300",
      style: "italic",
    },
    {
      path: "../fonts/roboto/roboto-latin-400-normal.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-400-italic.woff2",
      weight: "400",
      style: "italic",
    },
    {
      path: "../fonts/roboto/roboto-latin-500-normal.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-500-italic.woff2",
      weight: "500",
      style: "italic",
    },
    {
      path: "../fonts/roboto/roboto-latin-700-normal.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "../fonts/roboto/roboto-latin-700-italic.woff2",
      weight: "700",
      style: "italic",
    },
  ],
  fallback: ["system-ui", "arial"],
  display: "swap",
});
