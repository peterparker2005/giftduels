import { DescMessage, Message, toBinary } from "@bufbuild/protobuf";
import type { ChannelWrapper } from "amqp-connection-manager";
import type { ConfirmChannel } from "amqplib";
import { getConnection } from "@/amqp/connection";
import { logger } from "@/logger";

const EXCHANGE_MAIN = "telegram.events";
const MAX_RETRY_ATTEMPTS = 5;
const BASE_BACKOFF_MS = 500;

/* ---------- ленивый канал ---------- */
let channelPromise: Promise<ChannelWrapper> | undefined;

async function getChannel(): Promise<ChannelWrapper> {
	if (channelPromise) return channelPromise; // уже создан

	channelPromise = getConnection().then(async (conn) => {
		const ch = conn.createChannel({
			json: false,
			confirm: true,
			setup: async (c: ConfirmChannel) => {
				await c.assertExchange(EXCHANGE_MAIN, "topic", { durable: true });
			},
		});
		await ch.waitForConnect(); // дожидаемся open
		logger.info("[AMQP] channel ready");
		return ch;
	});

	return channelPromise;
}

/* ---------- retry helper ---------- */
async function withRetry(
	action: () => Promise<unknown>,
	rk: string,
	attempt = 0,
): Promise<void> {
	try {
		await action();
	} catch (err) {
		if (attempt >= MAX_RETRY_ATTEMPTS) {
			logger.error({ err, rk }, "❌ Publish failed after retries");
			throw err;
		}
		const delay = BASE_BACKOFF_MS * 2 ** attempt;
		logger.warn({ rk, delay, attempt }, "🔁 Publish retry");
		await new Promise((r) => setTimeout(r, delay));
		return withRetry(action, rk, attempt + 1);
	}
}

/* ---------- JSON publisher ---------- */
export async function publish<T>(opts: {
	routingKey: string;
	body: T;
	headers?: Record<string, unknown>;
	messageId?: string;
}): Promise<void> {
	const payload = Buffer.from(JSON.stringify(opts.body));
	const headers = { "x-message-id": opts.messageId, ...opts.headers };

	const ch = await getChannel();
	await withRetry(
		() =>
			ch
				.publish(EXCHANGE_MAIN, opts.routingKey, payload, {
					persistent: true,
					contentType: "application/json",
					headers,
				})
				.then(() => undefined),
		opts.routingKey,
	);
}

export async function publishProto<T extends Message>(opts: {
	routingKey: string;
	schema: DescMessage;
	msg: T; // экземпляр protobuf-es сообщения
	headers?: Record<string, unknown>;
	messageId?: string;
}) {
	const buffer = Buffer.from(toBinary(opts.schema, opts.msg));

	const typeName = opts.msg.$typeName;

	const headers = {
		"x-proto-type": typeName ?? opts.routingKey,
		"x-message-id": opts.messageId,
		...opts.headers,
	};

	const ch = await getChannel();
	await withRetry(
		() =>
			ch
				.publish(EXCHANGE_MAIN, opts.routingKey, buffer, {
					persistent: true,
					contentType: "application/x-protobuf",
					headers,
				})
				.then(() => void 0),
		opts.routingKey,
	);
}
