import { fromBinary } from "@bufbuild/protobuf";
import { GiftWithdrawUserNotFoundEventSchema } from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import type { ConsumerHandler } from "@/amqp/consumer";
import { NotificationService } from "@/services/notification";

export class GiftWithdrawUserNotFoundHandler {
	constructor(private notification: NotificationService) {}

	public readonly handle: ConsumerHandler = async (raw, _props, ctrl) => {
		const evt = fromBinary(GiftWithdrawUserNotFoundEventSchema, raw);

		await this.notification.sendGiftWithdrawUserNotFoundNotification(
			Number(evt.ownerTelegramId?.value),
		);

		return ctrl.ack();
	};
}
