import { create } from "@bufbuild/protobuf";
import { GiftWithdrawUserNotFoundEventSchema } from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import {
	GiftIdSchema,
	GiftTelegramIdSchema,
	TelegramUserIdSchema,
} from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { Api, TelegramClient } from "telegram";
import { v4 as uuidv4 } from "uuid";
import { Publisher } from "@/amqp/publisher";

/**
 * Параметры перевода подарка, которые нужны как
 * для реального transfer, так и для публикации события
 */
export interface TransferParams {
	userId: number;
	messageId: number;
	giftId: string;
	telegramGiftId: number;
	collectibleId: number;
	title: string;
	slug: string;
}

export class GiftTransferer {
	constructor(
		private client: TelegramClient,
		private publisher: Publisher,
	) {}

	/**
	 * Основной метод — пытается перевести подарок «бесплатно»,
	 * при PAYMENT_REQUIRED идёт в платную ветку,
	 * если peer не найден — сразу публикует UserNotFound.
	 */
	async transferGift(
		peer: Api.TypeInputPeer | undefined,
		params: TransferParams,
	): Promise<void> {
		if (!peer) {
			await this.publishUserNotFound(params);
			return;
		}

		// 1) бесплатный transfer
		try {
			await this.client.invoke(
				new Api.payments.TransferStarGift({
					stargift: new Api.InputSavedStarGiftUser({ msgId: params.messageId }),
					toId: peer,
				}),
			);
			return;
		} catch (err) {
			if (!this.isPaymentRequired(err)) {
				throw err; // какая-то другая ошибка
			}
		}

		// 2) платный transfer (invoice → form → send)
		await this.sendInvoice(peer, params);
	}

	/** Проверяем, что Telegram вернул именно PAYMENT_REQUIRED */
	private isPaymentRequired(err: unknown): boolean {
		const msg = err instanceof Error ? err.message : String(err);
		return msg.includes("PAYMENT_REQUIRED");
	}

	/** Шаги выставления инвойса и оплаты подарка */
	private async sendInvoice(
		peer: Api.TypeInputPeer,
		params: TransferParams,
	): Promise<void> {
		const invoice = new Api.InputInvoiceStarGiftTransfer({
			stargift: new Api.InputSavedStarGiftUser({ msgId: params.messageId }),
			toId: peer,
		});

		// Получаем форму
		const form = await this.client.invoke(
			new Api.payments.GetPaymentForm({ invoice }),
		);

		// Посылаем форму
		await this.client.invoke(
			new Api.payments.SendStarsForm({
				formId: form.formId,
				invoice: invoice,
			}),
		);
	}

	/**
	 * Если peer не найден, публикуем event о том,
	 * что пользователь не стартовал диалог с ботом
	 */
	private async publishUserNotFound(params: TransferParams): Promise<void> {
		const evt = create(GiftWithdrawUserNotFoundEventSchema, {
			$typeName: GiftWithdrawUserNotFoundEventSchema.typeName,
			giftId: create(GiftIdSchema, {
				$typeName: GiftIdSchema.typeName,
				value: params.giftId,
			}),
			ownerTelegramId: create(TelegramUserIdSchema, {
				$typeName: TelegramUserIdSchema.typeName,
				value: BigInt(params.userId),
			}),
			telegramGiftId: create(GiftTelegramIdSchema, {
				$typeName: GiftTelegramIdSchema.typeName,
				value: BigInt(params.telegramGiftId),
			}),
			collectibleId: params.collectibleId,
			title: params.title,
			slug: params.slug,
		});

		await this.publisher.publishProto({
			routingKey: "telegram.gift.withdraw.user.not.found",
			schema: GiftWithdrawUserNotFoundEventSchema,
			msg: evt,
			opts: { messageId: uuidv4() },
		});
	}
}
