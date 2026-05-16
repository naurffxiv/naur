/**
 * Check QA Approval Gate
 * Blocks merge on QA PRs until 2 unique reviewers have approved,
 * with at least one approval from the QA team.
 * Non-QA PRs (Needs QA unchecked) always pass immediately.
 */

module.exports = async ({ github, context, core }) => {
  try {
    const shared = require("../shared.js");
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

    const qaTeamSlug = shared.QA_TEAM_SLUG;
    let qaMembers;
    try {
      qaMembers = await github.paginate(github.rest.teams.listMembersInOrg, {
        org: context.repo.owner,
        team_slug: qaTeamSlug,
      });
    } catch (error) {
      throw new Error(
        `Failed to fetch QA team members for '${qaTeamSlug}'. ${error.message}`,
      );
    }
    const qaMemberLogins = new Set(qaMembers.map((m) => m.login));

    const hasQAApproval = approvers.some((login) => qaMemberLogins.has(login));
    // "Dev approval" means any approver outside the QA team.
    // Two QA members approving does NOT satisfy the gate
    const hasDevApproval = approvers.some(
      (login) => !qaMemberLogins.has(login),
    );

    console.log(
      `QA gate: QA approval=${hasQAApproval}, dev approval=${hasDevApproval}`,
    );

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
      try {
        await github.rest.repos.createCommitStatus({
          ...shared.repoParams(context),
          sha,
          state: "success",
          context: "QA Approval Gate",
          description: "Code reviewer + QA team approved.",
        });
      } catch (error) {
        console.warn(
          `Failed to set QA Approval Gate status to success for sha ${sha}:`,
          error.message,
        );
      }
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
      try {
        await github.rest.repos.createCommitStatus({
          ...shared.repoParams(context),
          sha,
          state: "pending",
          context: "QA Approval Gate",
          description: pendingDescription.slice(0, 140),
        });
      } catch (error) {
        console.warn(
          `Failed to set QA Approval Gate status to pending for sha ${sha}:`,
          error.message,
        );
      }
    }
  } catch (error) {
    console.error("QA Approval Gate Error:", error);
    try {
      const sha = context.payload.pull_request?.head?.sha;
      if (sha) {
        const shared = require("../shared.js");
        await github.rest.repos.createCommitStatus({
          ...shared.repoParams(context),
          sha,
          state: "failure",
          context: "QA Approval Gate",
          description:
            `Script error: ${error.message ?? "Unknown error"}`.slice(0, 140),
        });
      }
    } catch (_) {}
    core.setFailed(`QA Approval Gate Failure: ${error.message}`);
  }
};
