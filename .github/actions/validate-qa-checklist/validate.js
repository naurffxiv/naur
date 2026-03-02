/**
 * Validate QA Checklist
 * Updates a sticky validation status comment
 */

module.exports = async ({ github, context, core }) => {
  let existingStatusComment = null;

  try {
    const shared = require("../shared.js");
    const checkItems = require("./check-items.js");
    const draft = require("./manage-draft.js");
    const pr = context.payload.pull_request;
    const body = pr.body || "";

    const comments = await shared.getComments(github, context, pr.number);
    existingStatusComment = shared.findComment(comments, shared.STATUS_MARKER);

    const needsQA = shared.needsQA(body);

    // Clean up status comment if "Needs QA" unchecked
    if (!needsQA) {
      console.log("PR does not need QA. Skipping validation.");
      if (existingStatusComment) {
        await github.rest.issues.deleteComment({
          ...shared.repoParams(context),
          comment_id: existingStatusComment.id,
        });
        console.log("Cleaned up orphaned status comment");
      }
      await draft.removeQAReviewRequest(
        github,
        context,
        pr,
        shared.QA_TEAM_SLUG,
        shared,
      );
      return;
    }

    const checklistComment = shared.findComment(comments, shared.BOT_MARKER);

    const createOrUpdateStatus = async (statusMessage) => {
      const fullBody = statusMessage + "\n" + shared.STATUS_MARKER;
      if (existingStatusComment) {
        const { data } = await github.rest.issues.updateComment({
          ...shared.repoParams(context),
          comment_id: existingStatusComment.id,
          body: fullBody,
        });
        return data.node_id;
      } else {
        const { data } = await github.rest.issues.createComment({
          ...shared.repoParams(context),
          issue_number: pr.number,
          body: fullBody,
        });
        return data.node_id;
      }
    };

    let errors = [];
    let errorDetails = [];

    if (!checklistComment) {
      errors.push(
        "❌ **QA Checklist Missing** This PR is being converted to **Draft** so the checklist can be generated. (If it still doesn't appear, please check the 'Post QA Checklist' action logs for configuration errors like a missing Wiki URL).",
      );
      errorDetails.push(null);
    } else {
      ({ errors, errorDetails } = checkItems(checklistComment.body, shared));
    }

    if (errors.length > 0) {
      // Check for existing approvals to prevent demotion loops.
      // Conservative: any historical APPROVED review blocks draft conversion,
      // even if a later "changes requested" review was submitted by the same
      // reviewer. This is intentional; we prefer to never demote an approved
      // PR. To switch to latest-per-reviewer semantics, reduce `reviews` to
      // the last state per reviewer before checking for APPROVED.
      let hasApprovals = false;
      try {
        const reviews = await github.paginate(github.rest.pulls.listReviews, {
          ...shared.repoParams(context),
          pull_number: pr.number,
        });
        hasApprovals = reviews.some((review) => review.state === "APPROVED");
      } catch (reviewError) {
        console.warn("Could not fetch PR reviews:", reviewError.message);
      }

      const errorSection = errors
        .map((err, i) => {
          const detail = errorDetails[i];
          if (!detail) return err;
          return `${err}\n\n<details>\n<summary>Show details</summary>\n\n${detail}\n\n</details>`;
        })
        .join("\n\n");

      const draftWarning = hasApprovals
        ? "Please fix the items above."
        : "This PR has been converted back to **Draft** because the QA checklist is incomplete or generic. Please fix the items above before marking it as Ready for Review again.";

      const statusBody = `### 🚫 QA Validation Failed\n\n${errorSection}\n\n**Action Required:** ${draftWarning}\n\ncc: @${pr.user.login}`;

      // Unminimize before posting failure so it becomes visible if previously resolved
      if (existingStatusComment?.node_id) {
        try {
          await github.graphql(
            `mutation($id: ID!) {
              unminimizeComment(input: { subjectId: $id }) {
                unminimizedComment { isMinimized }
              }
            }`,
            { id: existingStatusComment.node_id },
          );
        } catch (unminimizeError) {
          console.warn(
            `Failed to unminimizeComment (${existingStatusComment.node_id}):`,
            unminimizeError.message,
          );
        }
      }

      await createOrUpdateStatus(statusBody);
      await draft.removeQAReviewRequest(
        github,
        context,
        pr,
        shared.QA_TEAM_SLUG,
        shared,
      );

      if (hasApprovals) {
        console.log(
          "PR already has approvals. Skipping conversion to Draft to prevent demotion loop.",
        );
        core.setFailed(
          "QA checklist validation failed - fix checklist items. (Skipped draft conversion due to existing approvals)",
        );
      } else {
        await draft.convertToDraft(github, pr);
        core.setFailed("QA checklist validation failed - PR returned to Draft");
      }
    } else {
      const statusBody = `### ✅ QA Validation Passed\n\nAll checklist items completed and customized. Ready for review!`;

      const nodeId = await createOrUpdateStatus(statusBody);
      console.log("✅ Validation passed");

      // Minimize the comment as resolved so it collapses in the PR timeline
      try {
        await github.graphql(
          `mutation($id: ID!) {
            minimizeComment(input: { subjectId: $id, classifier: RESOLVED }) {
              minimizedComment { isMinimized }
            }
          }`,
          { id: nodeId },
        );
        console.log("Status comment marked as resolved.");
      } catch (minimizeError) {
        console.warn("Could not minimize comment:", minimizeError.message);
      }
    }
  } catch (error) {
    console.error("QA Checklist Validation Error:", error);
    core.setFailed(`QA Checklist Validation Hub Failure: ${error.message}`);
  }
};
