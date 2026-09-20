interface PasswordCredentialData {
  id: string;
  password: string;
  name?: string;
  iconURL?: string;
  origin?: string;
}

interface PasswordCredential extends Credential {
  readonly id: string;
  readonly password: string;
}

interface PasswordCredentialConstructor {
  new (data: PasswordCredentialData): PasswordCredential;
}

declare global {
  const PasswordCredential: PasswordCredentialConstructor;
  interface CredentialRequestOptions {
    password?: boolean;
    mediation?: "silent" | "optional" | "required";
  }
  interface CredentialsContainer {
    get(options?: CredentialRequestOptions): Promise<Credential | null>;
    store(credential: Credential): Promise<void>;
    preventSilentAccess?: () => Promise<void>;
  }
}

export {};
