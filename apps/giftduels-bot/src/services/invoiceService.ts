import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { InvoicePaymentEventSchema } from "@giftduels/protobuf-js/giftduels/telegrambot/v1/events_pb";
import {
	StarInvoicePayload,
	StarInvoicePayloadSchema,
} from "@giftduels/protobuf-js/giftduels/telegrambot/v1/invoice_payload_pb";
import { Bot } from "grammy";
import { LabeledPrice } from "grammy/types";
import { publisher } from "@/amqp/publisher";
import { container } from "@/container";
import { ExtendedContext } from "@/types/context";

export class InvoiceService {
	private bot: Bot<ExtendedContext>;
	constructor() {
		this.bot = container.resolve("bot");
	}

	/**
	 * Вызывается после успешного pre_checkout (зарезервировали платёж),
	 * можно тут хранить в БД «pending invoice».
	 */
	async handlePreCheckout(): Promise<boolean> {
		return true;
	}

	/** successful_payment */
	async handleSuccessfulPayment(
		telegramUserId: number,
		payload: string,
		invoiceId: string,
		starsAmount: number,
	) {
		// 1) Опционально: повторно декодим, если нужны детали
		const obj = this.decodeInvoicePayload(payload);

		// 2) Делаем AMQP-событие
		const evt = create(InvoicePaymentEventSchema, {
			invoiceId,
			telegramUserId: { value: BigInt(telegramUserId) },
			starsAmount: { value: starsAmount },
			payload: obj,
		});

		await publisher.publishProto({
			routingKey: "invoice.payment.completed",
			schema: InvoicePaymentEventSchema,
			msg: evt,
			opts: { messageId: invoiceId },
		});
	}

	/**
	 * Создаёт Telegram-инвойс на списание «звёзд» у пользователя,
	 * возвращает ссылку на оплату.
	 */
	async createStarInvoice(params: {
		starsAmount: number;
		title: string;
		description: string;
		payload: string;
	}): Promise<string> {
		const { starsAmount, title, description, payload } = params;

		const prices: LabeledPrice[] = [
			{
				label: "XTR",
				amount: starsAmount,
			},
		];

		const url = await this.bot.api.createInvoiceLink(
			title,
			description,
			payload,
			"",
			"XTR",
			prices,
		);
		return url;
	}

	/**
	 * Принимает готовый объект-пayload (по вашей .proto схеме)
	 * и сериализует его в URL-safe Base64.
	 */
	encodeInvoicePayload(obj: StarInvoicePayload): string {
		// 1) создаём protobuf-сообщение
		const msg = create(StarInvoicePayloadSchema, obj);
		// 2) сериализуем в Uint8Array
		const bin = toBinary(StarInvoicePayloadSchema, msg);
		// 3) Base64URL без padding
		return Buffer.from(bin)
			.toString("base64")
			.replace(/\+/g, "-")
			.replace(/\//g, "_")
			.replace(/=+$/, "");
	}

	/**
	 * Декодирует URL-safe Base64 обратно в объект вашей схемы.
	 */
	decodeInvoicePayload(encoded: string): StarInvoicePayload {
		// 1) дополняем до корректного base64 (padding)
		let b64 = encoded.replace(/-/g, "+").replace(/_/g, "/");
		while (b64.length % 4) b64 += "=";
		// 2) парсим бинарные
		const bin = Buffer.from(b64, "base64");
		// 3) десериализуем в protobuf-сообщение
		return fromBinary(StarInvoicePayloadSchema, bin);
	}
}
