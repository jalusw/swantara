import type { AxiosInstance } from "axios";
import { endpoints } from "./endpoints";
import type {
  Contact,
  ContactAddress,
  ContactAddressRequest,
  ContactBankAccount,
  ContactBankAccountRequest,
  CreateContactRequest,
  CustomerProfile,
  CustomerProfileRequest,
  ListQuery,
  Member,
  MemberRole,
  Organization,
  Permission,
  Profile,
  PublicUser,
  SuccessEnvelope,
  SupplierProfile,
  SupplierProfileRequest,
  UpdateContactRequest,
  User,
  UserSession,
} from "./types";
import { withListMeta } from "./types";

export type LoginRequest = {
  email: string;
  password: string;
};

export type TokenPair = {
  accessToken: string;
  refreshToken: string;
};

export type RegisterRequest = {
  firstName: string;
  lastName?: string;
  email: string;
  password: string;
};

export class Auth {
  constructor(private readonly axios: AxiosInstance) {}

  async login(request: LoginRequest): Promise<TokenPair> {
    const response = await this.axios.post<SuccessEnvelope<TokenPair>>(
      endpoints.auth.login,
      request,
    );
    return withListMeta(response.data);
  }

  async register(request: RegisterRequest): Promise<{ user: User }> {
    const response = await this.axios.post<SuccessEnvelope<{ user: User }>>(
      endpoints.auth.register,
      request,
    );
    return withListMeta(response.data);
  }

  async checkEmail(email: string): Promise<{ available: boolean }> {
    const response = await this.axios.post<SuccessEnvelope<{ available: boolean }>>(
      endpoints.auth.checkEmail,
      { email },
    );
    return withListMeta(response.data);
  }

  async refresh(refreshToken: string): Promise<TokenPair> {
    const response = await this.axios.post<SuccessEnvelope<TokenPair>>(endpoints.auth.refresh, {
      refreshToken,
    });
    return withListMeta(response.data);
  }

  async logout(refreshToken: string): Promise<void> {
    await this.axios.post(endpoints.auth.logout, { refreshToken });
  }

  async requestEmailVerification(email: string): Promise<void> {
    await this.axios.post(endpoints.auth.requestEmailVerification, { email });
  }

  async verifyEmail(token: string): Promise<void> {
    await this.axios.post(endpoints.auth.verifyEmail, { token });
  }

  async requestPasswordReset(email: string): Promise<void> {
    await this.axios.post(endpoints.auth.requestPasswordReset, { email });
  }

  async resetPassword(token: string, password: string): Promise<void> {
    await this.axios.post(endpoints.auth.resetPassword, { token, password });
  }

  async sessions(): Promise<{ sessions: UserSession[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ sessions: UserSession[] }>>(
      endpoints.auth.sessions,
    );
    return withListMeta(response.data);
  }

  async revokeSession(id: number): Promise<void> {
    await this.axios.delete(endpoints.auth.revokeSession(String(id)));
  }
}

export class Me {
  constructor(private readonly axios: AxiosInstance) {}

  async me(): Promise<{ user: Profile }> {
    const response = await this.axios.get<SuccessEnvelope<{ user: Profile }>>(endpoints.me.me);
    return withListMeta(response.data);
  }

  async organizations(): Promise<{ organizations: Organization[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ organizations: Organization[] }>>(
      endpoints.me.organizations,
    );
    return withListMeta(response.data);
  }

