import AxeBuilder from "@axe-core/playwright";
import { expect } from "@playwright/test";

export type A11yOptions = {
  ruleTags?: Array<"wcag2a" | "wcag2aa" | "wcag21a" | "wcag21aa" | "wcag22aa">;
  disabledRules?: string[];
  /** Skip known false-positives for placeholder stubs. */
  exclude?: string[];
};

export async function expectAccessible(page: any, options: A11yOptions = {}) {
  const builder = new AxeBuilder({ page });
  if (options.ruleTags) builder.withTags(options.ruleTags);
  if (options.disabledRules) builder.disableRules(options.disabledRules);
  if (options.exclude?.length) builder.exclude(options.exclude);

  const results = await builder.analyze();

  expect(results.violations, formatViolations(results.violations)).toEqual([]);
}

export function formatViolations(
  violations: Array<{
    id: string;
    help: string;
    helpUrl?: string;
    nodes: unknown[];
  }>,
) {
  if (violations.length === 0) return "";
  return violations
    .map(
      (v, i) =>
        `\n${i + 1}. ${v.id} (${v.help})\n   ${v.helpUrl ?? ""}\n   ${v.nodes.length} node(s)`,
    )
    .join("\n");
}
