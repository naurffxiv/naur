// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using NetCord;

namespace Naur.Authingway.Discord.Extensions;

/// <summary>
/// Reads submitted values from modal components.
/// </summary>
public static class ModalComponentExtensions
{
    /// <summary>
    /// Gets the value of a submitted text input.
    /// </summary>
    /// <param name="components">The submitted components.</param>
    /// <param name="customId">The text input custom identifier.</param>
    /// <returns>The value, or an empty string when absent.</returns>
    public static string GetText(this IEnumerable<IModalComponent> components, string customId)
    {
        return components.Find<TextInput>(input => input.CustomId == customId)?.Value ?? string.Empty;
    }

    /// <summary>
    /// Gets the checked values of a submitted checkbox group.
    /// </summary>
    /// <param name="components">The submitted components.</param>
    /// <param name="customId">The checkbox group custom identifier.</param>
    /// <returns>The checked values, or none when absent.</returns>
    public static IReadOnlyList<string> GetCheckedValues(this IEnumerable<IModalComponent> components, string customId)
    {
        return components.Find<CheckboxGroup>(group => group.CustomId == customId)?.CheckedValues ?? [];
    }

    /// <summary>
    /// Gets whether a submitted checkbox is checked.
    /// </summary>
    /// <param name="components">The submitted components.</param>
    /// <param name="customId">The checkbox custom identifier.</param>
    /// <returns>True when the checkbox is present and checked.</returns>
    public static bool IsChecked(this IEnumerable<IModalComponent> components, string customId)
    {
        return components.Find<Checkbox>(checkbox => checkbox.CustomId == customId)?.Checked ?? false;
    }

    private static T? Find<T>(this IEnumerable<IModalComponent> components, Func<T, bool> predicate)
        where T : class
    {
        return components.OfType<Label>()
            .Select(label => label.Component)
            .OfType<T>()
            .FirstOrDefault(predicate);
    }
}
