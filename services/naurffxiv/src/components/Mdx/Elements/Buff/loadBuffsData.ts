import fs from "fs";
import path from "path";
import { parse } from "smol-toml";

import { BuffMap } from "./types";

/**
 * Return loaded icon data from either json or toml
 * */
export function loadBuffsData(
  mdxDir: string,
  datapath: string | undefined = "",
): BuffMap {
  // If `datapath` is not explicitly passed:
  if (!datapath) {
    if (fs.existsSync(path.join(mdxDir, "buffs.json"))) {
      // First attempt to grab the default .json file
      datapath = "buffs.json";
    } else if (fs.existsSync(path.join(mdxDir, "buffs.toml"))) {
      // Second attempt to grab default .toml file
      datapath = "buffs.toml";
    } else {
      // No file exists, return empty data
      return {};
    }
  }

  // mdxDir is always confined to src/markdown (see getMdxDir), Turbopack
  // just can't prove that across the prop chain it's threaded through
  const fileAsString = String(
    fs.readFileSync(
      /*turbopackIgnore: true*/ path.join(
        /*turbopackIgnore: true*/ mdxDir,
        datapath,
      ),
    ),
  );

  return datapath.endsWith(".json")
    ? JSON.parse(fileAsString)
    : (parse(fileAsString) as BuffMap);
}
