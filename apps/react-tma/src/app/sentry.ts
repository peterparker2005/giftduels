import * as Sentry from "@sentry/react";
import { logger } from "@/shared/logger";
import { config } from "./config";

logger.debug("initializing sentry", {
	dsn: config.sentry.dsn,
});

Sentry.init({
	dsn: config.sentry.dsn,
	integrations: [Sentry.browserTracingIntegration()],
	tracesSampleRate: 1,
	environment: import.meta.env.MODE,
	sendDefaultPii: true,
});
