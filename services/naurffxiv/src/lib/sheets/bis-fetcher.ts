import { JOB_NAMES } from "@/config/bis.config";
import type { BisConfig, JobInfo, UltimateConfig } from "@/config/bis.config";

const SHEET_ID = "1V5HIN_yUxc-PIQQDv8zKkQ9TQNZ3OxQc66AxKIC_KaQ";

const SHEET_GIDS = {
  FRU: "1484629370",
  TOP: "519890421",
  DSR: "1992849386",
  TEA: "2057135619",
  UWU: "469937442",
  UCOB: "615990936",
} as const;

const ULTIMATE_INFO = {
  FRU: { name: "Futures Rewritten", abbreviation: "FRU" },
  TOP: { name: "The Omega Protocol", abbreviation: "TOP" },
  DSR: { name: "Dragonsong's Reprise", abbreviation: "DSR" },
  TEA: { name: "The Epic of Alexander", abbreviation: "TEA" },
  UWU: { name: "The Weapon's Refrain", abbreviation: "UWU" },
  UCOB: { name: "The Unending Coil of Bahamut", abbreviation: "UCOB" },
} as const;

interface SheetRow {
  job: string;
  gcdSpeed: string;
  level: string;
  link: string;
  notes: string;
}

/**
 * Fetch CSV data from Google Sheets
 */
async function fetchSheetCSV(gid: string): Promise<string> {
  const url = `https://docs.google.com/spreadsheets/d/${SHEET_ID}/export?format=csv&gid=${gid}`;

  try {
    const response = await fetch(url, {
      next: { revalidate: 0 }, // Disable caching
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch sheet: ${response.statusText}`);
    }

    return await response.text();
  } catch (error) {
    console.error(`Error fetching sheet with gid ${gid}:`, error);
    throw error;
  }
}

/**
 * Parse CSV text into rows
 */
function parseCSV(csv: string): string[][] {
  const lines = csv.split(/\r?\n/);
  return lines.map((line) => {
    const values: string[] = [];
    let current = "";
    let inQuotes = false;

    for (let i = 0; i < line.length; i++) {
      const char = line[i];

      if (char === '"') {
        // Handle escaped quotes (RFC 4180: "" inside quotes = literal ")
        if (inQuotes && line[i + 1] === '"') {
          current += '"';
          i++; // Skip next quote
        } else {
          inQuotes = !inQuotes;
        }
      } else if (char === "," && !inQuotes) {
        values.push(current.trim());
        current = "";
      } else {
        current += char;
      }
    }
    values.push(current.trim());

    return values;
  });
}

/**
 * Extract job entries from parsed CSV rows
 */
function extractJobEntries(rows: string[][]): SheetRow[] {
  const jobEntries: SheetRow[] = [];

  for (const row of rows) {
    // Data starts at column D (index 3)
    const job = row[3]?.trim();
    const gcdSpeed = row[4]?.trim();
    const level = row[5]?.trim();
    const link = row[6]?.trim();
    const notes = row[7]?.trim();

    // Skip header rows and empty rows
    // Allow missing link if notes starts with "Synced Gear"
    const isSyncedGear = notes?.toLowerCase().startsWith("synced gear");
    if (!job || job === "Job" || (!link && !isSyncedGear)) {
      continue;
    }

    // Clean notes: remove URLs and trim
    let cleanNotes = notes || "";
    if (cleanNotes.includes("http")) {
      // Remove everything from http onwards if it looks like a URL
      cleanNotes = cleanNotes.replace(/https?:\/\/\S+/g, "").trim();
    }

    jobEntries.push({
      job,
      gcdSpeed: gcdSpeed || "N/A",
      level: level || "100",
      link: link || "",
      notes: cleanNotes,
    });
  }

  return jobEntries;
}

/**
 * Transform sheet rows into BiS jobs object
 */
function transformToJobs(entries: SheetRow[]): Record<string, JobInfo> {
  const jobs: Record<string, JobInfo> = {};

  for (const entry of entries) {
    const jobAbbr = entry.job.toUpperCase();

    // Only include jobs that exist in JOB_NAMES (skip special builds)
    if (!(jobAbbr in JOB_NAMES)) {
      continue;
    }

    const jobInfo = JOB_NAMES[jobAbbr as keyof typeof JOB_NAMES];

    // If we already have a set for this job, skip subsequent ones (First one wins)
    if (jobAbbr in jobs) {
      continue;
    }

    jobs[jobAbbr] = {
      name: jobInfo.name,
      role: jobInfo.role,
      xivGearUrl: entry.link,
      ...(entry.gcdSpeed &&
        entry.gcdSpeed !== "N/A" &&
        !entry.gcdSpeed.toLowerCase().includes("multiple") && {
          gcdSpeed: entry.gcdSpeed,
        }),
      ...(entry.notes && { notes: entry.notes }),
      color: jobInfo.color,
    };
  }

  return jobs;
}

/**
 * Fetch BiS data for a specific ultimate
 */
async function fetchUltimateData(
  ultimate: keyof typeof SHEET_GIDS,
): Promise<UltimateConfig> {
  const gid = SHEET_GIDS[ultimate];
  const csv = await fetchSheetCSV(gid);
  const rows = parseCSV(csv);
  const entries = extractJobEntries(rows);
  const jobs = transformToJobs(entries);

  return {
    ...ULTIMATE_INFO[ultimate],
    jobs,
  };
}

/**
 * Fetch complete BiS configuration from all Google Sheets
 */
export async function getBisConfig(): Promise<BisConfig> {
  try {
    const [fru, top, dsr, tea, uwu, ucob] = await Promise.all([
      fetchUltimateData("FRU"),
      fetchUltimateData("TOP"),
      fetchUltimateData("DSR"),
      fetchUltimateData("TEA"),
      fetchUltimateData("UWU"),
      fetchUltimateData("UCOB"),
    ]);

    return {
      FRU: fru,
      TOP: top,
      DSR: dsr,
      TEA: tea,
      UWU: uwu,
      UCOB: ucob,
    };
  } catch (error) {
    console.error("Error fetching BiS config:", error);
    throw new Error("Failed to load BiS configuration from Google Sheets");
  }
}
