"use client";
import { useEffect, useRef, type ReactElement } from "react";
import Script from "next/script";
import clsx from "clsx";
import type { VideoEmbedProps } from "./types";

interface TwitchEmbedOptions {
  width: string;
  height: string;
  video: string;
  autoplay: string;
  layout: string;
}

declare global {
  interface Window {
    Twitch?: {
      Embed: new (id: string, options: TwitchEmbedOptions) => void;
    };
  }
}

export default function TwitchVoD({
  width = "100%",
  height = "100%",
  videoId,
  className,
}: VideoEmbedProps): ReactElement {
  const embedRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const initTwitch = (): void => {
      if (window.Twitch && window.Twitch.Embed) {
        new window.Twitch.Embed(`twitch-embed-${videoId}`, {
          width: width as string,
          height: height as string,
          video: videoId,
          autoplay: "false",
          layout: "video",
        });
      }
    };

    // Ensure script is loaded before initializing
    if (window.Twitch) {
      initTwitch();
    } else {
      window.addEventListener(`twitchScriptLoaded${videoId}`, initTwitch);
    }

    return (): void => {
      window.removeEventListener(`twitchScriptLoaded${videoId}`, initTwitch);
    };
  }, [videoId, width, height]);

  return (
    <>
      <Script
        src="https://embed.twitch.tv/embed/v1.js"
        strategy="lazyOnload"
        onLoad={() => {
          window.dispatchEvent(new Event(`twitchScriptLoaded${videoId}`));
        }}
      />
      <div
        id={`twitch-embed-${videoId}`}
        ref={embedRef}
        className={clsx("aspect-video", className)}
      ></div>
    </>
  );
}
