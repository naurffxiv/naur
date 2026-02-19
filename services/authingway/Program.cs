// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Core;
using Naur.Authingway.Data;

namespace Naur.Authingway;

/// <summary>
/// Provides the entry point for the application.
/// </summary>
public class Program
{
    /// <summary>
    /// Serves as the entry point for the application.
    /// </summary>
    /// <param name="args">An array of command-line arguments supplied to the application.</param>
    public static void Main(string[] args)
    {
        var builder = WebApplication.CreateBuilder(args);

        builder.ConfigureCore()
            .ConfigureData();

        var app = builder.Build();

        app.UseCore();

        app.Run();
    }
}
