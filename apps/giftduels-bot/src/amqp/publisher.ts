import { DescMessage, Message, toBinary } from "@bufbuild/protobuf";
import { ChannelWrapper } from "amqp-connection-manager";
import { ConfirmChannel, Options } from "amqplib";
import { createChannel } from "@/amqp/connection";
import { logger } from "@/logger";

export interface PublishOptions {
	messageId?: string;
	persistent?: boolean;
	mandatory?: boolean;
	headers?: Record<string, unknown>;
}

export class Publisher {
	private channel: ChannelWrapper | null = null;
	private channelPromise: Promise<ChannelWrapper> | null = null;

	constructor(private readonly exchange: string) {
		// Don't create channel immediately - defer until first use
	}

	private async getChannel(): Promise<ChannelWrapper> {
		if (this.channel) return this.channel;

		if (!this.channelPromise) {
			this.channelPromise = this.createChannelWhenReady();
		}

		return this.channelPromise;
	}

	private async createChannelWhenReady(): Promise<ChannelWrapper> {
		// setupChannel will be called automatically on (re)connect
		this.channel = createChannel(this.setupChannel.bind(this), { json: false });
		return this.channel;
	}

	private async setupChannel(ch: ConfirmChannel) {
		// Declare главный exchange вашего сервиса
		await ch.assertExchange(this.exchange, "topic", { durable: true });
		// TODO: тут же можно объявить DLX-exchange для poison-очередей
	}

	/**
	 * Публикует protobuf-сообщение.
	 */
	public async publishProto<T extends Message>(args: {
		routingKey: string;
		schema: DescMessage;
		msg: T;
		opts?: PublishOptions;
	}): Promise<void> {
		const { routingKey, schema, msg, opts } = args;

		const channel = await this.getChannel();

		const buffer = Buffer.from(toBinary(schema, msg));

		const typeName = msg.$typeName;

		const headers = {
			...opts?.headers,
			"x-proto-type": typeName ?? routingKey,
			"x-message-id": opts?.messageId,
		};

		await channel.publish(this.exchange, routingKey, buffer, {
			persistent: opts?.persistent ?? true,
			mandatory: opts?.mandatory ?? false,
			messageId: opts?.messageId,
			headers,
			contentType: "application/x-protobuf",
		} as Options.Publish);
		logger.info(
			{ exchange: this.exchange, routingKey, payloadType: "protobuf" },
			"[AMQP] proto message published",
		);
	}
}

export const publisher = new Publisher("telegrambot.events");
