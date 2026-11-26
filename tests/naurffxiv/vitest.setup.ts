import { beforeEach, vi } from "vitest";

// The app logs through console; keep test output readable.
beforeEach(() => {
  for (const level of ["info", "warn", "error", "debug"] as const) {
    vi.spyOn(console, level).mockImplementation(() => {});
  }
});
