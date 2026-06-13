"use client";

import { ReactElement, useMemo } from "react";
import Image from "next/image";
import Link from "next/link";
import {
  JOB_ICONS,
  ROLE_ICONS,
  JOB_PRIORITY,
} from "@/components/BIS/job-icons";
import type { JobInfo } from "@/config/bis.config";
import Banner from "@/components/Mdx/Elements/Banner";

interface UltimateViewProps {
  ultimateName: string;
  canonicalAbbr: string;
  jobs: Record<string, JobInfo>;
}

type RoleConfigEntry = {
  label: string;
  color: string;
  icon?: string;
};

const ROLE_CONFIG: Record<string, RoleConfigEntry> = {
  tank: { label: "Tank", icon: ROLE_ICONS.tank, color: "#60a5fa" }, // Blue
  healer: { label: "Healer", icon: ROLE_ICONS.healer, color: "#4ade80" }, // Green
  melee: { label: "Melee", icon: ROLE_ICONS.melee, color: "#f87171" }, // Red
  ranged: { label: "Phys Ranged", icon: ROLE_ICONS.ranged, color: "#fb923c" }, // Orange
  caster: { label: "Magic Ranged", icon: ROLE_ICONS.caster, color: "#c084fc" }, // Purple
};

type FilterRole = keyof typeof ROLE_CONFIG;

const JobCard = ({
  abbr,
  info,
  canonicalAbbr,
}: {
  abbr: string;
  info: JobInfo;
  canonicalAbbr: string;
}): ReactElement => {
  const hasLink = !!info.xivGearUrl;

  const CardContent = (
    <>
      <div className="absolute inset-0 bg-white/5 opacity-100 transition-opacity duration-300 group-hover:opacity-0" />
      <div className="absolute inset-0 bg-white/10 opacity-0 transition-opacity duration-300 group-hover:opacity-100" />

      <div className="relative flex h-full items-center p-3 pl-4 md:p-2 md:pl-6">
        <div className="relative mr-3 h-10 w-10 shrink-0 overflow-hidden transition-transform duration-300 group-hover:scale-110 md:mr-4 md:h-12 md:w-12">
          {JOB_ICONS[abbr] && (
            <Image
              src={JOB_ICONS[abbr]}
              alt={info.name}
              fill
              className="object-cover"
            />
          )}
        </div>

        <div className="flex flex-col justify-center">
          <span className="font-bold text-gray-200 transition-colors group-hover:text-white text-sm md:text-base">
            {info.name}
          </span>
        </div>

        {hasLink && (
          <div className="ml-auto text-gray-600 transition-all duration-300 group-hover:translate-x-1 group-hover:text-cyan-400">
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </div>
        )}
      </div>
    </>
  );

  if (hasLink) {
    return (
      <Link
        href={`/bis/${canonicalAbbr}/${abbr}`}
        target="_blank"
        rel="noopener noreferrer"
        className="group relative block h-16 w-full overflow-hidden rounded-xl border border-white/5 shadow-lg transition-all duration-300 hover:border-white/20 hover:-translate-y-1 hover:shadow-cyan-900/10 md:h-16"
      >
        {CardContent}
      </Link>
    );
  }

  return (
    <div className="group relative h-16 w-full cursor-not-allowed overflow-hidden rounded-xl border border-white/5 bg-white/5 opacity-40 grayscale transition-all duration-300 hover:opacity-50 md:h-16">
      {CardContent}
    </div>
  );
};

const RoleSection = ({
  role,
  roleJobs,
  canonicalAbbr,
}: {
  role: FilterRole;
  roleJobs: [string, JobInfo][];
  canonicalAbbr: string;
}): ReactElement | null => {
  if (roleJobs.length === 0) return null;

  return (
    <div className="w-full">
      {/* Role Header */}
      <div className="mb-3 flex items-center gap-2 border-b border-white/10 pb-2">
        {ROLE_CONFIG[role].icon && (
          <div className="relative h-5 w-5 opacity-70">
            <Image
              src={ROLE_CONFIG[role].icon!}
              alt=""
              fill
              className="object-contain"
            />
          </div>
        )}
        <h2 className="text-lg font-bold tracking-tight text-gray-200">
          {ROLE_CONFIG[role].label}
        </h2>
      </div>
      {/* Role Grid */}
      <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2">
        {roleJobs.map(([abbr, info]) => (
          <JobCard
            key={abbr}
            abbr={abbr}
            info={info}
            canonicalAbbr={canonicalAbbr}
          />
        ))}
      </div>
    </div>
  );
};

export default function UltimateView({
  ultimateName,
  canonicalAbbr,
  jobs,
}: UltimateViewProps): ReactElement {
  const jobsByRole = useMemo(() => {
    const groups: Record<string, [string, JobInfo][]> = {};

    Object.entries(jobs).forEach(([abbr, info]) => {
      const role = info.role;
      if (!groups[role]) {
        groups[role] = [];
      }
      groups[role].push([abbr, info]);
    });

    // Sort each group
    Object.keys(groups).forEach((role) => {
      groups[role].sort((a, b) => {
        const indexA = JOB_PRIORITY.indexOf(a[0]);
        const indexB = JOB_PRIORITY.indexOf(b[0]);
        return (indexA === -1 ? 999 : indexA) - (indexB === -1 ? 999 : indexB);
      });
    });

    return groups;
  }, [jobs]);

  return (
    <div className="flex w-full flex-col items-center">
      {/* Hero Section */}
      <div className="mb-8 w-full max-w-[90ch] md:mb-10">
        <Banner
          src={`/images/thumbnails/ultimate/${canonicalAbbr.toLowerCase()}.avif`}
          alt={`${ultimateName} Banner`}
        />
        <div className="mt-6 text-center md:mt-8">
          <h1 className="mb-2 text-3xl font-black tracking-tight text-white sm:text-5xl md:mb-4 md:text-6xl">
            {ultimateName}
          </h1>
          <p className="mx-auto max-w-xl text-base text-gray-400 md:text-lg">
            Select your job to access the optimized Best-in-Slot gear set.
          </p>
        </div>
      </div>

      {/* Content Area - Split Layout */}
      <div className="w-full max-w-[1400px]">
        <div className="grid grid-cols-1 gap-8 lg:grid-cols-2 lg:gap-12">
          {/* Left Column: Tank & Melee */}
          <div className="flex flex-col gap-8">
            <RoleSection
              role="tank"
              roleJobs={jobsByRole["tank"] || []}
              canonicalAbbr={canonicalAbbr}
            />
            <RoleSection
              role="melee"
              roleJobs={jobsByRole["melee"] || []}
              canonicalAbbr={canonicalAbbr}
            />
          </div>

          {/* Right Column: Healer & Ranged */}
          <div className="flex flex-col gap-8">
            <RoleSection
              role="healer"
              roleJobs={jobsByRole["healer"] || []}
              canonicalAbbr={canonicalAbbr}
            />
            <RoleSection
              role="ranged"
              roleJobs={jobsByRole["ranged"] || []}
              canonicalAbbr={canonicalAbbr}
            />
            <RoleSection
              role="caster"
              roleJobs={jobsByRole["caster"] || []}
              canonicalAbbr={canonicalAbbr}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
