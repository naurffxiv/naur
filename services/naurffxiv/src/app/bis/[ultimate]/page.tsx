import { getBisConfig } from "@/config/bis.config";
import Link from "next/link";
import { notFound } from "next/navigation";
import type { ReactElement } from "react";
import UltimateView from "@/components/BIS/UltimateView";

// Disable caching to fetch fresh data every time
export const revalidate = 0;

/**
 * Fight-level job selection page
 * Route: /bis/[ultimate]
 * Displays all jobs for a specific ultimate
 */
export default async function UltimateBisPage({
  params,
}: {
  params: Promise<{ ultimate: string }>;
}): Promise<ReactElement> {
  const { ultimate: ultimateAbbr } = await params;
  const bisConfig = await getBisConfig();

  // Get ultimate config (case-insensitive)
  const ultimateConfig = bisConfig[ultimateAbbr.toUpperCase()];
  const canonicalAbbr = ultimateConfig?.abbreviation?.toUpperCase();

  // If ultimate not found, show 404
  if (!ultimateConfig || !canonicalAbbr) {
    notFound();
  }

  return (
    <div className="flex min-h-screen flex-col items-center px-4 pt-4 pb-12">
      <div className="w-full max-w-7xl">
        {/* Breadcrumb */}
        <nav className="mb-4 text-sm">
          <Link
            href="/bis"
            className="flex items-center gap-2 text-gray-400 transition-colors hover:text-white"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M19 12H5M12 19l-7-7 7-7" />
            </svg>
            <span>Back to Directory</span>
          </Link>
        </nav>

        <UltimateView
          ultimateName={ultimateConfig.name}
          canonicalAbbr={canonicalAbbr}
          jobs={ultimateConfig.jobs}
        />
      </div>
    </div>
  );
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ ultimate: string }>;
}): Promise<import("next").Metadata> {
  const { ultimate } = await params;
  const bisConfig = await getBisConfig();
  const config = bisConfig[ultimate.toUpperCase()];

  if (!config) {
    return {
      title: "Ultimate Not Found | NAUR",
    };
  }

  return {
    title: `${config.name} (${config.abbreviation}) BiS | NAUR`,
    description: `Optimized Best-in-Slot gear sets for ${config.name} in FFXIV.`,
  };
}
