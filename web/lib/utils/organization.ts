import type { Organization } from "@/lib/services/swantara";

export type OrganizationPatch = {
  name?: string;
  legalName?: string;
  parentId?: number | null;
  baseCurrency?: string;
  countryCode?: string | null;
  taxId?: string;
  timezone?: string;
  taxYearStartMonth?: number;
};

/**
 * Builds a full PUT /organizations/:id payload from the current organization
 * plus the fields a form wants to change. The backend overwrites legal name,
 * parent, country and tax id unconditionally, so callers must always send the
 * current values for the fields they are not editing.
 */
export function toOrganizationUpdateRequest(
  org: Organization,
  patch: OrganizationPatch = {},
): {
  name: string;
  legalName: string;
  parentId: number | null;
  baseCurrency: string;
  countryCode: string | null;
  taxId: string;
  timezone: string;
  taxYearStartMonth: number;
} {
  return {
    name: patch.name ?? org.name,
    legalName: patch.legalName ?? org.legalName ?? "",
    parentId: patch.parentId !== undefined ? patch.parentId : (org.parentId ?? null),
    baseCurrency: patch.baseCurrency ?? org.baseCurrency,
    countryCode: patch.countryCode !== undefined ? patch.countryCode : (org.countryCode ?? null),
    taxId: patch.taxId ?? org.taxId ?? "",
    timezone: patch.timezone ?? org.timezone,
    taxYearStartMonth: patch.taxYearStartMonth ?? org.taxYearStartMonth,
  };
}
