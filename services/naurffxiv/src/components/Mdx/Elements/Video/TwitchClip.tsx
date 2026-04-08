"use client";
import clsx from "clsx";
import { useEffect, useState, type ReactElement } from "react";
import type { VideoEmbedProps } from "./types";

export default function TwitchClip({
  width = "100%",
  height = "100%",
  videoId,
  className,
}: VideoEmbedProps): ReactElement | "" {
  const [hostname, setHostname] = useState<string>("");

  useEffect(() => {
    setHostname(window.location.hostname);
  }, []);

  if (!hostname) return "";

  return (
    <iframe
      src={`https://clips.twitch.tv/embed?clip=${videoId}&parent=${hostname}`}
      height={height}
      width={width}
      className={clsx("aspect-video", className)}
      allowFullScreen
    />
  );
}
