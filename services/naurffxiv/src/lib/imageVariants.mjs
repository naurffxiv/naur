// Image widths the site serves, shared by next.config.mjs (next/image sizes),
// scripts/image-variants.ts (pre-generates each public image at every width)
// and src/lib/imageLoader.ts (points next/image at those files).
export const DEVICE_SIZES = [640, 828, 1200, 1920, 2560];
export const IMAGE_SIZES = [64, 128, 256, 384];

export const VARIANT_DIR = "_variants";

const RESIZABLE = /\.(png|jpe?g|webp|avif)$/i;

/**
 * URL of the pre-generated variant of a public image, e.g.
 * "/images/fru/banner.avif" at 828 -> "/_variants/images/fru/banner.avif.828.webp".
 * Null for images that aren't pre-generated: remote, bundled (/_next/static),
 * svg and gif.
 *
 * @param {string} src
 * @param {number} width
 * @returns {string | null}
 */
export function variantUrl(src, width) {
  if (
    !src.startsWith("/") ||
    src.startsWith("//") ||
    src.startsWith("/_next/") ||
    !RESIZABLE.test(src)
  ) {
    return null;
  }
  return `/${VARIANT_DIR}${src}.${width}.webp`;
}
