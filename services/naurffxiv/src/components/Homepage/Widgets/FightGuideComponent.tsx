import { JSX } from "react";
import Image, { StaticImageData } from "next/image";
import clsx from "clsx";

export interface FightGuideEntry {
  title: string;
  url: string;
  img?: StaticImageData | string;
  alt?: string;
}

interface FightGuideComponentProps {
  entries: FightGuideEntry[];
  title: string;
  left?: boolean;
}

export default function FightGuideComponent({
  entries,
  title,
  left,
}: FightGuideComponentProps): JSX.Element {
  const validEntries = entries.filter(
    (
      entry,
    ): entry is FightGuideEntry & {
      img: StaticImageData | string;
      alt: string;
    } => !!(entry.img && entry.alt),
  );

  if (validEntries.length === 0) return <></>;

  return (
    <div>
      <h3 className="mt-12 text-2xl text-center">{title}</h3>
      <div
        className={clsx(
          "mt-8 grid gap-x-10 gap-y-4 mx-auto",
          validEntries.length === 1
            ? "grid-cols-1 justify-center"
            : "grid-cols-1 sm:grid-cols-2",
        )}
      >
        {validEntries.map(
          (
            entry: FightGuideEntry & {
              img: StaticImageData | string;
              alt: string;
            },
          ): JSX.Element => (
            <div
              key={entry.url}
              className="bg-[#162835] overflow-hidden rounded-lg group"
            >
              <a href={entry.url}>
                <div className="relative h-48 group-hover:brightness-75">
                  <Image
                    src={entry.img}
                    fill
                    alt={entry.alt}
                    className={clsx("object-cover", {
                      "object-left": left,
                      "object-center": !left,
                    })}
                  />
                </div>
                <p className="px-3 py-2 text-lg font-medium text-center group-hover:underline">
                  {entry.title}
                </p>
              </a>
            </div>
          ),
        )}
      </div>
    </div>
  );
}
