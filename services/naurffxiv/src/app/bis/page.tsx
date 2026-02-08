import { getBisConfig } from "@/config/bis.config";
import Link from "next/link";
import Image from "next/image";

import type { ReactElement } from "react";
import type { Metadata } from "next";

// Disable caching to fetch fresh data every time
export const revalidate = 0;

export const metadata: Metadata = {
  title: "Best-in-Slot Directory | NAUR",
  description:
    "Comprehensive best-in-slot gear sets for all FFXIV Ultimate Raids. Optimized for every job.",
};

/**
 * BiS Directory Page
 * Route: /bis
 * Lists all available ultimates
 */
export default async function BisDirectoryPage(): Promise<ReactElement> {
  const bisConfig = await getBisConfig();
  const ultimates = Object.values(bisConfig);

  return (
    <div className="flex flex-col items-center">
      {/* Header */}
      <div className="w-full max-w-7xl px-4 pt-12 text-center md:px-0">
        <h1 className="text-3xl font-bold tracking-wider text-white md:text-4xl">
          Best-in-Slot Directory
        </h1>
        <p className="mt-2 text-gray-400">
          Select an Ultimate to view optimized gear sets
        </p>
      </div>

      <div className="w-full max-w-7xl px-4 py-8 md:py-12">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {ultimates.map((ultimate) => (
            <Link
              key={ultimate.abbreviation}
              href={`/bis/${ultimate.abbreviation.toLowerCase()}`}
              className="group block overflow-hidden rounded-lg bg-[#162835]"
            >
              {/* Image Area */}
              <div className="relative h-48 group-hover:brightness-75">
                <Image
                  src={`/images/thumbnails/ultimate/${ultimate.abbreviation.toLowerCase()}.avif`}
                  alt={ultimate.name}
                  fill
                  className="object-cover"
                  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                />
              </div>

              {/* Text Area */}
              <div className="px-3 py-2 text-center">
                <h2 className="text-lg font-medium text-white group-hover:underline">
                  {ultimate.name}
                </h2>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
