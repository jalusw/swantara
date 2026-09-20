"use client";

import { create } from "zustand";

export type DigitalFingerprintState = {
  deviceName: string | null;
  fingerprint: string | null;
  browser: string | null;
  os: string | null;
};

export type DigitalFingerprintActions = {
  setFingerprint: (fingerprint: string) => void;
  setBrowser: (fingerprint: string) => void;
  setOS: (fingerprint: string) => void;
};

export type DigitalFingerprintStore = DigitalFingerprintState & DigitalFingerprintActions;

const defaultValues: DigitalFingerprintState = {
  deviceName: null,
  fingerprint: null,
  browser: null,
  os: null,
};

export const useDigitalFingerprintStore = create<DigitalFingerprintStore>((set) => ({
  ...defaultValues,
  setFingerprint: (fingerprint: string) => set({ fingerprint }),
  setBrowser: (browser: string) => set({ browser }),
  setOS: (os: string) => set({ os }),
}));
