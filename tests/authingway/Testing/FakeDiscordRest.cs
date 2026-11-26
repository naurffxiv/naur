// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Net;
using System.Text.Json;
using NetCord.Rest;

namespace Naur.Authingway.Tests.Testing;

/// <summary>
/// Answers Discord REST requests in memory and records each interaction response the bot sends.
/// </summary>
public sealed class FakeDiscordRest : IRestRequestHandler
{
    private readonly List<JsonElement> _callbacks = [];

    /// <summary>
    /// Gets the interaction callback bodies the bot sent, in order.
    /// </summary>
    public IReadOnlyList<JsonElement> Callbacks => _callbacks;

    /// <summary>
    /// Gets the content of the latest message response the bot sent.
    /// </summary>
    public string MessageContent => LatestCallback().GetProperty("data").GetProperty("content").GetString()!;

    /// <summary>
    /// Gets the data of the latest modal the bot opened.
    /// </summary>
    public JsonElement Modal
    {
        get
        {
            var callback = LatestCallback();
            if (callback.GetProperty("type").GetInt32() != (int)InteractionCallbackType.Modal)
            {
                throw new InvalidOperationException($"Expected a modal response but got {callback}");
            }

            return callback.GetProperty("data");
        }
    }

    /// <inheritdoc/>
    public async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(request);

        if (request.RequestUri!.AbsolutePath.EndsWith("/callback", StringComparison.Ordinal) && request.Content is not null)
        {
            _callbacks.Add(await ReadJsonPayloadAsync(request.Content, cancellationToken));
        }

        return new HttpResponseMessage(HttpStatusCode.NoContent) { RequestMessage = request };
    }

    /// <inheritdoc/>
    public void AddDefaultHeader(string name, IEnumerable<string> values)
    {
    }

    /// <inheritdoc/>
    public void Dispose()
    {
    }

    private JsonElement LatestCallback()
    {
        return _callbacks.Count > 0
            ? _callbacks[^1]
            : throw new InvalidOperationException("The bot sent no interaction response.");
    }

    private static async Task<JsonElement> ReadJsonPayloadAsync(HttpContent content, CancellationToken cancellationToken)
    {
        // Responses are sent as multipart form data with a payload_json part, or as plain JSON.
        if (content is MultipartContent multipart)
        {
            foreach (var part in multipart)
            {
                if (part.Headers.ContentDisposition?.Name?.Trim('"') == "payload_json")
                {
                    return JsonElement.Parse(await part.ReadAsStringAsync(cancellationToken));
                }
            }
        }

        return JsonElement.Parse(await content.ReadAsStringAsync(cancellationToken));
    }
}
