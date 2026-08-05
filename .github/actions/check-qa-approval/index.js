/**
 * Check QA Approval Gate
 * Blocks merge on QA PRs until 2 unique reviewers have approved,
 * with at least one approval from the QA team.
 * Non-QA PRs (Needs QA unchecked) always pass immediately.
 */

const shared = require("../shared.js");

const REVIEW_PROPAGATION_DELAY_MS = 5000;
const REVIEW_PROPAGATION_RETRIES = 3;

module.exports = async ({ github, context, core }) => {
  try {
    const pr = context.payload.pull_request;
    const body = pr.body || "";
    const sha = pr.head.sha;

    // Fail-safe: pending immediately so crashes don't look like gate blocks.
    await github.rest.repos.createCommitStatus({
      ...shared.repoParams(context),
      sha,
      state: "pending",
      context: "QA Approval Gate",
      description: "Waiting for code review...",
    });

    if (!shared.needsQA(body)) {
      console.log("PR does not need QA. Gate passes.");
      await github.rest.repos.createCommitStatus({
        ...shared.repoParams(context),
        sha,
        state: "success",
        context: "QA Approval Gate",
        description: "QA not required for this PR.",
      });
      return;
    }

    const qaMemberLogins = await shared.getTeamMemberLogins(
      github,
      context,
      shared.QA_TEAM_SLUG,
      true,
    );

    const cmMemberLogins = await shared.getTeamMemberLogins(
      github,
      context,
      shared.CM_TEAM_SLUG,
      false,
    );

    // On pull_request_review events, GitHub's API may not yet reflect the
    // new review that triggered this run. Retry with backoff to handle this
    // propagation delay.
    const isReviewEvent = context.eventName === "pull_request_review";
    const maxAttempts = isReviewEvent ? REVIEW_PROPAGATION_RETRIES : 1;

    let hasQAApproval = false;
    let hasDevApproval = false;

    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      if (attempt > 1) {
        console.log(
          `QA approval not yet visible, retrying in ${REVIEW_PROPAGATION_DELAY_MS / 1000}s... (attempt ${attempt}/${maxAttempts})`,
        );
        await new Promise((resolve) =>
          setTimeout(resolve, REVIEW_PROPAGATION_DELAY_MS),
        );
      }

      const reviews = await github.paginate(github.rest.pulls.listReviews, {
        ...shared.repoParams(context),
        pull_number: pr.number,
      });

      // Last non-comment state per reviewer is the one that counts.
      const latestByReviewer = new Map();
      for (const review of reviews) {
        if (review.state === "COMMENTED" || !review.user) continue;
        if (review.user.login === pr.user.login) continue;
        latestByReviewer.set(review.user.login, review.state);
      }

      const approvers = [...latestByReviewer.entries()]
        .filter(([, state]) => state === "APPROVED")
        .map(([login]) => login);

      hasQAApproval = approvers.some((login) => qaMemberLogins.has(login));
      // "Dev approval" means any approver outside the QA and CM teams.
      // Two QA members approving does NOT satisfy the gate
      hasDevApproval = approvers.some(
        (login) => !qaMemberLogins.has(login) && !cmMemberLogins.has(login),
      );

      console.log(
        `QA gate (attempt ${attempt}): QA approval=${hasQAApproval}, dev approval=${hasDevApproval}`,
      );

      if (hasQAApproval && hasDevApproval) break;

      // If the triggering review's approval type is already reflected
      // in the API, no point retrying.
      if (isReviewEvent) {
        const triggeringLogin = context.payload.review?.user?.login;
        const triggeringState = context.payload.review?.state?.toUpperCase();
        if (
          triggeringLogin &&
          triggeringState === "APPROVED" &&
          latestByReviewer.get(triggeringLogin) === "APPROVED"
        ) {
          console.log(
            "Triggering review is now visible in API. Stopping retry.",
          );
          break;
        }
      }
    }

    const gatePass = hasQAApproval && hasDevApproval;

    const stateLabels = [
      "needs-code-review",
      "needs-qa-approval",
      "ready-to-merge",
    ];
    let targetLabel = "needs-code-review";
    if (gatePass) {
      targetLabel = "ready-to-merge";
    } else if (hasDevApproval) {
      targetLabel = "needs-qa-approval";
    }

    try {
      const { data: existingLabels } =
        await github.rest.issues.listLabelsOnIssue({
          ...shared.repoParams(context),
          issue_number: pr.number,
        });

      const existingLabelNames = existingLabels.map((l) => l.name);

      for (const label of stateLabels) {
        if (label !== targetLabel && existingLabelNames.includes(label)) {
          await github.rest.issues.removeLabel({
            ...shared.repoParams(context),
            issue_number: pr.number,
            name: label,
          });
        }
      }

      if (!existingLabelNames.includes(targetLabel)) {
        await github.rest.issues.addLabels({
          ...shared.repoParams(context),
          issue_number: pr.number,
          labels: [targetLabel],
        });
      }
    } catch (error) {
      console.warn("Failed to manage PR labels:", error.message);
    }

    if (gatePass) {
      console.log("✅ QA gate passed.");
      await github.rest.repos.createCommitStatus({
        ...shared.repoParams(context),
        sha,
        state: "success",
        context: "QA Approval Gate",
        description: "Code reviewer + QA team approved.",
      });
    } else {
      let pendingDescription;
      if (!hasDevApproval && !hasQAApproval) {
        pendingDescription =
          "Waiting for a code reviewer approval and a QA team approval.";
      } else if (!hasDevApproval) {
        pendingDescription =
          "QA approved. Waiting for a code reviewer approval.";
      } else {
        pendingDescription =
          "Code reviewer approved. Waiting for a QA team approval.";
      }
      console.log(`QA gate pending: ${pendingDescription}`);
      await github.rest.repos.createCommitStatus({
        ...shared.repoParams(context),
        sha,
        state: "pending",
        context: "QA Approval Gate",
        description: pendingDescription.slice(0, 140),
      });
    }
  } catch (error) {
    console.error("QA Approval Gate Error:", error);
    try {
      const sha = context.payload.pull_request?.head?.sha;
      if (sha) {
        await github.rest.repos.createCommitStatus({
          ...shared.repoParams(context),
          sha,
          state: "failure",
          context: "QA Approval Gate",
          description:
            `Script error: ${error.message ?? "Unknown error"}`.slice(0, 140),
        });
      }
    } catch (statusError) {
      console.log(`Failed to post commit status: ${statusError.message}`);
    }
    core.setFailed(`QA Approval Gate Failure: ${error.message}`);
  }
};
