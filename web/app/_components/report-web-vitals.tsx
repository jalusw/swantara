"use client";

import { useReportWebVitals } from "next/web-vitals";
import { getCsrfToken } from "@/lib/constants/cookies";
import { logger } from "@/lib/utils/logger";

const VITALS_ENDPOINT = "/api/vitals";

type VitalMetric = Parameters<Parameters<typeof useReportWebVitals>[0]>[0];

function sendMetric(metric: VitalMetric) {
  const payload = JSON.stringify({
    id: metric.id,
    name: metric.name,
    value: metric.value,
    rating: metric.rating,
    path: window.location.pathname,
  });

  fetch(VITALS_ENDPOINT, {
    method: "POST",
    body: payload,
    keepalive: true,
    headers: {
      "Content-Type": "application/json",
      "x-csrf-token": getCsrfToken() ?? "",
    },
  }).catch(() => undefined);
}

export function ReportWebVitals() {
  useReportWebVitals((metric) => {
    if (process.env.NODE_ENV === "development") {
      logger.info(`[Web Vitals] ${metric.name}: ${metric.value}`);
      return;
    }
    sendMetric(metric);
  });
  return null;
}
