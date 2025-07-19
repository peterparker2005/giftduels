import { fromBinary } from "@bufbuild/protobuf";
import { DuelStartedEventSchema } from "@giftduels/protobuf-js/giftduels/duel/v1/event_pb";
import type { ConsumerHandler } from "@/amqp/consumer";
import { NotificationService } from "@/services/notification";

export class DuelStartedHandler {
	constructor(private notification: NotificationService) {}

	public readonly handle: ConsumerHandler = async (raw, _props, ctrl) => {
		const evt = fromBinary(DuelStartedEventSchema, raw);

		if (!evt.duelId?.value) {
			throw new Error("Duel ID is required");
		}

		if (!evt.totalStakeValue?.value) {
			throw new Error("Total stake value is required");
		}

		for (const participant of evt.participants) {
			await this.notification.sendDuelStartedNotification(
				Number(participant.telegramUserId?.value),
				evt.duelId?.value,
				evt.totalStakeValue?.value,
			);
		}

		// 3) Подтверждаем очередь
		return ctrl.ack();
	};
}
