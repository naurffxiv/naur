/**
 * Manages PR draft state and QA review requests.
 */
module.exports = {
  /**
   * Removes the QA team review request from the PR.
   * Silently ignores 422 errors (team wasn't a reviewer yet).
   */
  removeQAReviewRequest: async (github, context, pr, qaTeamSlug, shared) => {
    try {
      await github.rest.pulls.removeRequestedReviewers({
        ...shared.repoParams(context),
        pull_number: pr.number,
        reviewers: [],
        team_reviewers: [qaTeamSlug],
      });
      console.log(`Removed QA team review request.`);
    } catch (err) {
      // 422 = team wasn't a reviewer yet, which is fine
      if (err.status !== 422)
        console.warn("Could not remove QA review request:", err.message);
    }
  },

  /**
   * Converts the PR to Draft via GraphQL mutation.
   * Logs but does not throw on failure (e.g. permission errors).
   */
  convertToDraft: async (github, pr) => {
    if (pr.draft) {
      console.log("PR is already a Draft. Skipping mutation.");
      return;
    }
    console.log("Converting PR back to Draft due to validation failure...");
    try {
      await github.graphql(
        `mutation($id: ID!) {
          convertPullRequestToDraft(input: { pullRequestId: $id }) {
            pullRequest { isDraft }
          }
        }`,
        { id: pr.node_id },
      );
      console.log("Successfully converted PR to Draft.");
    } catch (mutationError) {
      console.error("Failed to convert PR to Draft:", mutationError.message);
    }
  },
};
