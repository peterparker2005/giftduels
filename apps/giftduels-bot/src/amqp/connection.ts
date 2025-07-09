import {
	AmqpConnectionManager,
	AmqpConnectionManagerOptions,
	ChannelWrapper,
	connect,
} from "amqp-connection-manager";
import { ConfirmChannel } from "amqplib";
import { config } from "@/config";
import { logger } from "@/logger";

let connection: AmqpConnectionManager;
let readyResolver: (conn: AmqpConnectionManager) => void;
const readyPromise = new Promise<AmqpConnectionManager>((res) => {
	readyResolver = res;
});

export async function connectAmqp(): Promise<AmqpConnectionManager> {
	if (connection) return readyPromise;

	const url = config.amqp.url();
	logger.info(`[AMQP] Connecting to ${url.replace(/:.*@/, ":***@")}...`);

	connection = connect([url], {
		heartbeatIntervalInSeconds: 15,
		reconnectTimeInSeconds: 5,
	} satisfies AmqpConnectionManagerOptions);

	connection.on("connect", () => {
		logger.info("[AMQP] connected");
		readyResolver(connection);
	});
	connection.on("disconnect", ({ err }) =>
		logger.warn({ err }, "[AMQP] disconnected"),
	);
	connection.on("error", (err) =>
		logger.error({ err }, "[AMQP] connection error"),
	);

	return readyPromise;
}

export function createChannel(
	onSetup: (channel: ConfirmChannel) => Promise<void>,
	opts?: { json?: boolean },
): ChannelWrapper {
	if (!connection) {
		throw new Error(
			"AMQP connection is not established, call connectAmqp() first",
		);
	}

	const channel = connection.createChannel({
		json: opts?.json ?? false,
		setup: onSetup,
	});

	channel.on("error", (err) => logger.error({ err }, "[AMQP] channel error"));
	channel.on("close", () => logger.info("[AMQP] channel closed"));

	return channel;
}

export async function closeAmqp() {
	if (connection) {
		await connection.close();
	}
}
