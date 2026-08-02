"use client";

import "./mdx.css";
import "./buffs.scss";
import "./highlights.scss";

import type { ReactNode, ReactElement } from "react";
import type { Toc } from "@stefanprobst/rehype-extract-toc";
import type { SiblingPage } from "@/components/Mdx/Layout/QuickLinks";

import MdxLayout from "@/components/Mdx/Layout/MdxLayout";
import MobileDrawer from "@/components/Mdx/Layout/MobileDrawer";
import QuickLinks from "@/components/Mdx/Layout/QuickLinks";
import TableOfContents from "@/components/Mdx/Utils/TableOfContents";

interface MDXPageProps {
  children: ReactNode;
  toc?: Toc;
  siblingData: SiblingPage[];
  slug: string;
  frontmatter?: Record<string, unknown>;
}

function MDXPage({
  children,
  toc,
  siblingData,
  slug,
  frontmatter,
}: MDXPageProps): ReactElement {
  const quickLinks = <QuickLinks siblingData={siblingData} slug={slug} />;

  return (
    <MdxLayout>
      <div className="grid grid-cols-1 lg:grid-cols-[90ch_1fr] xl:grid-cols-[1fr_90ch_1fr] max-w-screen-2xl mx-auto">
        <aside className="top-[5.5rem] self-start hidden xl:block sticky h-[calc(100vh-100px)] scrollbar">
          <TableOfContents
            toc={toc!}
            frontmatter={frontmatter as { collapseToc?: boolean }}
          />
        </aside>
        <main className="max-w-[90ch] min-h-screen prose prose-invert m-auto mx-6 pt-8">
          {children}
        </main>
        <aside className="prose prose-invert top-[5.5rem] self-start hidden lg:block sticky h-[calc(100vh-100px)] scrollbar">
          {quickLinks}
        </aside>
        <MobileDrawer>{quickLinks}</MobileDrawer>
      </div>
    </MdxLayout>
  );
}

export default MDXPage;
