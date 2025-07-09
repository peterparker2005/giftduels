import { fromBinary } from "@bufbuild/protobuf";
import { GiftDepositedEventSchema } from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import type { ConsumerHandler } from "@/amqp/consumer";
import { NotificationService } from "@/services/notification";

export class GiftDepositedHandler {
	constructor(private notification: NotificationService) {}

	public readonly handle: ConsumerHandler = async (raw, _props, ctrl) => {
		const evt = fromBinary(GiftDepositedEventSchema, raw);

		await this.notification.sendGiftDepositedNotification(
			Number(evt.ownerTelegramId?.value),
			{
				giftName: evt.title,
				slug: evt.slug,
			},
		);

		// 3) Подтверждаем очередь
		return ctrl.ack();
	};
}
