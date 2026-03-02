/**
 * Check QA Approval Gate
 * Blocks merge on QA PRs until 2 unique reviewers have approved.
 * Non-QA PRs (Needs QA unchecked) always pass immediately.
 */

module.exports = async ({ github, context, core }) => {
  try {
    const shared = require("../shared.js");
    const pr = context.payload.pull_request;
    const body = pr.body || "";

    if (!shared.needsQA(body)) {
      console.log("PR does not need QA. Gate passes.");
      return;
    }

    const reviews = await github.paginate(github.rest.pulls.listReviews, {
      ...shared.repoParams(context),
      pull_number: pr.number,
    });

    // Group by reviewer, keeping only the latest non-comment state.
    // COMMENTED reviews are skipped; they don't change approval status.
    const latestByReviewer = new Map();
    for (const review of reviews) {
      if (review.state === "COMMENTED" || !review.user) continue;
      latestByReviewer.set(review.user.login, review.state);
    }

    const approvalCount = [...latestByReviewer.values()].filter(
      (state) => state === "APPROVED",
    ).length;

    console.log(`QA gate: ${approvalCount}/2 approvals`);

    if (approvalCount >= 2) {
      console.log("✅ QA gate passed.");
    } else {
      core.setFailed(
        `QA gate: ${approvalCount}/2 approvals. Waiting for code reviewer + QA team approval before merge.`,
      );
    }
  } catch (error) {
    console.error("QA Approval Gate Error:", error);
    core.setFailed(`QA Approval Gate Failure: ${error.message}`);
  }
};
