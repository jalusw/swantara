import { useEffect } from "react";

export function useLineResetEffect<T extends { id: string }>(
  isOpen: boolean,
  setLines: (lines: T[]) => void,
  createEmptyLine: () => T,
) {
  useEffect(() => {
    if (!isOpen) return;
    setLines([createEmptyLine()]);
  }, [isOpen, setLines, createEmptyLine]);
}
