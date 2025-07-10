import fastify from "fastify";
import { connectAmqp } from "./amqp/connection";
import { Consumer } from "./amqp/consumer";
import { GiftDepositedHandler } from "./amqp/handlers/GiftDepositedHandler";
import { GiftWithdrawFailedHandler } from "./amqp/handlers/GiftWithdrawFailedHandler";
import { GiftWithdrawnHandler } from "./amqp/handlers/GiftWithdrawnHandler";
import { getContainer } from "./container";
import { grpcServerPlugin } from "./grpc/plugin";
import { logger } from "./logger";

async function main() {
	const app = fastify({ logger: true });

	await connectAmqp();

	// Получаем сервисы из DI контейнера
	const container = getContainer();
	const bot = container.resolve("bot");
	const notificationService = container.resolve("notificationService");

	await app.register(grpcServerPlugin);
	app.grpcServer.start();

	const consumers = [
		new Consumer(
			{
				exchange: {
					name: "gift.events",
					type: "topic",
				},
				routingKey: "gift.deposited",
				maxRetries: 3,
			},
			async (message, properties, ctrl) => {
				new GiftDepositedHandler(notificationService).handle(
					message,
					properties,
					ctrl,
				);
			},
		),
		new Consumer(
			{
				exchange: {
					name: "telegram.events",
					type: "topic",
				},
				routingKey: "telegram.gift.withdraw.failed",
				maxRetries: 3,
			},
			async (message, properties, ctrl) => {
				new GiftWithdrawFailedHandler(notificationService).handle(
					message,
					properties,
					ctrl,
				);
			},
		),
		new Consumer(
			{
				exchange: {
					name: "telegram.events",
					type: "topic",
				},
				routingKey: "telegram.gift.withdrawn",
				maxRetries: 3,
			},
			async (message, properties, ctrl) => {
				new GiftWithdrawnHandler(notificationService).handle(
					message,
					properties,
					ctrl,
				);
			},
		),
	];

	await Promise.all(consumers.map((consumer) => consumer.start()));

	await bot.start();
	await app.listen({ port: 50061 });
}

main()
	.then(() => logger.info("Bot started"))
	.catch((err) => logger.error({ err }, "Bot failed to start"));
