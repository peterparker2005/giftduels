import { TelegramGiftReceivedEventSchema } from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import { connectAmqp } from "@/amqp/connection";
import { publisher } from "@/amqp/publisher";
import { parseSavedStarGiftToEvent } from "@/domain/gift";
import { Userbot } from "@/telegram/userbot";

async function seed() {
	await connectAmqp();
	const userbot = new Userbot();

	await userbot.start();

	const { gifts } = await userbot.getUserGifts("@GiftsToPortals", 10);

	for (const savedGift of gifts) {
		const ownerTelegramId = 7350079261;

		const event = parseSavedStarGiftToEvent(savedGift, ownerTelegramId);

		await publisher.publishProto({
			routingKey: "gift.received",
			schema: TelegramGiftReceivedEventSchema,
			msg: event,
		});
	}

	await userbot.close();
}

seed().catch((err) => {
	console.error("[FATAL] Unhandled exception:", err);
	process.exit(1);
});
