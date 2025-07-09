import { connectAmqp } from "./amqp/connection";
import { Consumer } from "./amqp/consumer";
import { GiftDepositedHandler } from "./amqp/handlers/GiftDepositedHandler";
import { GiftWithdrawFailedHandler } from "./amqp/handlers/GiftWithdrawFailedHandler";
import { GiftWithdrawnHandler } from "./amqp/handlers/GiftWithdrawnHandler";
import { createBot } from "./bot";
import { logger } from "./logger";
import { NotificationService } from "./services/notification";

async function main() {
	await connectAmqp();

	// 1) Сначала создаём бота (пока без notificationService)
	const bot = createBot();

	// 2) Теперь, когда bot уже есть, можно инициализировать NotificationService
	const notificationService = new NotificationService(bot);

	const giftDepositedConsumer = new Consumer(
		{
			exchange: {
				name: "gift.events",
				type: "topic",
			},
			routingKey: "gift.deposited",
			maxRetries: 1,
		},
		async (message, properties, ctrl) => {
			new GiftDepositedHandler(notificationService).handle(
				message,
				properties,
				ctrl,
			);
		},
	);

	const giftWithdrawFailedConsumer = new Consumer(
		{
			exchange: {
				name: "telegram.events",
				type: "topic",
			},
			routingKey: "telegram.gift.withdraw.failed",
			maxRetries: 1,
		},
		async (message, properties, ctrl) => {
			new GiftWithdrawFailedHandler(notificationService).handle(
				message,
				properties,
				ctrl,
			);
		},
	);

	const giftWithdrawnConsumer = new Consumer(
		{
			exchange: {
				name: "telegram.events",
				type: "topic",
			},
			routingKey: "telegram.gift.withdrawn",
		},
		async (message, properties, ctrl) => {
			new GiftWithdrawnHandler(notificationService).handle(
				message,
				properties,
				ctrl,
			);
		},
	);

	await giftDepositedConsumer.start();
	await giftWithdrawFailedConsumer.start();
	await giftWithdrawnConsumer.start();

	// 4) Запускаем polling / webhook’и
	await bot.start();
}

main()
	.then(() => logger.info("Bot started"))
	.catch((err) => logger.error("Bot failed to start", { err }));
