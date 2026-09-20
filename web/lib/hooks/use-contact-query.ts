"use client";

import { useQuery } from "@tanstack/react-query";
import { getSwantaraService } from "@/lib/services/swantara";

export function useContactQuery(orgId: string, contactId: string) {
  return useQuery({
    queryKey: ["contact", orgId, contactId],
    queryFn: () => getSwantaraService().contacts.get(Number(orgId), Number(contactId)),
    enabled: Boolean(orgId && contactId),
    select: (data) => data.contact,
  });
}

export function useContactAddressesQuery(orgId: string, contactId: string) {
  return useQuery({
    queryKey: ["contactAddresses", orgId, contactId],
    queryFn: () => getSwantaraService().contacts.addresses.list(Number(orgId), Number(contactId)),
    enabled: Boolean(orgId && contactId),
    select: (data) => data.addresses,
  });
}

export function useContactBankAccountsQuery(orgId: string, contactId: string) {
  return useQuery({
    queryKey: ["contactBankAccounts", orgId, contactId],
    queryFn: () =>
      getSwantaraService().contacts.bankAccounts.list(Number(orgId), Number(contactId)),
    enabled: Boolean(orgId && contactId),
    select: (data) => data.bankAccounts,
  });
}
