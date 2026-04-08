import {
  type ReactElement,
  type ElementType,
  type ComponentPropsWithoutRef,
} from "react";
import Banner from "../Elements/Banner";
import { Buff } from "../Elements/Buff/Buff";
import { BuffAppendix } from "../Elements/Buff/BuffAppendix";
import Callout from "../Elements/Callout";
import { CopyToClipboard } from "@/components/Mdx/Utils/CopyToClipboard";
import Details from "../Elements/Details";
import ImageModal from "../Utils/ImageModal";
import LastUpdated from "../Elements/LastUpdated";
import Streamable from "../Elements/Video/Streamable";
import TwitchClip from "../Elements/Video/TwitchClip";
import TwitchVoD from "../Elements/Video/TwitchVoD";
import UnderConstruction from "../Elements/UnderConstruction";
import YouTube from "../Elements/Video/YouTube";

interface MDXComponentsResult {
  [key: string]: ElementType | undefined;
}

export default function MDXComponents(
  mdxDir: string,
  lastUpdated: string | null | undefined,
): MDXComponentsResult {
  // NB: We track the first H1 to ensure <LastUpdated /> only renders once,
  // right after the main page title (the first <h1> in the MDX file).
  // We don't want the "Last Updated" date showing up after every section title.
  let firstH1Rendered = false;

  return {
    a: (props: ComponentPropsWithoutRef<"a">): ReactElement => {
      if (typeof props.href === "string" && props.href.startsWith("http"))
        return <a target="_blank" rel="noopener noreferrer" {...props} />;
      return <a {...props} />;
    },
    h1: (props: ComponentPropsWithoutRef<"h1">): ReactElement => {
      const isFirst = !firstH1Rendered;
      if (isFirst) {
        firstH1Rendered = true;
      }
      return (
        <>
          <h1 className="scroll-mt-[5.5rem]" {...props} />
          {isFirst && <LastUpdated lastUpdated={lastUpdated} />}
        </>
      );
    },
    img: (props: ComponentPropsWithoutRef<typeof ImageModal>): ReactElement => (
      <ImageModal {...props} />
    ),
    pre: (props: ComponentPropsWithoutRef<"pre">): ReactElement => (
      <CopyToClipboard>
        <pre {...props}></pre>
      </CopyToClipboard>
    ),
    Banner,
    ImageModal,
    YouTube,
    TwitchClip,
    TwitchVoD,
    Streamable,
    Buff: ({
      mdxDir: _,
      ...props
    }: ComponentPropsWithoutRef<typeof Buff>): ReactElement => (
      <Buff mdxDir={mdxDir} {...props} />
    ),
    BuffAppendix: ({
      mdxDir: _,
      ...props
    }: ComponentPropsWithoutRef<typeof BuffAppendix>): ReactElement => (
      <BuffAppendix mdxDir={mdxDir} {...props} />
    ),
    UnderConstruction,
    Callout,
    Details,
  };
}
