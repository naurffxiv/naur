const TEMPLATE_MARKER_RE =
  /^<!-- QA Checklist Template Start -->([\s\S]*?)^<!-- QA Checklist Template End -->/im;
const TEMPLATE_HEADER_RE =
  /## Test Checklist Template\s*\n([\s\S]*?)(?=\n## |$)/;

async function fetchUrl(url, retries = 2) {
  for (let attempt = 0; attempt <= retries; attempt++) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 10000);
    try {
      const res = await fetch(url, {
        redirect: "follow",
        signal: controller.signal,
      });
      clearTimeout(timer);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.text();
    } catch (err) {
      clearTimeout(timer);
      if (attempt < retries) {
        await new Promise((r) => setTimeout(r, 1000));
        continue;
      }
      throw err;
    }
  }
}

/**
 * Fetches the QA checklist template from the given URL.
 * Handles GitHub wiki API, raw githubusercontent fallback, and plain HTTP URLs.
 * Returns the extracted template content string, or throws if unavailable.
 */
module.exports = async function fetchTemplate(github, wikiUrl) {
  if (!wikiUrl) throw new Error("QA_WIKI_URL not set");

  let content;

  if (wikiUrl.includes("github.com") && wikiUrl.includes("/wiki")) {
    const cleanUrl = wikiUrl.split("#")[0];
    const parts = cleanUrl.match(
      /github\.com\/([^/]+)\/([^/]+)\/wiki\/?([^?#]*)/,
    );
    if (parts) {
      const [, owner, repo, page] = parts;
      const pageName = page || "Home";
      console.log(`Fetching wiki page: ${owner}/${repo}/${pageName}`);
      try {
        const { data } = await github.rest.repos.getContent({
          owner,
          repo: `${repo}.wiki`,
          path: `${pageName}.md`,
        });
        content = Buffer.from(data.content, "base64").toString();
      } catch (apiErr) {
        console.warn(`Wiki API fetch failed: ${apiErr.message}`);
        const rawUrl = `https://raw.githubusercontent.com/wiki/${owner}/${repo}/${pageName}.md`;
        console.log(`Trying raw wiki URL: ${rawUrl}`);
        try {
          content = await fetchUrl(rawUrl);
        } catch (rawErr) {
          console.warn(`Raw wiki fetch failed: ${rawErr.message}`);
        }
      }
    }
  }

  if (!content) {
    console.log(`Fetching template from: ${wikiUrl}`);
    content = await fetchUrl(wikiUrl);
  }

  if (!content) {
    throw new Error(`Failed to retrieve content from ${wikiUrl}`);
  }

  const markerMatch = content.match(TEMPLATE_MARKER_RE);
  const headerMatch = content.match(TEMPLATE_HEADER_RE);
  const templateContent = markerMatch?.[1]?.trim() || headerMatch?.[1]?.trim();

  if (!templateContent) {
    throw new Error(
      "Could not find QA template markers in the wiki content. Ensure the page contains <!-- QA Checklist Template Start -->...<!-- QA Checklist Template End --> markers or a '## Test Checklist Template' section.",
    );
  }

  return templateContent;
};
