import { ChannelWrapper } from "amqp-connection-manager";
import { ConfirmChannel, ConsumeMessage, Options } from "amqplib";
import { createChannel } from "@/amqp/connection";
import { config } from "@/config";
import { logger } from "@/logger";

export interface ConsumerOptions {
	exchange: { name: string; type: "direct" | "topic" | "fanout" | "headers" };
	routingKey: string;
	prefetch?: number;
	maxRetries?: number;

	/**
	 * Куда отправлять «отпавшие» сообщения (poison).
	 * Если не указано, будет использоваться `${exchange.name}.poison`
	 */
	poisonExchange?: string;

	/**
	 * routingKey для poison-сообщений.
	 * Если не указано — совпадёт с обычным routingKey
	 */
	poisonRoutingKey?: string;
}

export type AckControl = {
	ack(): void;
	retry(): void;
	fail(): void;
	poison(): void;
};

export type ConsumerHandler = (
	content: Buffer,
	properties: ConsumeMessage["properties"],
	ctrl: AckControl,
) => Promise<void>;

export class Consumer {
	private channel!: ChannelWrapper;
	private queueName!: string;

	constructor(
		private opts: ConsumerOptions,
		private handler: ConsumerHandler,
	) {}

	public async start() {
		this.queueName = `${config.serviceName}.${this.opts.routingKey}`;
		this.channel = createChannel(this.setupChannel.bind(this), { json: false });
		this.channel.on("error", (err) => logger.error({ err }, "Channel error"));
		this.channel.on("close", () => logger.warn("Channel closed"));
	}

	public async stop() {
		await this.channel.close();
		logger.info("Consumer stopped");
	}

	private async setupChannel(ch: ConfirmChannel) {
		const { exchange, routingKey, prefetch = 1 } = this.opts;
		await ch.assertExchange(exchange.name, exchange.type, { durable: true });
		await ch.assertQueue(this.queueName, { durable: true });
		await ch.bindQueue(this.queueName, exchange.name, routingKey);
		await ch.prefetch(prefetch);
		await ch.consume(this.queueName, (msg) => this.onMessage(msg), {
			noAck: false,
		});
		logger.info(
			{ queue: this.queueName, routingKey, prefetch },
			"Consumer ready",
		);
	}

	private async onMessage(rawMsg: ConsumeMessage | null) {
		if (!rawMsg) return;
		const { properties, content, fields } = rawMsg;
		const max = this.opts.maxRetries ?? 3;
		const prev = Number(properties.headers?.["x-attempts"] ?? 0);

		// Готовим все методы управления
		const ctrl: AckControl = {
			ack: () => this.channel.ack(rawMsg),
			retry: () => {
				const newAttempts = prev + 1;
				this.channel
					.publish(this.opts.exchange.name, this.opts.routingKey, content, {
						headers: {
							...properties.headers,
							"x-attempts": newAttempts,
						},
						persistent: true,
						messageId: properties.messageId,
					} as Options.Publish)
					.then(() => {
						logger.info(
							{ messageId: properties.messageId, nextAttempt: newAttempts },
							"Retrying",
						);
						this.channel.ack(rawMsg);
					})
					.catch((err) => {
						logger.error({ err }, "Retry publish failed, dropping");
						this.channel.nack(rawMsg, false, false);
					});
			},
			fail: () => this.channel.nack(rawMsg, false, false),
			poison: () => {
				const pEx =
					this.opts.poisonExchange ?? `${this.opts.exchange.name}.poison`;
				const pKey = this.opts.poisonRoutingKey ?? this.opts.routingKey;
				// републикуем payload и оригинальные заголовки + инфу о попытках
				this.channel
					.publish(pEx, pKey, content, {
						...properties,
						headers: {
							...properties.headers,
							"x-original-routingKey": fields.routingKey,
							"x-attempts": prev,
						},
					} as Options.Publish)
					.then(() => {
						logger.info({ messageId: properties.messageId }, "Sent to poison");
						this.channel.ack(rawMsg);
					})
					.catch((err) => {
						logger.error({ err }, "Failed to publish to poison, dropping");
						this.channel.nack(rawMsg, false, false);
					});
			},
		};

		try {
			// 1) Запускаем бизнес-логику
			await this.handler(content, properties, ctrl);
		} catch (err) {
			logger.error({ err, attempt: prev }, "Handler error");
			if (prev < max) {
				// 3) Переотправляем автоматически с инкрементом заголовка
				const newHeaders = {
					...(properties.headers || {}),
					"x-attempts": prev + 1,
				};
				this.channel
					.publish(this.opts.exchange.name, this.opts.routingKey, content, {
						...properties,
						headers: newHeaders,
						persistent: true,
					} as Options.Publish)
					.then(() => {
						logger.info({ nextAttempt: prev + 1 }, "Requeued for retry");
						ctrl.ack();
					})
					.catch((pubErr) => {
						logger.error({ pubErr }, "Failed to requeue, dropping");
						ctrl.fail();
					});
			} else {
				logger.error({ err, attempt: prev }, "Handler error");
				if (prev < max) ctrl.retry();
				else ctrl.poison();
			}
		}
	}
}
