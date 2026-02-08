// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Tests.Testing;
using Quartz;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Contains tests that verify quartz components are correctly registered in the dependency injection container.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve registered services for testing.</param>
[AuthingwayDataSource]
public class QuartzTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that the SchedulerFactory service is registered in the dependency injection container.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task SchedulerFactoryRegistered()
    {
        var stdSchedulerFactory = serviceProvider.GetService<ISchedulerFactory>();

        await Assert.That(stdSchedulerFactory)
            .IsNotNull();
    }
}
