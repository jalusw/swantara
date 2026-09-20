import { cloneElement, isValidElement, type ReactElement, type ReactNode } from "react";

export type PermissionGateMode = "hide" | "disable";

export type PermissionGateProps = {
  required: string | string[];
  granted: string[];
  requireAll?: boolean;
  mode?: PermissionGateMode;
  fallback?: ReactNode;
  children: ReactNode;
};

export function hasPermission(required: string | string[], granted: string[], requireAll = true) {
  const requiredList = Array.isArray(required) ? required : [required];
  if (requiredList.length === 0) {
    return true;
  }
  return requireAll
    ? requiredList.every((code) => granted.includes(code))
    : requiredList.some((code) => granted.includes(code));
}

export function PermissionGate({
  required,
  granted,
  requireAll = true,
  mode = "hide",
  fallback,
  children,
}: PermissionGateProps) {
  const allowed = hasPermission(required, granted, requireAll);

  if (allowed) {
    return <>{children}</>;
  }

  if (mode === "hide") {
    return <>{fallback ?? null}</>;
  }

  if (fallback !== undefined) {
    return <>{fallback}</>;
  }

  if (isValidElement(children)) {
    const child = children as ReactElement<{ disabled?: boolean }>;
    if ("disabled" in child.props) {
      return (
        <span data-slot="permission-gate-disabled" className="opacity-50">
          {cloneElement(child, { disabled: true })}
        </span>
      );
    }
    return (
      <span aria-disabled="true" data-slot="permission-gate-disabled" className="opacity-50">
        {children}
      </span>
    );
  }
  return null;
}
