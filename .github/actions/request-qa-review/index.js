/**
 * Request QA Review
 * Called when a code reviewer approves a PR requests QA team review if "Needs QA" is checked.
 * Safe to call multiple times (e.g. after QA finds issues, dev fixes, code reviewer re-approves).
 */

const shared = require("../shared.js");

module.exports = async ({ github, context, core }) => {
  try {
    const pr = context.payload.pull_request;
    if (!pr) {
      core.setFailed(
        "No pull_request found in context. This action should run on pull_request_review events.",
      );
      return;
    }
    const body = pr.body || "";

    if (!shared.needsQA(body)) {
      core.info("PR does not need QA. Skipping.");
      return;
    }

    // Skip if QA team is already in the pending reviewers list
    const { data: reviewRequests } =
      await github.rest.pulls.listRequestedReviewers({
        ...shared.repoParams(context),
        pull_number: pr.number,
      });

    const alreadyRequested = reviewRequests.teams?.some(
      (t) => t.slug === shared.QA_TEAM_SLUG,
    );

    if (alreadyRequested) {
      core.info(
        `QA team (@${context.repo.owner}/${shared.QA_TEAM_SLUG}) is already a requested reviewer. Skipping.`,
      );
      return;
    }

    // Fetch members of the QA team to cross-reference with reviewers
    const teamMembers = await github.paginate(
      github.rest.teams.listMembersInOrg,
      {
        org: context.repo.owner,
        team_slug: shared.QA_TEAM_SLUG,
      },
    );

    const qaLogins = new Set(teamMembers.map((m) => m.login));

    // Process reviews to see if any QA member has already approved
    const reviews = await github.paginate(github.rest.pulls.listReviews, {
      ...shared.repoParams(context),
      pull_number: pr.number,
    });

    // Track the latest non-comment state per QA member.
    // Explicitly sort by date to ensure the Map captures the actual final state correctly
    // even if the API returns pages out of order.
    const latestQAStates = new Map();
    const sortedReviews = [...reviews].sort(
      (a, b) => new Date(a.submitted_at) - new Date(b.submitted_at),
    );

    for (const review of sortedReviews) {
      if (!review.user || review.state === "COMMENTED") continue;
      if (!qaLogins.has(review.user.login)) continue;
      // Note: DISMISSED state correctly overwrites previous approvals
      latestQAStates.set(review.user.login, review.state);
    }

    const qaHasApproval = [...latestQAStates.values()].some(
      (state) => state === "APPROVED",
    );

    if (qaHasApproval) {
      core.info(
        "QA team member has already approved this PR. Skipping review request.",
      );
      return;
    }

    // Request the QA Team review
    await github.rest.pulls.requestReviewers({
      ...shared.repoParams(context),
      pull_number: pr.number,
      team_reviewers: [shared.QA_TEAM_SLUG],
    });

    core.info(
      `Requested review from @${context.repo.owner}/${shared.QA_TEAM_SLUG}.`,
    );
  } catch (error) {
    core.setFailed(`Request QA Review Failure: ${error.message}`);
  }
};
