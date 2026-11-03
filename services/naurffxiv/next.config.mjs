// next.config.mjs
import path from "path";
import { fileURLToPath } from "url";
import createMDX from "@next/mdx";
import rehypeAutolinkHeadings from "rehype-autolink-headings";
import rehypeImgSize from "rehype-img-size";
import rehypeSlug from "rehype-slug";
import remarkFrontmatter from "remark-frontmatter";
import remarkToc from "remark-toc";
import { DEVICE_SIZES, IMAGE_SIZES } from "./src/lib/imageVariants.mjs";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const withMDX = createMDX({
  extension: /\.mdx?$/,
  options: {
    remarkPlugins: [remarkFrontmatter, remarkToc],
    rehypePlugins: [
      rehypeSlug,
      [rehypeImgSize, { dir: "public" }],
      [
        rehypeAutolinkHeadings,
        {
          behaviour: "append",
          properties: {
            ariaHidden: true,
            tabIndex: -1,
            className: "hash-link",
          },
        },
      ],
    ],
  },
});

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  pageExtensions: ["js", "jsx", "ts", "tsx", "md", "mdx"],
  images: {
    // pre-generated variants served as static files (see src/lib/imageLoader.ts)
    loader: "custom",
    loaderFile: "./src/lib/imageLoader.ts",
    deviceSizes: DEVICE_SIZES,
    imageSizes: IMAGE_SIZES,
    remotePatterns: [
      {
        protocol: "https",
        hostname: "xivapi.com",
        port: "",
        pathname: "/i/**",
        search: "",
      },
      {
        protocol: "https",
        hostname: "cdn.discordapp.com",
        port: "",
        pathname: "/avatars/**",
      },
      {
        protocol: "https",
        hostname: "cdn.discordapp.com",
        port: "",
        pathname: "/embed/avatars/**",
      },
    ],
  },
  outputFileTracingRoot: path.join(__dirname, "../../"),
  async redirects() {
    return [
      {
        source: "/ultimates/:slug",
        destination: "/ultimate/:slug",
        permanent: true,
      },
      {
        source: "/ultimate/dmu",
        destination: "/ultimate/kefka",
        permanent: true,
      },
      {
        source: "/ultimate/dmu/:slug*",
        destination: "/ultimate/kefka/:slug*",
        permanent: true,
      },
      {
        source: "/ultimate/umad",
        destination: "/ultimate/kefka",
        permanent: true,
      },
      {
        source: "/ultimate/umad/:slug*",
        destination: "/ultimate/kefka/:slug*",
        permanent: true,
      },
    ];
  },
};

export default withMDX(nextConfig);
