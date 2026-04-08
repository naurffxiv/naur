import clsx from "clsx";
import { type ReactElement } from "react";
import type { VideoEmbedProps } from "./types";

export default function Streamable({
  width = "100%",
  height = "100%",
  videoId,
  className,
}: VideoEmbedProps): ReactElement {
  return (
    <iframe
      src={`https://streamable.com/e/${videoId}?quality=highest`}
      width={width}
      height={height}
      className={clsx("aspect-video", className)}
      allowFullScreen
    />
  );
}
