"use client";

import { ContentCopy } from "@mui/icons-material";
import { useRef, ReactNode, ReactElement } from "react";

interface CopyToClipboardProps {
  children: ReactNode;
}

export const CopyToClipboard = ({
  children,
}: CopyToClipboardProps): ReactElement => {
  const textInput = useRef<HTMLDivElement>(null);

  const onCopy = (): void => {
    if (textInput.current !== null && textInput.current.textContent !== null)
      navigator.clipboard.writeText(textInput.current.textContent);
  };

  return (
    <div ref={textInput} className="relative">
      <button
        aria-label="Copy code"
        type="button"
        className="absolute w-10 h-10 p-1 transition-colors rounded right-2 top-2 hover:bg-gray-700 active:bg-gray-600"
        onClick={onCopy}
      >
        <ContentCopy />
      </button>
      {children}
    </div>
  );
};
