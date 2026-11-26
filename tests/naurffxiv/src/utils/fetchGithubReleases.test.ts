import { afterEach, expect, test, vi } from "vitest";
import { fetchGithubReleases } from "@/utils/fetchGithubReleases";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

function stubFetch(response: Partial<Response>) {
  const fetchMock = vi.fn().mockResolvedValue(response);
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

test("fetchGithubReleases requests the repo releases and drops entries without a tag", async () => {
  const fetchMock = stubFetch({
    ok: true,
    json: async () => [
      { tag_name: "v1.0.0" },
      { tag_name: "" },
      null,
      { name: "untagged" },
    ],
  });

  await expect(fetchGithubReleases()).resolves.toEqual([
    { tag_name: "v1.0.0" },
  ]);
  expect(fetchMock).toHaveBeenCalledWith(
    "https://api.github.com/repos/naurffxiv/naurffxiv/releases",
    expect.objectContaining({
      headers: { Accept: "application/vnd.github+json" },
    }),
  );
});

test("fetchGithubReleases logs and rethrows when GitHub responds with an error", async () => {
  stubFetch({ ok: false, status: 403, statusText: "Forbidden" });
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  await expect(fetchGithubReleases()).rejects.toThrow(
    "GitHub API responded with 403: Forbidden",
  );
  expect(consoleError).toHaveBeenCalledWith(
    "Error fetching GitHub releases:",
    expect.any(Error),
  );
});
