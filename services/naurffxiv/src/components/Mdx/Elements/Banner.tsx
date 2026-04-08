import Image, { type ImageProps } from "next/image";
import clsx from "clsx";
import { type ReactElement } from "react";

interface BannerProps extends ImageProps {
  left?: boolean;
}

export default function Banner({
  src,
  alt,
  left = false,
  ...props
}: BannerProps): ReactElement {
  return (
    <div className="relative w-full aspect-[3/1] not-prose">
      <Image
        className={clsx("object-cover rounded-sm", {
          "object-left": left,
          "object-center": !left,
        })}
        src={src}
        alt={alt}
        {...props}
        fill
        sizes="100vw"
        priority
        decoding="sync"
      />
    </div>
  );
}
