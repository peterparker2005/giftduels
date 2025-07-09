import { closeAmqp, connectAmqp } from "@/amqp/connection";
import { Userbot } from "@/telegram/userbot";
import { Consumer } from "./amqp/consumer";
import { handleGiftWithdrawRequested } from "./services/eventhandler/withdraw";
import { setupShutdownHooks } from "./shutdown";
import { nftGiftHandler } from "./telegram/handlers/gift";

async function main() {
	// 1) AMQP
	await connectAmqp();

	// 2) Telegram
	const userbot = new Userbot();
	await userbot.start();

	// userbot
	// 	.getClient()
	// 	.getMessages(404181517, { limit: 1 })
	// 	.then((messages) => {
	// 		console.log(messages);
	// 	});
	// 3) Telegram-хэндлеры
	await nftGiftHandler(userbot.getClient());

	// // 4) AMQP-хэндлеры
	const giftWithdrawRequestedConsumer = new Consumer(
		{
			exchange: {
				name: "gift.events",
				type: "topic",
			},
			maxRetries: 3,
			prefetch: 3,
			routingKey: "gift.withdraw.requested",
		},
		(msg, properties, ctrl) =>
			handleGiftWithdrawRequested(msg, properties, ctrl, userbot),
	);

	await giftWithdrawRequestedConsumer.start();

	// 5) graceful shutdown
	setupShutdownHooks(async () => {
		await userbot.close();
		await closeAmqp();
	});
}

main().catch((err) => {
	console.error("[FATAL] Unhandled exception:", err);
	process.exit(1);
});
