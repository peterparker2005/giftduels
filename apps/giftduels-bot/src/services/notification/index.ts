import type { Bot } from "grammy";
import { InlineKeyboard } from "grammy";
import { config } from "@/config";
import { container } from "@/container";
import { logger } from "@/logger";
import { ExtendedContext } from "@/types/context";

type NotificationMessage = {
	text: string;
	parseMode: "HTML";
	keyboard?: InlineKeyboard;
};

export class NotificationService {
	private bot: Bot<ExtendedContext>;

	constructor() {
		this.bot = container.resolve("bot");
	}

	private async send(telegramUserId: number, msg: NotificationMessage) {
		try {
			await this.bot.api.sendMessage(telegramUserId, msg.text, {
				parse_mode: msg.parseMode,
				reply_markup: msg.keyboard,
			});
			logger.info(
				{ telegramUserId, textLen: msg.text.length },
				"Notification sent",
			);
		} catch (err) {
			logger.error({ err, telegramUserId }, "Failed to send notification");
		}
	}

	async sendGiftDepositedNotification(
		telegramUserId: number,
		payload: { giftName: string; slug: string },
	) {
		const { giftName, slug } = payload;
		const text = `🎁 <a href="https://t.me/nft/${slug}">${giftName}</a> was successfully deposited to your profile!

Find game or create your own`;

		const keyboard = new InlineKeyboard().webApp(
			"🚀 Launch App",
			config.telegram.webAppUrl,
		);
		await this.send(telegramUserId, { text, keyboard, parseMode: "HTML" });
	}

	async sendGiftWithdrawnNotification(
		telegramUserId: number,
		payload: { giftName: string; slug: string },
	) {
		const { giftName, slug } = payload;
		const text = `🎁 <a href="https://t.me/nft/${slug}">${giftName}</a> was successfully withdrawn!`;

		await this.send(telegramUserId, { text, parseMode: "HTML" });
	}

	async sendGiftFailedNotification(
		telegramUserId: number,
		payload: { giftName: string; slug: string },
	) {
		const { giftName, slug } = payload;
		const text = `⚠️ Your withdrawal for <a href="https://t.me/nft/${slug}">${giftName}</a> couldn't be completed.

It might be technical issue or temporary Telegram API limitation.

Please try again later or contact support if the issue persists 👉 @GiftDuelsHelp`;

		await this.send(telegramUserId, { text, parseMode: "HTML" });
	}

	async sendGiftWithdrawUserNotFoundNotification(telegramUserId: number) {
		const text = `⚠️ Warning! It seems you haven't interacted with our bot for a while, so we can't send you the gift.

Please send a sticker and try withdrawing your gift again.`;

		await this.send(telegramUserId, { text, parseMode: "HTML" });
	}

	async sendDuelStartedNotification(
		telegramUserId: number,
		duelId: string,
		totalStakeValue: string,
	) {
		const text = `🎲 Duel for you joined just started!
		
		Watch it now! Total stakes: ${totalStakeValue} TON`;

		const keyboard = new InlineKeyboard().webApp(
			"🎲 Watch Duel",
			`${config.telegram.webAppUrl}?startapp&href=duels/${duelId}`,
		);
		await this.send(telegramUserId, {
			text,
			keyboard,
			parseMode: "HTML",
		});
	}
}
