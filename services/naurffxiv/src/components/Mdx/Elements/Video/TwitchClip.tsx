"use client";
import clsx from "clsx";
import { useSyncExternalStore, type ReactElement } from "react";
import type { VideoEmbedProps } from "./types";

function subscribe(): () => void {
  return () => {};
}

function getHostnameSnapshot(): string {
  return window.location.hostname;
}

function getServerHostnameSnapshot(): string {
  return "";
}

export default function TwitchClip({
  width = "100%",
  height = "100%",
  videoId,
  className,
}: VideoEmbedProps): ReactElement | "" {
  const hostname = useSyncExternalStore(
    subscribe,
    getHostnameSnapshot,
    getServerHostnameSnapshot,
  );

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
