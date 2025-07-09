import { fromBinary } from "@bufbuild/protobuf";
import { GiftWithdrawnEventSchema } from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import type { ConsumerHandler } from "@/amqp/consumer";
import { NotificationService } from "@/services/notification";

export class GiftWithdrawnHandler {
	constructor(private notification: NotificationService) {}

	public readonly handle: ConsumerHandler = async (raw, _props, ctrl) => {
		const evt = fromBinary(GiftWithdrawnEventSchema, raw);

		await this.notification.sendGiftWithdrawnNotification(
			Number(evt.ownerTelegramId?.value),
			{
				giftName: evt.title,
				slug: evt.slug,
			},
		);

		return ctrl.ack();
	};
}
