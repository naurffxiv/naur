"use client";

import { useRef, useState, type ReactNode, type ReactElement } from "react";
import Fullscreen from "@mui/icons-material/Fullscreen";
import OpenInNew from "@mui/icons-material/OpenInNew";

interface DetailsProps {
  children: ReactNode;
  title: ReactNode;
  href?: string;
  fullscreen?: boolean;
}

function Details({
  children,
  title,
  href,
  fullscreen,
}: DetailsProps): ReactElement {
  const detailsRef = useRef<HTMLDetailsElement>(null);
  const [isOpen, setIsOpen] = useState(false);

  function handleFullscreen(e: React.MouseEvent) {
    e.stopPropagation();
    const iframe = detailsRef.current?.querySelector("iframe");
    if (iframe?.requestFullscreen) {
      iframe.requestFullscreen();
    }
  }

  return (
    <details
      ref={detailsRef}
      className="my-4"
      onToggle={() => setIsOpen(detailsRef.current?.open ?? false)}
    >
      <summary className="flex items-center gap-2 cursor-pointer">
        {href && (
          <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            onClick={(e) => e.stopPropagation()}
            className="opacity-60 hover:opacity-100 shrink-0"
          >
            <OpenInNew fontSize="small" />
          </a>
        )}
        <span className="flex-1">{title}</span>
        {fullscreen && isOpen && (
          <button
            type="button"
            onClick={handleFullscreen}
            className="opacity-60 hover:opacity-100 cursor-pointer shrink-0"
            title="Fullscreen"
          >
            <Fullscreen fontSize="small" />
          </button>
        )}
      </summary>

      <div className="p-2">{children}</div>
    </details>
  );
}

export default Details;
