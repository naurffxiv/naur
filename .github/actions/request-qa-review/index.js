/**
 * Request QA Review
 * Called when a code reviewer approves a PR requests QA team review if "Needs QA" is checked.
 * Safe to call multiple times (e.g. after QA finds issues, dev fixes, code reviewer re-approves).
 */

module.exports = async ({ github, context, core }) => {
  try {
    const shared = require("../shared.js");
    const pr = context.payload.pull_request;
    const body = pr.body || "";

    if (!shared.needsQA(body)) {
      console.log("PR does not need QA. Skipping.");
      return;
    }

    try {
      await github.rest.pulls.requestReviewers({
        ...shared.repoParams(context),
        pull_number: pr.number,
        team_reviewers: [shared.QA_TEAM_SLUG],
      });
      console.log(
        `Requested review from @${context.repo.owner}/${shared.QA_TEAM_SLUG}.`,
      );
    } catch (err) {
      console.warn("Could not request QA team review:", err.message);
    }
  } catch (error) {
    console.error("Request QA Review Error:", error);
    core.setFailed(`Request QA Review Failure: ${error.message}`);
  }
};
