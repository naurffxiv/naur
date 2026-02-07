"use client";

import React, { type ReactElement } from "react";
import Link from "next/link";
import Image from "next/image";
import { JOB_ICONS, ROLE_ICONS } from "@/components/BIS/job-icons";
import type { JobInfo } from "@/config/bis.config";

const ROLE_CONFIG = {
  tank: { label: "Tank", icon: ROLE_ICONS.tank, color: "#60a5fa" }, // Blue-400
  healer: { label: "Healer", icon: ROLE_ICONS.healer, color: "#4ade80" }, // Green-400
  melee: { label: "Melee", icon: ROLE_ICONS.melee, color: "#f87171" }, // Red-400
  ranged: { label: "Phys Ranged", icon: ROLE_ICONS.ranged, color: "#fb923c" }, // Orange-400
  caster: { label: "Magic Ranged", icon: ROLE_ICONS.caster, color: "#c084fc" }, // Purple-400
} as const;

interface JobListProps {
  groupedJobs: Record<string, [string, JobInfo][]>;
  canonicalAbbr: string;
}

export default function JobList({
  groupedJobs,
  canonicalAbbr,
}: JobListProps): ReactElement | null {
  return (
    <div className="space-y-4">
      {(Object.keys(ROLE_CONFIG) as Array<keyof typeof ROLE_CONFIG>).map(
        (role) => {
          const jobs = groupedJobs[role];
          if (!jobs || jobs.length === 0) return null;

          const {
            label,
            // icon: roleIcon, // Unused
            color: roleColor,
          } = ROLE_CONFIG[role];

          return (
            <div
              key={role}
              className="flex flex-col md:flex-row md:items-center border border-gray-800 rounded-lg p-4 gap-4"
            >
              {/* Left Column: Role Label */}
              <div className="flex items-center gap-3 w-full md:w-48 shrink-0">
                <span
                  className="font-bold text-sm tracking-wider uppercase"
                  style={{ color: roleColor }}
                >
                  {label}
                </span>
              </div>

              {/* Right Column: Job Chips */}
              <div className="flex flex-col md:flex-row md:flex-wrap gap-2 w-full">
                {jobs.map(([abbr, info]) => {
                  const hasLink = !!info.xivGearUrl;
                  const icon = JOB_ICONS[abbr];
                  const jobColor = info.color || "#ffffff";

                  const content = (
                    <>
                      <div
                        className="w-1.5 self-stretch mr-3"
                        style={{ backgroundColor: jobColor }}
                      />
                      {icon && (
                        <div className="relative h-10 w-10 shrink-0">
                          <Image
                            src={icon}
                            alt={info.name}
                            fill
                            sizes="40px"
                            className="object-contain p-1"
                          />
                        </div>
                      )}
                      <span className="text-sm font-medium text-gray-200 pr-4">
                        {info.name}
                      </span>
                    </>
                  );

                  if (hasLink) {
                    return (
                      <Link
                        key={abbr}
                        href={`/bis/${canonicalAbbr}/${abbr}`}
                        className="flex items-center bg-[#161b22] hover:bg-[#1f2937] border border-gray-800 hover:border-gray-600 rounded overflow-hidden h-10 transition-all w-full md:w-48"
                      >
                        {content}
                      </Link>
                    );
                  }

                  return (
                    <div
                      key={abbr}
                      className="flex items-center bg-[#161b22] border border-gray-800 rounded overflow-hidden h-10 opacity-100 w-full md:w-48 cursor-default"
                    >
                      {content}
                    </div>
                  );
                })}
              </div>
            </div>
          );
        },
      )}
    </div>
  );
}