  async permissions(organizationId: number): Promise<{ permissions: Permission[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ permissions: Permission[] }>>(
      endpoints.me.permissions(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async updateAvatar(file: File): Promise<void> {
    const formData = new FormData();
    formData.append("file", file);
    await this.axios.put(endpoints.me.avatar, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  }
}

export type CreateUserRequest = {
  firstName: string;
  lastName?: string;
  email: string;
  password: string;
};

export type UpdateUserRequest = {
  firstName?: string;
  lastName?: string;
  avatar?: string;
  bio?: string;
  birthday?: string;
  active?: boolean;
  sex?: string;
  address?: string;
  city?: string;
  postalCode?: string;
};

export class Users {
  constructor(private readonly axios: AxiosInstance) {}

  async list(params?: ListQuery): Promise<{ users: PublicUser[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ users: PublicUser[] }>>(
      endpoints.users.list,
      { params },
    );
    return withListMeta(response.data);
  }

  async search(query: string): Promise<{ users: PublicUser[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ users: PublicUser[] }>>(
      endpoints.users.list,
      { params: { q: query } },
    );
    return withListMeta(response.data);
  }

  async get(id: number): Promise<{ user: PublicUser }> {
    const response = await this.axios.get<SuccessEnvelope<{ user: PublicUser }>>(
      endpoints.users.get(String(id)),
    );
    return withListMeta(response.data);
  }

  async create(request: CreateUserRequest): Promise<{ user: User }> {
    const response = await this.axios.post<SuccessEnvelope<{ user: User }>>(
      endpoints.users.create,
      request,
    );
    return withListMeta(response.data);
  }

  async update(id: number, request: UpdateUserRequest): Promise<{ user: User }> {
    const response = await this.axios.put<SuccessEnvelope<{ user: User }>>(
      endpoints.users.update(String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(id: number): Promise<void> {
    await this.axios.delete(endpoints.users.delete(String(id)));
  }
}

export type CreateMemberRequest = {
  userId: number;
  organizationId: number;
  position?: string;
};

export class Members {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ members: Member[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ members: Member[] }>>(
      endpoints.members.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, request: CreateMemberRequest): Promise<{ member: Member }> {
    const response = await this.axios.post<SuccessEnvelope<{ member: Member }>>(
      endpoints.members.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.members.delete(String(organizationId), String(id)));
  }
}

export type CreateMemberRoleRequest = {
  name: string;
  code: string;
  description?: string;
};

export class MemberRoles {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number): Promise<{ roles: MemberRole[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ roles: MemberRole[] }>>(
      endpoints.members.roles.list(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateMemberRoleRequest,
  ): Promise<{ role: MemberRole }> {
    const response = await this.axios.post<SuccessEnvelope<{ role: MemberRole }>>(
      endpoints.members.roles.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class Permissions {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number): Promise<{ permissions: Permission[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ permissions: Permission[] }>>(
      endpoints.members.permissions(String(organizationId)),
    );
    return withListMeta(response.data);
  }
}

export type PublicOrganization = {
  id: number;
  name: string;
  logo: string | null;
  website: string | null;
  description: string;
};

export type OrganizationRequest = {
  name: string;
  legalName: string;
  description?: string;
  logo?: string;
  address?: string;
  website?: string;
  city?: string;
  postalCode?: string;
  languageId?: number;
  countryId?: number;
  continentId?: number;
  timezoneId?: number;
  currencyId?: number;
  accountingStandardId?: number;
  industryId?: number;
  legalFormId?: number;
  businessId?: string;
  taxId?: string;
};

export type QuickCreateOrganizationRequest = {
  name: string;
  countryCode: string;
  standardCode?: string;
};

export type OrgModule = {
  moduleId: string;
  active: boolean;
};

export type UpdateOrgModuleRequest = {
  moduleId: string;
  active: boolean;
};

export class Organizations {
  constructor(private readonly axios: AxiosInstance) {}

  async list(params?: ListQuery): Promise<{ organizations: PublicOrganization[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ organizations: PublicOrganization[] }>>(
      endpoints.organizations.list,
      { params },
    );
    return withListMeta(response.data);
  }

  async get(id: number): Promise<{ organization: PublicOrganization }> {
    const response = await this.axios.get<SuccessEnvelope<{ organization: PublicOrganization }>>(
      endpoints.organizations.get(String(id)),
    );
    return withListMeta(response.data);
  }

  async create(request: OrganizationRequest): Promise<{ organization: Organization }> {
    const response = await this.axios.post<SuccessEnvelope<{ organization: Organization }>>(
      endpoints.organizations.create,
      request,
    );
    return withListMeta(response.data);
  }

  async quickCreate(
    request: QuickCreateOrganizationRequest,
  ): Promise<{ organization: Organization }> {
    const response = await this.axios.post<SuccessEnvelope<{ organization: Organization }>>(
      endpoints.organizations.quickCreate,
      request,
    );
    return withListMeta(response.data);
  }

  async listModules(organizationId: number): Promise<{ modules: OrgModule[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ modules: OrgModule[] }>>(
      endpoints.organizations.modules(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async updateModule(
    organizationId: number,
    request: UpdateOrgModuleRequest,
  ): Promise<{ module: OrgModule }> {
    const response = await this.axios.put<SuccessEnvelope<{ module: OrgModule }>>(
      endpoints.organizations.updateModule(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(id: number, request: OrganizationRequest): Promise<{ organization: Organization }> {
    const response = await this.axios.put<SuccessEnvelope<{ organization: Organization }>>(
      endpoints.organizations.update(String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(id: number): Promise<void> {
    await this.axios.delete(endpoints.organizations.delete(String(id)));
  }
}

export class Contacts {
  constructor(private readonly axios: AxiosInstance) {}

  addresses = {
    list: async (
      organizationId: number,
      contactId: number,
      params?: ListQuery,
    ): Promise<{ addresses: ContactAddress[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ addresses: ContactAddress[] }>>(
        endpoints.contacts.addresses.list(String(organizationId), String(contactId)),
        { params },
      );
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      contactId: number,
      request: ContactAddressRequest,
    ): Promise<{ address: ContactAddress }> => {
      const response = await this.axios.post<SuccessEnvelope<{ address: ContactAddress }>>(
        endpoints.contacts.addresses.create(String(organizationId), String(contactId)),
        request,
      );
      return withListMeta(response.data);
    },

    update: async (
      organizationId: number,
      contactId: number,
      addressId: number,
      request: ContactAddressRequest,
    ): Promise<{ address: ContactAddress }> => {
      const response = await this.axios.put<SuccessEnvelope<{ address: ContactAddress }>>(
        endpoints.contacts.addresses.update(
          String(organizationId),
          String(contactId),
          String(addressId),
        ),
        request,
      );
      return withListMeta(response.data);
    },

    delete: async (organizationId: number, contactId: number, addressId: number): Promise<void> => {
      await this.axios.delete(
        endpoints.contacts.addresses.delete(
          String(organizationId),
          String(contactId),
          String(addressId),
        ),
      );
    },

    setDefault: async (
      organizationId: number,
      contactId: number,
      addressId: number,
    ): Promise<{ address: ContactAddress }> => {
      const response = await this.axios.post<SuccessEnvelope<{ address: ContactAddress }>>(
        endpoints.contacts.addresses.setDefault(
          String(organizationId),
          String(contactId),
          String(addressId),
        ),
      );
      return withListMeta(response.data);
    },
  };

  bankAccounts = {
    list: async (
      organizationId: number,
      contactId: number,
      params?: ListQuery,
    ): Promise<{ bankAccounts: ContactBankAccount[] }> => {
      const response = await this.axios.get<
        SuccessEnvelope<{ bankAccounts: ContactBankAccount[] }>
      >(endpoints.contacts.bankAccounts.list(String(organizationId), String(contactId)), {
        params,
      });
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      contactId: number,
      request: ContactBankAccountRequest,
    ): Promise<{ bankAccount: ContactBankAccount }> => {
      const response = await this.axios.post<SuccessEnvelope<{ bankAccount: ContactBankAccount }>>(
        endpoints.contacts.bankAccounts.create(String(organizationId), String(contactId)),
        request,
      );
      return withListMeta(response.data);
    },

    update: async (
      organizationId: number,
      contactId: number,
      accountId: number,
      request: ContactBankAccountRequest,
    ): Promise<{ bankAccount: ContactBankAccount }> => {
      const response = await this.axios.put<SuccessEnvelope<{ bankAccount: ContactBankAccount }>>(
        endpoints.contacts.bankAccounts.update(
          String(organizationId),
          String(contactId),
          String(accountId),
        ),
        request,
      );
      return withListMeta(response.data);
    },

    delete: async (organizationId: number, contactId: number, accountId: number): Promise<void> => {
      await this.axios.delete(
        endpoints.contacts.bankAccounts.delete(
          String(organizationId),
          String(contactId),
          String(accountId),
        ),
      );
    },
  };

  async list(organizationId: number, params?: ListQuery): Promise<{ contacts: Contact[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ contacts: Contact[] }>>(
      endpoints.contacts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ contact: Contact }> {
    const response = await this.axios.get<SuccessEnvelope<{ contact: Contact }>>(
      endpoints.contacts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateContactRequest,
  ): Promise<{ contact: Contact }> {
    const response = await this.axios.post<SuccessEnvelope<{ contact: Contact }>>(
      endpoints.contacts.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateContactRequest,
  ): Promise<{ contact: Contact }> {
    const response = await this.axios.put<SuccessEnvelope<{ contact: Contact }>>(
      endpoints.contacts.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.contacts.delete(String(organizationId), String(id)));
  }

  async customer(
    organizationId: number,
    contactId: number,
    request: CustomerProfileRequest,
  ): Promise<{ customer: CustomerProfile }> {
    const response = await this.axios.put<SuccessEnvelope<{ customer: CustomerProfile }>>(
      endpoints.contacts.customer(String(organizationId), String(contactId)),
      request,
    );
    return withListMeta(response.data);
  }

  async supplier(
    organizationId: number,
    contactId: number,
    request: SupplierProfileRequest,
  ): Promise<{ supplier: SupplierProfile }> {
    const response = await this.axios.put<SuccessEnvelope<{ supplier: SupplierProfile }>>(
      endpoints.contacts.supplier(String(organizationId), String(contactId)),
      request,
    );
    return withListMeta(response.data);
  }
}
