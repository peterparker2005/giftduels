import { create } from "@bufbuild/protobuf";
import {
	RollDiceResponse,
	RollDiceResponseSchema,
} from "@giftduels/protobuf-js/giftduels/telegrambot/v1/private_service_pb";
import { Bot } from "grammy";
import { config } from "@/config";
import { container } from "@/container";
import { dateToProto } from "@/shared/utils/dateToProto";
import { ExtendedContext } from "@/types/context";

export class DuelService {
	private bot: Bot<ExtendedContext>;
	constructor() {
		this.bot = container.resolve("bot");
	}

	async rollDice(
		telegramUserId: number,
		displayNumber: string,
		duelId: string,
	): Promise<RollDiceResponse> {
		const msg = await this.sendDiceMessage(
			telegramUserId,
			displayNumber,
			duelId,
		);
		const response = create(RollDiceResponseSchema, {
			value: msg.dice.value,
			telegramMessageId: msg.message_id,
			telegramChatId: msg.chat.id.toString(),
			rolledAt: dateToProto(new Date()),
		});
		return response;
	}

	async sendDiceMessage(
		telegramUserId: number,
		displayNumber: string,
		duelId: string,
	) {
		const duelUrl = `tg://devpp2_bot?startapp&href=duels/${duelId}`;

		return await this.bot.api.sendDice(config.telegram.duelChannelId, "🎲", {
			reply_markup: {
				inline_keyboard: [
					[
						{
							text: `Duel #${displayNumber}`,
							url: duelUrl,
						},
					],
					[
						{
							text: `Player`,
							url: `tg://user?id=${telegramUserId}`,
						},
					],
				],
			},
		});
	}
}
