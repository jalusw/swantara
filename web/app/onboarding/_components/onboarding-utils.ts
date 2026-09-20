import type { QuestionnaireItemDefinition } from "@shadcn/react/questionnaire";

export const ONBOARDING_ITEMS: readonly QuestionnaireItemDefinition[] = [
  { name: "companyInfo", required: true },
];

export function resolveStandardCode(countryCode: string): string {
  if (countryCode.toUpperCase() !== "ID") {
    return "IFRS";
  }
  return "PSAK_EMKM";
}
