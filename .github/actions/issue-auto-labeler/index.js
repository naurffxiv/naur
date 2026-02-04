/**
 * Issue Auto-Labeler
 * Syncs GitHub Issue Form selections to labels.
 * Only manages labels defined in LABEL_MAPS; respects manual labels.
 */

module.exports = async ({ github, context }) => {
  const body = context.payload.issue.body || "";
  const labels = [];

  // Parses the LABEL_MAPS input structured as:
  // { "Form Field Name": { "Option A": "Label A", "Option B": "Label B" } }
  const labelMaps = JSON.parse(process.env.LABEL_MAPS);

  // GitHub Issue Forms render as "### <Label>\n\n<Value>\n" in the body.
  // This pulls out the value for a given form field.
  function getFormValue(label) {
    const safeLabel = label.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const regex = new RegExp(
      `### ${safeLabel}\\s*\\n\\n([\\s\\S]*?)(?:\\n{2,}|\\n###|$)`,
      "i",
    );
    const match = body.match(regex);
    const value = match ? match[1].trim() : null;

    // These are GitHub's default "nothing selected" states — treat them as empty
    return value === "_No response_" || value === "Other" || value === "None"
      ? null
      : value;
  }

  function getLabelsFromMap(map, rawValue) {
    if (!rawValue) return [];

    // GitHub sends multi-selects as "Value A, Value B"
    const values = rawValue
      .split(",")
      .map((v) => v.trim())
      .filter(Boolean);

    const foundLabels = [];

    for (const val of values) {
      const foundKey = Object.keys(map).find(
        (k) => k.toLowerCase() === val.toLowerCase(),
      );
      if (foundKey) {
        foundLabels.push(map[foundKey]);
      }
    }

    return foundLabels;
  }

  const allManagedLabels = new Set();

  for (const [fieldName, map] of Object.entries(labelMaps)) {
    Object.values(map).forEach((labelName) =>
      allManagedLabels.add(labelName.toLowerCase()),
    );

    const value = getFormValue(fieldName);
    const newLabels = getLabelsFromMap(map, value);
    labels.push(...newLabels);
  }

  const currentLabels = context.payload.issue.labels.map(
    (labelObj) => labelObj.name,
  );

  const toAdd = labels.filter(
    (labelName) =>
      !currentLabels.some((c) => c.toLowerCase() === labelName.toLowerCase()),
  );

  // Only remove labels we own that are no longer selected in the form
  const toRemove = currentLabels.filter((labelName) => {
    const lowerLabelName = labelName.toLowerCase();
    const isManaged = allManagedLabels.has(lowerLabelName);
    const isDesired = labels.some((d) => d.toLowerCase() === lowerLabelName);
    return isManaged && !isDesired;
  });

  if (toAdd.length > 0) {
    await github.rest.issues.addLabels({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: context.payload.issue.number,
      labels: toAdd,
    });
    console.log("Added labels:", toAdd);
  }

  for (const label of toRemove) {
    try {
      await github.rest.issues.removeLabel({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: context.payload.issue.number,
        name: label,
      });
      console.log("Removed stale label:", label);
    } catch (e) {
      // Race condition
      console.log(`Label "${label}" already gone, skipping.`);
    }
  }

  if (toAdd.length === 0 && toRemove.length === 0) {
    console.log("No label changes required.");
  }
};
