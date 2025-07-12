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

	async rollDice(telegramUserId: number): Promise<RollDiceResponse> {
		const msg = await this.sendDiceMessage(telegramUserId);
		const response = create(RollDiceResponseSchema, {
			value: msg.dice.value,
			telegramMessageId: msg.message_id,
			telegramChatId: msg.chat.id.toString(),
			rolledAt: dateToProto(new Date()),
		});
		return response;
	}

	async sendDiceMessage(telegramUserId: number) {
		return await this.bot.api.sendDice(config.telegram.duelChannelId, "🎲", {
			reply_markup: {
				inline_keyboard: [
					[
						{
							text: "Duel #0",
							url: "tg://devpp2_bot?startapp&href=duels/id",
						},
					],
					[
						{
							text: "Player",
							url: `tg://user?id=${telegramUserId}`,
						},
					],
				],
			},
		});
	}
}
