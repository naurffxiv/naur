import { JOB_NAMES, ULTIMATE_INFO } from "@/config/bis.config";
import type { BisConfig } from "@/config/bis.config";

const SHEET_ID = "1KTjZvCpuWsq38MwT5sO-BlntwUa3XlCaEvFcEpNxMNk";

const SHEET_GIDS = {
  BIS_PUBLIC: "0",
} as const;

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
 * Fetch complete BiS configuration from the single staging Google Sheet
 */
export async function getBisConfig(): Promise<BisConfig> {
  try {
    const csv = await fetchSheetCSV(SHEET_GIDS.BIS_PUBLIC);
    const rows = parseCSV(csv);

    if (rows.length === 0) {
      throw new Error("Sheet is empty");
    }

    const header = rows[0].map((cell) => cell.trim().toUpperCase());
    const ultimates = Object.keys(
      ULTIMATE_INFO,
    ) as (keyof typeof ULTIMATE_INFO)[];

    // Dynamically find column indices for each ultimate
    const columnMap: Record<string, number> = {};
    for (const ult of ultimates) {
      const index = header.indexOf(ult);
      if (index !== -1) {
        columnMap[ult] = index;
      }
    }

    const config: BisConfig = {
      FRU: { ...ULTIMATE_INFO.FRU, jobs: {} },
      TOP: { ...ULTIMATE_INFO.TOP, jobs: {} },
      DSR: { ...ULTIMATE_INFO.DSR, jobs: {} },
      TEA: { ...ULTIMATE_INFO.TEA, jobs: {} },
      UWU: { ...ULTIMATE_INFO.UWU, jobs: {} },
      UCOB: { ...ULTIMATE_INFO.UCOB, jobs: {} },
    };

    // Skip header row, process rows 1 onwards
    for (let i = 1; i < rows.length; i++) {
      const row = rows[i];
      const jobAbbr = row[0]?.trim().toUpperCase();

      if (!jobAbbr || !(jobAbbr in JOB_NAMES)) {
        continue;
      }

      const jobInfo = JOB_NAMES[jobAbbr as keyof typeof JOB_NAMES];

      // Map each ultimate column using the dynamic map
      for (const [ultAbbr, colIndex] of Object.entries(columnMap)) {
        const link = row[colIndex]?.trim();

        if (link && (link.startsWith("http") || link.startsWith("https"))) {
          config[ultAbbr as keyof BisConfig].jobs[jobAbbr] = {
            name: jobInfo.name,
            role: jobInfo.role,
            xivGearUrl: link,
            color: jobInfo.color,
          };
        }
      }
    }

    return config;
  } catch (error) {
    console.error("Error fetching BiS config:", error);
    throw new Error("Failed to load BiS configuration from Google Sheets");
  }
}
