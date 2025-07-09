import { fromBinary } from "@bufbuild/protobuf";
import { GiftWithdrawFailedEventSchema } from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import type { ConsumerHandler } from "@/amqp/consumer";
import { NotificationService } from "@/services/notification";

export class GiftWithdrawFailedHandler {
	constructor(private notification: NotificationService) {}

	public readonly handle: ConsumerHandler = async (raw, _props, ctrl) => {
		const evt = fromBinary(GiftWithdrawFailedEventSchema, raw);

		await this.notification.sendGiftFailedNotification(
			Number(evt.ownerTelegramId?.value),
			{
				giftName: evt.title,
				slug: evt.slug,
			},
		);

		return ctrl.ack();
	};
}
