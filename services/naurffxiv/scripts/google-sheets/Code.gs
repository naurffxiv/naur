/**
 * Code.gs - BiS Management System
 * Handles secure updates from a Google Form to a locked master sheet.
 */

// --- CONFIGURATION ---
const CONFIG = {
  // Sheet Names
  SHEET_PUBLIC: "BIS_Public", // The Target Sheet (Updated by Form)
  SHEET_ADMIN: "Admin_Config", // Database of Roles

  // The names of your Google Form Questions (Case Sensitive!)
  FORM_Q_FIGHT: "Fight",
  FORM_Q_JOB: "Job",
  FORM_Q_LINK: "BiS Link",
};

/**
 * TRIGGER: Form Submission
 * Set up in Apps Script Triggers -> "From spreadsheet" -> "On form submit"
 */
function onFormSubmit(e) {
  try {
    // 1. Parse Form Data
    // e.namedValues uses the exact Question Titles from your Form
    const responses = e.namedValues;

    // "Email Address" is added automatically by Google if you check 'Collect Emails'
    const email = responses["Email Address"]
      ? responses["Email Address"][0]
      : "";

    // Get Question Answers
    const fight = responses[CONFIG.FORM_Q_FIGHT]
      ? responses[CONFIG.FORM_Q_FIGHT][0].trim().toUpperCase()
      : "";
    const job = responses[CONFIG.FORM_Q_JOB]
      ? responses[CONFIG.FORM_Q_JOB][0].trim().toUpperCase()
      : "";
    const link = responses[CONFIG.FORM_Q_LINK]
      ? responses[CONFIG.FORM_Q_LINK][0].trim()
      : "";

    Logger.log(`Form Recieved from: ${email} for ${fight}/${job}`);

    // 2. Validation
    if (!email) {
      Logger.log(
        "Error: No email collected. Ensure Form Settings > 'Collect email addresses' is VERIFIED.",
      );
      return;
    }

    if (!fight || !job || !link) {
      Logger.log("Error: Missing Fight, Job, or Link data.");
      return;
    }

    // URL Validation: Restrict to known gear planners
    if (
      !link.startsWith("https://xivgear.app/") &&
      !link.startsWith("https://etro.gg/")
    ) {
      throw new Error(
        "Invalid Link (must start with https://xivgear.app/ or https://etro.gg/)",
      );
    }

    // 3. Security Check
    if (checkPermission(email, fight, job)) {
      // 4. Update the Sheet
      updatePublicSheet(fight, job, link);
      Logger.log(`Success: Updated ${fight}/${job} by ${email}`);
    } else {
      Logger.log(`Permission Denied: ${email} tried to edit ${fight}/${job}`);
      // Optional: You could email them back here using MailApp.sendEmail()
    }
  } catch (err) {
    Logger.log(`System Error: ${err.message}`);
    throw err;
  }
}

/**
 * Permission check with role-based access control
 */
