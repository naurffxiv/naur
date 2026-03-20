/**
 * Post QA Checklist
 * Creates or updates a sticky checklist comment
 */

const FRESH_TEMPLATE_RE = /(?:[-*+]\s*)?\[x\]\s*Request a fresh template/i;

module.exports = async ({ github, context, core }) => {
  const shared = require("../shared.js");
  const fetchTemplate = require("./fetch-template.js");
  const pr = context.payload.pull_request;
  const body = pr.body || "";
  const wikiUrl = process.env.QA_WIKI_URL;

  const comments = await shared.getComments(github, context, pr.number);
  const existingComment = shared.findComment(comments, shared.BOT_MARKER);
  const needsQA = shared.needsQA(body);

  if (!needsQA) {
    if (existingComment) {
      // "Needs QA" unchecked but checklist exists → archive to preserve position
      if (!shared.isArchived(existingComment.body)) {
        console.log(
          "Needs QA unchecked archiving checklist to preserve position in timeline.",
        );
        const archivedBody = `> [!NOTE]\n${shared.ARCHIVE_PREFIX} this checklist has been archived.\n> Re-check the box to re-activate QA validation.\n\n- [ ] Request a fresh template (check this before re-checking **Needs QA** to get a clean slate)\n\n---\n\n${existingComment.body}`;
        await github.rest.issues.updateComment({
          ...shared.repoParams(context),
          comment_id: existingComment.id,
          body: archivedBody,
        });
      } else {
        console.log(
          "Needs QA unchecked checklist already archived, skipping.",
        );
      }
    } else {
      console.log("PR does not need QA. Skipping.");
    }
    return;
  }

  try {
    const templateContent = await fetchTemplate(github, wikiUrl);

    const commentBody = `# QA Test Checklist

**Please fill out this checklist while in Draft mode before marking the PR as Ready for Review.**

See the [QA Workflow Documentation](${wikiUrl}) for full details.

---

${templateContent}

${shared.BOT_MARKER}`;

    if (existingComment) {
      if (shared.isArchived(existingComment.body)) {
        // "Needs QA" was re-checked after being unchecked.
        // If the dev edited the checklist while archived, restore their content.
        // If it was never touched, replace with a fresh template.
        const ARCHIVE_SEPARATOR = "\n\n---\n\n";
        const separatorIndex = existingComment.body.indexOf(ARCHIVE_SEPARATOR);
        const innerContent =
          separatorIndex >= 0
            ? existingComment.body.slice(
                separatorIndex + ARCHIVE_SEPARATOR.length,
              )
            : existingComment.body;

        const requestedFresh = FRESH_TEMPLATE_RE.test(existingComment.body);
        const wasEdited =
          !requestedFresh && !shared.isTemplateUnedited(innerContent);
        const restoredBody = wasEdited ? innerContent : commentBody;
        const logMessage = requestedFresh
          ? "Found archived checklist (fresh template requested). Replacing with fresh template."
          : wasEdited
            ? "Found archived checklist (edited). Restoring dev's content."
            : "Found archived checklist (unedited). Replacing with fresh template.";
        const successMessage = wasEdited
          ? "✅ QA Checklist restored (dev's edits preserved)."
          : "✅ Fresh QA Checklist posted (updated in place).";

        console.log(logMessage);
        try {
          await github.rest.issues.updateComment({
            ...shared.repoParams(context),
            comment_id: existingComment.id,
            body: restoredBody,
          });
          console.log(successMessage);
        } catch (updateErr) {
          if (updateErr.status === 404) {
            console.warn(
              "Archived comment no longer exists, creating a new one.",
            );
            await github.rest.issues.createComment({
              ...shared.repoParams(context),
              issue_number: pr.number,
              body: restoredBody,
            });
            console.log("✅ QA Checklist posted.");
          } else {
            throw updateErr;
          }
        }
      } else {
        console.log(
          "QA Checklist already exists. Skipping update to preserve user edits.",
        );
      }
    } else {
      await github.rest.issues.createComment({
        ...shared.repoParams(context),
        issue_number: pr.number,
        body: commentBody,
      });
      console.log("✅ QA Checklist posted.");
    }
  } catch (err) {
    console.error(`❌ Error: ${err.message}`);
    core.setFailed(`QA Checklist Posting Failure: ${err.message}`);
  }
};
