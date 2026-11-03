import { glob } from "glob";
import fs from "fs";
import path from "path";
import sharp from "sharp";
import {
  DEVICE_SIZES,
  IMAGE_SIZES,
  variantUrl,
} from "../src/lib/imageVariants.mjs";

/*
  pre-generates every public image at each next/image width (as webp), so
  src/lib/imageLoader.ts can serve them as static files instead of going
  through /_next/image, which runs the Worker on every image request

  runs as part of `pnpm build`, or on its own with `pnpm images:variants`
  output goes to public/_variants (gitignored)
  variants newer than their source image are skipped
  images are never enlarged: widths past the original get an original-size copy
*/

const PUBLIC_DIR = "public";
const WIDTHS = [...IMAGE_SIZES, ...DEVICE_SIZES];
const QUALITY = 75; // matches next/image's default
const CONCURRENCY = 4;

let generated = 0;
let skipped = 0;

async function generateVariants(imagePath: string): Promise<void> {
  const src = "/" + path.posix.relative(PUBLIC_DIR, imagePath);
  const sourceTime = fs.statSync(imagePath).mtimeMs;
  const image = sharp(imagePath);

  for (const width of WIDTHS) {
    const url = variantUrl(src, width);
    if (!url) {
      continue;
    }
    const outputPath = path.join(PUBLIC_DIR, url);
    if (
      fs.existsSync(outputPath) &&
      fs.statSync(outputPath).mtimeMs >= sourceTime
    ) {
      skipped++;
      continue;
    }

    fs.mkdirSync(path.dirname(outputPath), { recursive: true });
    await image
      .clone()
      .resize({ width, withoutEnlargement: true })
      .webp({ quality: QUALITY })
      .toFile(outputPath);
    generated++;
  }
}

async function main(): Promise<void> {
  const started = Date.now();
  const images = await glob(`${PUBLIC_DIR}/**/*.{png,jpg,jpeg,webp,avif}`, {
    ignore: `${PUBLIC_DIR}/_variants/**`,
    nocase: true,
    posix: true,
  });

  // a few images at a time; some sources are 5000px+ wide
  const queue = [...images];
  await Promise.all(
    Array.from({ length: CONCURRENCY }, async () => {
      for (let next = queue.shift(); next; next = queue.shift()) {
        await generateVariants(next);
      }
    }),
  );

  console.log(
    `image variants: ${generated} generated, ${skipped} up to date, ` +
      `${images.length} images x ${WIDTHS.length} widths ` +
      `(${((Date.now() - started) / 1000).toFixed(1)}s)`,
  );
}

main().catch((err: unknown) => {
  console.error(err);
  process.exit(1);
});
