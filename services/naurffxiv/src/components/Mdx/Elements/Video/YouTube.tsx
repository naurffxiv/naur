import clsx from "clsx";
import { type ReactElement } from "react";
import type { VideoEmbedProps } from "./types";

export default function YouTube({
  width = "100%",
  height = "100%",
  videoId,
  className,
}: VideoEmbedProps): ReactElement {
  return (
    <iframe
      width={width}
      height={height}
      src={`https://www.youtube.com/embed/${videoId}`}
      title="YouTube video player"
      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
      className={clsx("aspect-video", className)}
      allowFullScreen
    />
  );
}
