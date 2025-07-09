import { createLogger } from "@giftduels/logger-ts";
import { config } from "./config";

const log = createLogger({
	env: config.isDev ? "development" : "production",
	level: config.logLevel,
});

export const logger = log.child({
	service: "gifts-to-duels-bot",
});
