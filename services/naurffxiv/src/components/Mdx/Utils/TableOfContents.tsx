"use client";

import { useState, useEffect, useCallback, useMemo, ReactElement } from "react";
import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";
import type { Toc } from "@stefanprobst/rehype-extract-toc";

interface MdxFrontmatter {
  collapseToc?: boolean;
}
interface TableOfContentsProps {
  toc: Toc;
  frontmatter: MdxFrontmatter;
}

/**
 * Custom hook to track the active section ID based on scroll position.
 */
function useActiveId(ids: string[], offset: number = 88): string {
  const [activeId, setActiveId] = useState("");

  const updateActiveId = useCallback((): void => {
    let currentActiveId = "";
    for (const id of ids) {
      const element = document.getElementById(id);
      if (element) {
        const rect = element.getBoundingClientRect();
        // If the top of the element is above or at the offset point, it's a candidate.
        // Since we iterate in DOM order, the last one found will be the most specific
        if (rect.top <= offset + 5) {
          currentActiveId = id;
        } else {
          break;
        }
      }
    }
    setActiveId(currentActiveId);
  }, [ids, offset]);

  useEffect((): (() => void) => {
    window.addEventListener("scroll", updateActiveId, { passive: true });
    updateActiveId();
    return () => window.removeEventListener("scroll", updateActiveId);
  }, [updateActiveId]);

  return activeId;
}

function recursiveToc(
  toc: Toc,
  collapse: boolean | undefined,
  activeId: string,
  level: number = 0,
): ReactElement {
  // ignore first level headers, head straight to h2
  const currLevel = level
    ? toc.map((li) => (
        <li key={li.id}>
          <div className="w-fit">
            <a
              href={`#${li.id}`}
              className={`toc-links ${activeId === li.id ? "toc-current" : ""}`}
              onClick={(e): void => {
                e.preventDefault();
                const element = document.getElementById(li.id!);
                if (element) {
                  // Offset 88px aligns heading text exactly below the navbar.
                  const offset = 88;
                  const elementPosition = element.getBoundingClientRect().top;
                  const offsetPosition =
                    elementPosition + window.scrollY - offset;

                  window.scrollTo({
                    top: offsetPosition,
                    behavior: "smooth",
                  });

                  // Update URL fragment without jumping
                  window.history.pushState(null, "", `#${li.id}`);
                }
              }}
            >
              <div className="toc-url-container">
                {li.value}
                {collapse && li.children ? (
                  <span className="ml-1">
                    <ArrowDropDownIcon
                      className={`transition-all ${
                        activeId === li.id ? "" : "-rotate-90"
                      }`}
                    />
                  </span>
                ) : (
                  <></>
                )}
              </div>
            </a>
          </div>
          <div>
            {li.children ? (
              recursiveToc(li.children, collapse, activeId, level + 1)
            ) : (
              <></>
            )}
          </div>
        </li>
      ))
    : toc.map((li) => (
        <div key={li.id}>
          {li.children ? (
            recursiveToc(li.children, collapse, activeId, level + 1)
          ) : (
            <></>
          )}
        </div>
      ));

  return level ? (
    <ul className="list-none ps-4 toc">{currLevel}</ul>
  ) : (
    <>{currLevel}</>
  );
}

export default function TableOfContents({
  toc,
  frontmatter,
}: TableOfContentsProps): ReactElement {
  const allIds = useMemo((): string[] => {
    const ids: string[] = [];
    const recurse = (entries: Toc): void => {
      entries.forEach((entry) => {
        if (entry.id !== undefined) ids.push(entry.id);
        if (entry.children) recurse(entry.children);
      });
    };
    recurse(toc);
    return ids;
  }, [toc]);

  const activeId = useActiveId(allIds);

  return (
    <nav className={`toc${frontmatter.collapseToc ? " toc-collapse" : ""}`}>
      {recursiveToc(toc, frontmatter.collapseToc, activeId)}
    </nav>
  );
}
