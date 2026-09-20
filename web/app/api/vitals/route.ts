import type { NextRequest } from "next/server";
import { isCsrfValid } from "@/lib/server/csrf";
import { withHandler } from "@/lib/server/handler";
import { logger } from "@/lib/server/logger";
import { createForbiddenResponse, createNoContentResponse } from "@/lib/server/response";

export type VitalMetricPayload = {
  id: string;
  name: string;
  value: number;
  rating?: string;
  path?: string;
};

export const POST = withHandler("POST", "/api/vitals", async (request: NextRequest) => {
  if (!isCsrfValid(request)) {
    return createForbiddenResponse("CSRF token missing.");
  }

  const metric = (await request.json()) as Partial<VitalMetricPayload>;

  if (metric.name && typeof metric.value === "number") {
    logger.info(
      {
        vital: metric.name,
        value: metric.value,
        rating: metric.rating,
        path: metric.path,
        id: metric.id,
      },
      `[Web Vitals] ${metric.name}: ${metric.value}`,
    );
  }

  return createNoContentResponse();
});
