import { closeAmqp, connectAmqp } from "./amqp/connection";
import { Userbot } from "./core/userbot";
import { setupShutdownHooks } from "./shutdown";
import { nftGiftHandler } from "./telegram/handlers/gift";

const userbot = new Userbot();

async function main() {
	await connectAmqp();
	await userbot.start();

	await nftGiftHandler(userbot.getClient());

	setupShutdownHooks(async () => {
		await closeAmqp();
		await userbot.close();
	});

	// const messages = await userbot
	// 	.getClient()
	// 	.getMessages(404181517, { limit: 10 })
	// for (const msg of messages) {
	// 	if (
	// 		msg instanceof Api.MessageService &&
	// 		msg.action instanceof Api.MessageActionStarGiftUnique
	// 	) {
	// 		logger.info(
	// 			{
	// 				messageId: msg.id,
	// 				giftId: msg.action.gift?.id?.toString(),
	// 				upgradeMsgId: msg.action.gift?.upgradeMsgId?.toString(),
	// 			},
	// 			'🎁 Found StarGift action'
	// 		)
	// 	}
	// }

	// await userbot.transferGift({
	// 	messageId: 139967,
	// 	userId: 404181517,
	// })
}

main().catch((err) => {
	console.error("[FATAL] Unhandled exception:", err);
	process.exit(1);
});
