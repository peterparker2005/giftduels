import { Bot } from "grammy";
import { config } from "./config";
import { rootRouter } from "./routers";
import { ExtendedContext } from "./types/context";

export function createBot() {
	const bot = new Bot<ExtendedContext>(config.telegram.botToken);

	bot.use(async (_ctx, next) => {
		return next();
	});

	bot.use(rootRouter);
	return bot;
}
