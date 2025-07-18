import fastify from "fastify";
import { connectAmqp } from "./amqp/connection";
import { Consumer } from "./amqp/consumer";
import { DuelStartedHandler } from "./amqp/handlers/DuelStartedHandler";
import { GiftDepositedHandler } from "./amqp/handlers/GiftDepositedHandler";
import { GiftWithdrawFailedHandler } from "./amqp/handlers/GiftWithdrawFailedHandler";
import { GiftWithdrawnHandler } from "./amqp/handlers/GiftWithdrawnHandler";
import { GiftWithdrawUserNotFoundHandler } from "./amqp/handlers/GiftWithdrawUserNotFoundHandler";
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
				routingKey: "gift.withdraw.failed",
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
				routingKey: "gift.withdrawn",
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
		new Consumer(
			{
				exchange: {
					name: "telegram.events",
					type: "topic",
				},
				routingKey: "gift.withdraw.user-not-found",
				maxRetries: 3,
			},
			async (message, properties, ctrl) => {
				new GiftWithdrawUserNotFoundHandler(notificationService).handle(
					message,
					properties,
					ctrl,
				);
			},
		),
		new Consumer(
			{
				exchange: {
					name: "duel.events",
					type: "topic",
				},
				routingKey: "duel.started",
				maxRetries: 3,
			},
			async (message, properties, ctrl) => {
				new DuelStartedHandler(notificationService).handle(
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
