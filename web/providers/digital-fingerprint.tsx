"use client";

import { useEffect } from "react";
import { useDigitalFingerprintStore } from "@/stores/digital-fingerprint.store";

export function DigitalFingerPrintProvider({ children }: { children: React.ReactNode }) {
  const { setFingerprint, setBrowser, setOS } = useDigitalFingerprintStore();

  useEffect(() => {
    import("clientjs")
      .then(({ ClientJS }) => {
        const clientJS = new ClientJS();
        setFingerprint(clientJS.getFingerprint().toString());
        setBrowser(clientJS.getBrowser());
        setOS(clientJS.getOS());
      })
      .catch(() => undefined);
  }, [setBrowser, setFingerprint, setOS]);

  return children;
}
