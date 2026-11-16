// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Data;

/// <summary>
/// Admits one database migration at a time within this process, without waiting.
/// </summary>
/// <remarks>
/// Registered as a singleton. EF Core's own migration lock still serializes migrations across processes; this gate
/// lets a second caller in the same process learn immediately that a migration is already running.
/// </remarks>
public sealed class MigrationGate : IDisposable
{
    private readonly SemaphoreSlim _semaphore = new(1, 1);

    /// <summary>
    /// Attempts to enter the gate without waiting.
    /// </summary>
    /// <returns><see langword="true"/> when the gate was entered; <see langword="false"/> when a migration is already running.</returns>
    public bool TryEnter()
    {
        return _semaphore.Wait(0);
    }

    /// <summary>
    /// Leaves the gate after a successful <see cref="TryEnter"/>.
    /// </summary>
    public void Exit()
    {
        _semaphore.Release();
    }

    /// <inheritdoc/>
    public void Dispose()
    {
        _semaphore.Dispose();
    }
}
