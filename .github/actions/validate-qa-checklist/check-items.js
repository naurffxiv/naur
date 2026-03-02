/**
 * Validates the content of the QA checklist comment.
 * Returns { errors, errorDetails } arrays describing any issues found.
 */
module.exports = function checkItems(checklistBody, shared) {
  const errors = [];
  const errorDetails = [];

  if (shared.hasUnfilledPlaceholders(checklistBody)) {
    errors.push(
      "❌ **Generic Checklist** Replace template examples with specific steps",
    );
    errorDetails.push(
      "Remove all placeholder text such as `[Name]`, `[One sentence: ...]`, `[Example: ...]`, and `[What should happen?]` and replace with steps specific to your feature.",
    );
  }

  const uncheckedItems = (
    checklistBody.match(/^\s*[-*+]\s*\[ \].*/gm) || []
  ).map((line) => line.trim());
  if (uncheckedItems.length > 0) {
    errors.push(
      `❌ **Incomplete Checklist** ${uncheckedItems.length} unchecked item(s)`,
    );
    errorDetails.push(
      `**Items still unchecked:**\n\n${uncheckedItems.join("\n")}`,
    );
  }

  return { errors, errorDetails };
};