function checkPermission(email, fight, job) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const configSheet = ss.getSheetByName(CONFIG.SHEET_ADMIN);

  if (!configSheet) {
    throw new Error(`Configuration sheet '${CONFIG.SHEET_ADMIN}' not found.`);
  }

  // Role definitions
  const ROLE_DEFINITIONS = {
    // Global roles
    ADMIN: { fights: "*", jobs: "*" },
    TANK: { fights: "*", jobs: ["PLD", "WAR", "DRK", "GNB"] },
    HEALER: { fights: "*", jobs: ["WHM", "SCH", "AST", "SGE"] },
    MELEE: { fights: "*", jobs: ["MNK", "DRG", "NIN", "SAM", "RPR", "VPR"] },
    PRANGE: { fights: "*", jobs: ["BRD", "MCH", "DNC"] },
    CASTER: { fights: "*", jobs: ["BLM", "SMN", "RDM", "PCT"] },

    // Fight-specific leads
    FRU_LEAD: { fights: ["FRU"], jobs: "*" },
    TOP_LEAD: { fights: ["TOP"], jobs: "*" },
    DSR_LEAD: { fights: ["DSR"], jobs: "*" },
    TEA_LEAD: { fights: ["TEA"], jobs: "*" },
    UWU_LEAD: { fights: ["UWU"], jobs: "*" },
    UCOB_LEAD: { fights: ["UCOB"], jobs: "*" },

    // Individual jobs - Tanks
    PLD: { fights: "*", jobs: ["PLD"] },
    WAR: { fights: "*", jobs: ["WAR"] },
    DRK: { fights: "*", jobs: ["DRK"] },
    GNB: { fights: "*", jobs: ["GNB"] },

    // Individual jobs - Healers
    WHM: { fights: "*", jobs: ["WHM"] },
    SCH: { fights: "*", jobs: ["SCH"] },
    AST: { fights: "*", jobs: ["AST"] },
    SGE: { fights: "*", jobs: ["SGE"] },

    // Individual jobs - Melee DPS
    MNK: { fights: "*", jobs: ["MNK"] },
    DRG: { fights: "*", jobs: ["DRG"] },
    NIN: { fights: "*", jobs: ["NIN"] },
    SAM: { fights: "*", jobs: ["SAM"] },
    RPR: { fights: "*", jobs: ["RPR"] },
    VPR: { fights: "*", jobs: ["VPR"] },

    // Individual jobs - Physical Ranged DPS
    BRD: { fights: "*", jobs: ["BRD"] },
    MCH: { fights: "*", jobs: ["MCH"] },
    DNC: { fights: "*", jobs: ["DNC"] },

    // Individual jobs - Magical Ranged DPS
    BLM: { fights: "*", jobs: ["BLM"] },
    SMN: { fights: "*", jobs: ["SMN"] },
    RDM: { fights: "*", jobs: ["RDM"] },
    PCT: { fights: "*", jobs: ["PCT"] },
  };

  // Fetch user roles
  const lastRow = configSheet.getLastRow();
  if (lastRow < 2) {
    Logger.log(`Admin_Config has no user data`);
    return false;
  }

  const data = configSheet.getRange(2, 1, lastRow - 1, 2).getValues();
  const userRow = data.find(
    (r) => r[0].toString().toLowerCase() === email.toLowerCase(),
  );

  if (!userRow) {
    Logger.log(`User not found in Admin_Config: ${email}`);
    return false;
  }

  // Parse roles (comma-separated)
  const userRoles = userRow[1]
    .toString()
    .split(",")
    .map((r) => r.trim().toUpperCase());
  Logger.log(`User ${email} has roles: ${userRoles.join(", ")}`);

  // Check each role
  for (const role of userRoles) {
    let def = ROLE_DEFINITIONS[role];

    // --- DYNAMIC ROLE PARSING ---
    // Support syntax like "TOP:SCH" or "FRU:SGE+SCH" or "*:GNB"
    if (!def && role.includes(":")) {
      const parts = role.split(":");

      // Validate format
      if (parts.length !== 2 || !parts[0].trim() || !parts[1].trim()) {
        Logger.log(`Invalid role syntax: ${role} (use FIGHT:JOB format)`);
        continue;
      }

      const [fPart, jPart] = parts;

      // Parse Fights (supports * and + separator)
      const fights =
        fPart.trim() === "*"
          ? "*"
          : fPart
              .split("+")
              .map((f) => f.trim().toUpperCase())
              .filter((f) => f);

      // Parse Jobs (supports * and + separator)
      const jobs =
        jPart.trim() === "*"
          ? "*"
          : jPart
              .split("+")
              .map((j) => j.trim().toUpperCase())
              .filter((j) => j);

      // Validate parsed values
      if (
        (Array.isArray(fights) && fights.length === 0) ||
        (Array.isArray(jobs) && jobs.length === 0)
      ) {
        Logger.log(`Invalid role: ${role} (empty fight or job list)`);
        continue;
      }

      def = { fights, jobs };
      Logger.log(
        `Parsed dynamic role: ${role} → fights:${JSON.stringify(fights)}, jobs:${JSON.stringify(jobs)}`,
      );
    }

    if (!def) {
      Logger.log(`Unknown role: ${role}`);
      continue;
    }

    // Check permissions
    const fightMatch =
      def.fights === "*" ||
      (Array.isArray(def.fights) &&
        def.fights.some((f) => f.toUpperCase() === fight.toUpperCase()));

    const jobMatch =
      def.jobs === "*" ||
      (Array.isArray(def.jobs) &&
        def.jobs.some((j) => j.toUpperCase() === job.toUpperCase()));

    if (fightMatch && jobMatch) {
      Logger.log(`✓ Permission granted via role: ${role}`);
      return true;
    }
  }

  Logger.log(`No matching permissions found`);
  return false;
}

/**
 * Update the locked public sheet
 */
function updatePublicSheet(fight, job, link) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const targetSheet = ss.getSheetByName(CONFIG.SHEET_PUBLIC);

  if (!targetSheet) {
    throw new Error(
      `Sheet '${CONFIG.SHEET_PUBLIC}' not found. Check sheet name.`,
    );
  }

  const lastRow = targetSheet.getLastRow();
  const lastCol = targetSheet.getLastColumn();

  if (lastRow < 2 || lastCol < 2) {
    throw new Error(
      `Sheet '${CONFIG.SHEET_PUBLIC}' is empty or improperly formatted.`,
    );
  }

  // Read entire grid once
  const data = targetSheet.getRange(1, 1, lastRow, lastCol).getValues();
  const headers = data[0];
  const jobList = data.map((row) => row[0]);

  // Find coordinates (case-insensitive)
  const fightIndex = headers.findIndex(
    (h) => h && h.toString().trim().toUpperCase() === fight.toUpperCase(),
  );
  const jobIndex = jobList.findIndex(
    (j) => j && j.toString().trim().toUpperCase() === job.toUpperCase(),
  );

  if (fightIndex === -1) {
    const available = headers.filter((h) => h).join(", ");
    throw new Error(`Fight '${fight}' not found. Available: ${available}`);
  }

  if (jobIndex === -1) {
    const available = jobList.filter((j) => j).join(", ");
    throw new Error(`Job '${job}' not found. Available: ${available}`);
  }

  // Update cell
  targetSheet.getRange(jobIndex + 1, fightIndex + 1).setValue(link);
  Logger.log(`✓ Updated ${fight}/${job} → ${link}`);
}
