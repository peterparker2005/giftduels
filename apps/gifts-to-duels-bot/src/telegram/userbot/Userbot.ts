import { Api, TelegramClient } from "telegram";
import { StringSession } from "telegram/sessions";
import { publisher } from "@/amqp/publisher";
import { config } from "@/config";
import { logger } from "@/logger";
import { GiftFetcher } from "@/telegram/userbot/GiftFetcher";
import {
	GiftTransferer,
	TransferParams,
} from "@/telegram/userbot/GiftTransferer";
import { PeerResolver } from "@/telegram/userbot/PeerResolver";

export class Userbot {
	private client: TelegramClient;
	public peerResolver: PeerResolver;
	public giftFetcher: GiftFetcher;
	public giftTransferer: GiftTransferer;

	constructor() {
		// 1) Инициализируем TelegramClient
		this.client = new TelegramClient(
			new StringSession(config.telegram.sessionString),
			config.telegram.apiId,
			config.telegram.apiHash,
			{ connectionRetries: 5 },
		);

		// 2) Собираем сервисы поверх client
		this.peerResolver = new PeerResolver(this.client);
		this.giftFetcher = new GiftFetcher(this.client);
		this.giftTransferer = new GiftTransferer(this.client, publisher);
	}

	/** Соединяемся и проверяем, что бот авторизован */
	async start(): Promise<void> {
		await this.client.connect();
		await this.client.getMe();
		logger.info("✅ Userbot connected");
	}

	/** Корректно закрыть сессию */
	async close(): Promise<void> {
		await this.client.destroy();
		logger.info("🔌 Userbot disconnected");
	}

	/** Проксируем чистый client, чтобы его могли использовать обычные Telegram-хэндлеры */
	getClient(): TelegramClient {
		return this.client;
	}

	/**
	 * Обёртка над GiftFetcher: принимает username|userId или undefined,
	 * резолвит peer и отдаёт список подарков.
	 */
	async getUserGifts(
		user: string | number | undefined,
		limit: number,
	): Promise<{
		total: number;
		gifts: Api.SavedStarGift[];
		users: Api.TypeUser[];
	}> {
		const peer: Api.TypeInputPeer =
			user == null
				? new Api.InputPeerSelf()
				: await this.peerResolver.resolve(user);
		return this.giftFetcher.getUserGifts(peer, limit);
	}

	/**
	 * Обёртка над GiftTransferer: резолвит peer по params.userId
	 * и вызывает transferGift.
	 */
	async transferGift(params: TransferParams): Promise<void> {
		const peer = await this.peerResolver.resolve(params.userId);
		return this.giftTransferer.transferGift(peer, params);
	}
}
