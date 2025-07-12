import { create } from "@bufbuild/protobuf";
import {
	GiftWithdrawFailedEventSchema,
	GiftWithdrawnEventSchema,
	GiftWithdrawRequestedEvent,
	GiftWithdrawRequestedEventSchema,
} from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import { ConsumeMessage } from "amqplib";
import { v4 as uuidv4 } from "uuid";
import type { AckControl } from "@/amqp/consumer";
import { publisher } from "@/amqp/publisher";
import { logger } from "@/logger";
import { Userbot } from "@/telegram/userbot/Userbot";
import { decodeProtobufMessage } from "@/utils/decodeProtobufMessage";

const MAX_RETRIES = 3;

// Хелпер для создания failedEvent - используется во всех случаях ошибок
async function createAndPublishFailedEvent(
	event: GiftWithdrawRequestedEvent,
	errorReason: string,
	attemptsMade: number,
	log: typeof logger,
): Promise<void> {
	try {
		const failedEvent = create(GiftWithdrawFailedEventSchema, {
			$typeName: GiftWithdrawFailedEventSchema.typeName,
			giftId: event.giftId,
			ownerTelegramId: event.ownerTelegramId,
			telegramGiftId: event.telegramGiftId,
			collectibleId: event.collectibleId,
			title: event.title,
			slug: event.slug,
			upgradeMessageId: event.upgradeMessageId,
			price: event.price,
			commissionAmount: event.commissionAmount,
			errorReason,
			attemptsMade,
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

		log.info("📤 Событие об ошибке опубликовано в poison queue");
	} catch (publishErr) {
		log.error(
			{ publishErr },
			"❌ Критическая ошибка при публикации failedEvent",
		);
	}
}

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
			"❌ В теле события нет необходимых полей — создаем failedEvent",
		);

		await createAndPublishFailedEvent(
			event,
			`Missing required fields: giftId=${giftId}, ownerTelegramId=${ownerTelegramId}, upgradeMessageId=${upgradeMessageId}`,
			prevAttempts + 1,
			logger,
		);

		return ctrl.fail();
	}

	const log = logger.child({ giftId, ownerTelegramId, upgradeMessageId });

	try {
		log.info("🎁 Пытаемся вывести подарок через Telegram API");
		await userbot.transferGift({
			userId: Number(ownerTelegramId),
			messageId: upgradeMessageId,
			giftId: giftId,
			telegramGiftId: Number(event.telegramGiftId?.value),
			collectibleId: event.collectibleId,
			title: event.title,
			slug: event.slug,
		});

		const withdrawnEvent = create(GiftWithdrawnEventSchema, {
			$typeName: GiftWithdrawnEventSchema.typeName,
			slug: event.slug,
			title: event.title,
			giftId: event.giftId,
			ownerTelegramId: event.ownerTelegramId,
			telegramGiftId: event.telegramGiftId,
			collectibleId: event.collectibleId,
		});

		await publisher.publishProto({
			routingKey: "telegram.gift.withdrawn",
			schema: GiftWithdrawnEventSchema,
			msg: withdrawnEvent,
			opts: {
				messageId: uuidv4(),
			},
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
				`⚠️ Превышено ${MAX_RETRIES} повторов — отправляем в poison queue`,
			);

			await createAndPublishFailedEvent(
				event,
				`Failed after ${MAX_RETRIES} attempts: ${err instanceof Error ? err.message : String(err)}`,
				prevAttempts + 1,
				log,
			);

			return ctrl.fail();
		}
	}
}
