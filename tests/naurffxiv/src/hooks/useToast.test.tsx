// @vitest-environment happy-dom
import { act, renderHook } from "@testing-library/react";
import { expect, test } from "vitest";
import { useToast } from "@/hooks/ui/useToast";

test("useToast opens with defaults and closes", () => {
  const { result } = renderHook(() => useToast());
  expect(result.current.open).toBe(false);

  act(() => result.current.showToast({ message: "Saved" }));
  expect(result.current).toMatchObject({
    open: true,
    message: "Saved",
    severity: "info",
    durationMs: 5000,
  });

  act(() =>
    result.current.showToast({
      message: "Failed",
      severity: "error",
      durationMs: 100,
    }),
  );
  expect(result.current).toMatchObject({
    message: "Failed",
    severity: "error",
    durationMs: 100,
  });

  act(() => result.current.closeToast());
  expect(result.current.open).toBe(false);
});
