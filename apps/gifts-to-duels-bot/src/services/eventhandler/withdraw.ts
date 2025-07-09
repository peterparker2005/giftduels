import { create } from "@bufbuild/protobuf";
import {
	GiftWithdrawFailedEventSchema,
	GiftWithdrawRequestedEvent,
	GiftWithdrawRequestedEventSchema,
} from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import { ConsumeMessage } from "amqplib";
import { v4 as uuidv4 } from "uuid";
import type { AckControl } from "@/amqp/consumer";
import { publisher } from "@/amqp/publisher";
import { logger } from "@/logger";
import { Userbot } from "@/telegram/userbot";
import { decodeProtobufMessage } from "@/utils/decodeProtobufMessage";

const MAX_RETRIES = 3;

export async function handleGiftWithdrawRequested(
	msg: Buffer,
	properties: ConsumeMessage["properties"],
	ctrl: AckControl,
	userbot: Userbot,
): Promise<void> {
	// 1) Сколько уже было попыток?
	const prevAttempts = Number(properties.headers?.["x-attempts"] ?? 0);

	// 2) Пробуем десериализовать
	const event = decodeProtobufMessage<GiftWithdrawRequestedEvent>(
		msg,
		GiftWithdrawRequestedEventSchema,
	);
	if (!event) {
		logger.warn(
			{ messageId: properties.messageId, attempts: prevAttempts },
			"⚠️ Не смогли декодировать Protobuf — убираем из очереди",
		);
		return ctrl.fail(); // сбрасываем без retry
	}

	// 3) Проверяем обязательные поля
	const giftId = event.giftId?.value;
	const ownerTelegramId = event.ownerTelegramId?.value;
	const upgradeMessageId = event.upgradeMessageId;
	if (!giftId || !ownerTelegramId || !upgradeMessageId) {
		logger.error(
			{
				messageId: properties.messageId,
				attempts: prevAttempts,
				giftId,
				ownerTelegramId,
				upgradeMessageId,
			},
			"❌ В теле события нет необходимых полей — дропаем",
		);
		return ctrl.fail();
	}

	const log = logger.child({ giftId, ownerTelegramId, upgradeMessageId });

	try {
		log.info("🎁 Пытаемся вывести подарок через Telegram API");
		await userbot.transferGift({
			userId: Number(ownerTelegramId),
			messageId: upgradeMessageId,
		});
		log.info("✅ Подарок успешно выведен");

		return ctrl.ack();
	} catch (err) {
		log.error({ err, attempts: prevAttempts }, "❌ Ошибка при обработке");

		// transient-ошибка? решаем по числу попыток
		if (prevAttempts < MAX_RETRIES) {
			log.info({ nextAttempt: prevAttempts + 1 }, "🔄 Будем повторять попытку");
			return ctrl.retry();
		} else {
			log.error(
				{ attempts: prevAttempts },
				`⚠️ Превышено ${MAX_RETRIES} повторов — публикуем событие об ошибке`,
			);

			try {
				const failedEvent = create(GiftWithdrawFailedEventSchema, {
					$typeName: GiftWithdrawFailedEventSchema.typeName,
					giftId: event.giftId,
					ownerTelegramId: event.ownerTelegramId,
					telegramGiftId: event.telegramGiftId,
					collectibleId: event.collectibleId,
					upgradeMessageId: event.upgradeMessageId,
					price: event.price,
					commissionAmount: event.commissionAmount,
					errorReason: `Failed after ${MAX_RETRIES} attempts: ${err instanceof Error ? err.message : String(err)}`,
					attemptsMade: prevAttempts + 1,
				});

				const messageId = uuidv4();

				await publisher.publishProto({
					routingKey: "gift.withdraw.failed",
					schema: GiftWithdrawFailedEventSchema,
					msg: failedEvent,
					opts: {
						messageId,
					},
				});

				log.info("📤 Событие об ошибке вывода опубликовано");
			} catch (publishErr) {
				log.error({ publishErr }, "❌ Ошибка при публикации события об ошибке");
			}

			return ctrl.fail();
		}
	}
}
