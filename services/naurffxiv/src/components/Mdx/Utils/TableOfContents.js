"use client";

import React, { useState, useEffect, useCallback, useMemo } from "react";
import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";

/**
 * Custom hook to track the active section ID based on scroll position.
 */
function useActiveId(ids, offset = 88) {
  const [activeId, setActiveId] = useState("");

  const updateActiveId = useCallback(() => {
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

  useEffect(() => {
    window.addEventListener("scroll", updateActiveId, { passive: true });
    updateActiveId();
    return () => window.removeEventListener("scroll", updateActiveId);
  }, [updateActiveId]);

  return activeId;
}

function recursiveToc(toc, collapse, activeId, level = 0) {
  // ignore first level headers, head straight to h2
  const currLevel = level
    ? toc.map((li) => (
        <li key={li.id}>
          <div className="w-fit">
            <a
              href={`#${li.id}`}
              className={`toc-links ${activeId === li.id ? "toc-current" : ""}`}
              onClick={(e) => {
                e.preventDefault();
                const element = document.getElementById(li.id);
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

export default function TableOfContents({ toc, frontmatter }) {
  const allIds = useMemo(() => {
    const ids = [];
    const recurse = (entries) => {
      entries.forEach((entry) => {
        ids.push(entry.id);
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
