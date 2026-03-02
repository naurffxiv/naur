/**
 * Shared constants and utility functions for QA Actions
 */

const ARCHIVE_PREFIX = '> ⚠️ **"Needs QA" was unchecked**';

// Markers that prove the template body was never touched at all.
// Used for the delete-vs-archive decision: only delete when the dev did zero work.
const TEMPLATE_UNEDITED_MARKERS = [
  "[Example:",
  "[Delete this line if not needed]",
  "[Delete examples above and add your own]",
];

// All placeholder strings that must be replaced before the checklist is valid.
// Superset of TEMPLATE_UNEDITED_MARKERS used by validation.
const TEMPLATE_PLACEHOLDERS = [
  ...TEMPLATE_UNEDITED_MARKERS,
  "[Name]",
  "[One sentence:",
  "[What should happen?]",
];

module.exports = {
  BOT_MARKER: "<!-- QA_CHECKLIST_BOT_MARKER -->",
  STATUS_MARKER: "<!-- QA_VALIDATION_STATUS -->",
  ARCHIVE_PREFIX,
  QA_TEAM_SLUG: process.env.QA_TEAM_SLUG || "qa",

  /**
   * Check if the PR body has the "Needs QA" checkbox checked.
   * NOTE: A functionally equivalent pattern is duplicated in
   * .github/workflows/sync-pr-ticket-status.yml; keep them in sync.
   * (Shell steps cannot import JS modules, so the regex must be inlined there.)
   */
  needsQA: (body) => /(?:[-*+]\s*)?\[x\]\s*Needs QA/i.test(body || ""),

  /**
   * Find a comment by its marker
   */
  findComment: (comments, marker) =>
    comments.find((c) => c.body?.includes(marker)),

  /**
   * Check if a checklist comment was archived when "Needs QA" was unchecked.
   */
  isArchived: (commentBody) => (commentBody || "").includes(ARCHIVE_PREFIX),

  /**
   * Check if a checklist comment is completely untouched (safe to delete outright).
   * Conservative: only checks markers that prove the dev customised nothing.
   * Used for the delete-vs-archive decision when "Needs QA" is unchecked.
   */
  isTemplateUnedited: (commentBody) =>
    TEMPLATE_UNEDITED_MARKERS.some((p) => (commentBody || "").includes(p)),

  /**
   * Check if a checklist comment still has any unfilled placeholder text.
   * Strict: catches metadata fields like [Name] as well as example steps.
   * Used by validation to ensure the template is fully completed.
   */
  hasUnfilledPlaceholders: (commentBody) =>
    TEMPLATE_PLACEHOLDERS.some((p) => (commentBody || "").includes(p)),

  /**
   * Paginate all issue comments for a PR.
   */
  getComments: (github, context, prNumber) =>
    github.paginate(github.rest.issues.listComments, {
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: prNumber,
    }),

  /**
   * Extract owner/repo from context for use as spread params in calls.
   */
  repoParams: (context) => ({
    owner: context.repo.owner,
    repo: context.repo.repo,
  }),
};
