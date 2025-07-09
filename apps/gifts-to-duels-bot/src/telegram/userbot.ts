import { create } from "@bufbuild/protobuf";
import { GiftWithdrawUserNotFoundEventSchema } from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import {
	GiftIdSchema,
	GiftTelegramIdSchema,
	TelegramUserIdSchema,
} from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { Api, TelegramClient } from "telegram";
import { StringSession } from "telegram/sessions";
import { v4 as uuidv4 } from "uuid";
import { publisher } from "@/amqp/publisher";
import { config } from "@/config";
import { logger } from "@/logger";

export class Userbot {
	private client: TelegramClient;

	constructor() {
		this.client = new TelegramClient(
			new StringSession(config.telegram.sessionString),
			config.telegram.apiId,
			config.telegram.apiHash,
			{ connectionRetries: 5 },
		);
	}

	async start() {
		await this.client.connect();
		await this.client.getMe();
		logger.info("✅ Userbot connected");
	}

	async close() {
		await this.client.destroy();
	}

	getClient(): TelegramClient {
		return this.client;
	}

	/**
	 * Robustly resolve a user ID to InputPeer with multiple fallback strategies
	 */
	private async getInputPeerForUser(
		userId: number,
	): Promise<Api.TypeInputPeer> {
		try {
			// Try direct getInputEntity first
			const inputPeer = await this.client.getInputEntity(userId);
			if (
				inputPeer instanceof Api.InputPeerUser ||
				inputPeer instanceof Api.InputPeerChannel
			) {
				return inputPeer;
			}
		} catch (e) {
			logger.warn(
				`⚠️ getInputEntity(${userId}) failed: ${(e as Error).message}`,
			);
		}

		try {
			// Try to get entity first, then input entity
			const entity = await this.client.getEntity(userId);
			const inputPeer = await this.client.getInputEntity(entity);
			if (
				inputPeer instanceof Api.InputPeerUser ||
				inputPeer instanceof Api.InputPeerChannel
			) {
				return inputPeer;
			}
		} catch (e) {
			logger.warn(`⚠️ getEntity(${userId}) failed: ${(e as Error).message}`);
		}

		// Try to find user in dialogs/chats
		try {
			const dialogs = await this.client.getDialogs({ limit: 100 });
			for (const dialog of dialogs) {
				if (
					dialog.entity instanceof Api.User &&
					Number(dialog.entity.id) === userId
				) {
					const inputPeer = await this.client.getInputEntity(dialog.entity);
					if (inputPeer instanceof Api.InputPeerUser) {
						logger.info(`✅ Found user ${userId} in dialogs`);
						return inputPeer;
					}
				}
			}
		} catch (e) {
			logger.warn(
				`⚠️ Could not find user ${userId} in dialogs: ${(e as Error).message}`,
			);
		}

		// Last resort: throw error with helpful message
		const error = new Error(
			`Could not resolve user entity for ID ${userId}. The user may need to start a conversation with the bot first or be found in a mutual chat.`,
		);
		logger.error({ userId }, error.message);
		throw error;
	}

	async getUserGifts(params?: {
		user?: string | number; // username | userId | undefined (self)
		limit?: number;
	}): Promise<{
		total: number;
		gifts: Api.SavedStarGift[];
		users: Api.TypeUser[];
	}> {
		const { user, limit = 10 } = params ?? {};

		/* ---------- 1. Готовим peer ---------- */

		const peer: Api.TypeInputPeer = await (async () => {
			// self
			if (!user) return new Api.InputPeerSelf();

			// ----- username -----
			if (typeof user === "string" && user.startsWith("@")) {
				// 1) ResolveUsername
				try {
					const { peer } = await this.client.invoke(
						new Api.contacts.ResolveUsername({ username: user.slice(1) }),
					);
					if (
						peer instanceof Api.InputPeerUser ||
						peer instanceof Api.InputPeerChannel
					) {
						return peer;
					}
				} catch (e) {
					logger.warn(
						`⚠️ ResolveUsername(${user}) не сработал: ${(e as Error).message}`,
					);
				}

				// 2) getEntity → getInputEntity
				try {
					const entity = await this.client.getEntity(user);
					const inputPeer = await this.client.getInputEntity(entity);
					if (
						inputPeer instanceof Api.InputPeerUser ||
						inputPeer instanceof Api.InputPeerChannel
					) {
						return inputPeer;
					}
				} catch (e) {
					logger.warn(
						`⚠️ getEntity(${user}) не сработал: ${(e as Error).message}`,
					);
				}

				// 3) fallback → self (будет пустой список)
				logger.warn(
					`⚠️ Не удалось получить InputPeer для ${user}. Верну 0 gifts.`,
				);
				return new Api.InputPeerSelf();
			}

			// ----- numeric userId -----
			try {
				const ent = await this.client.getInputEntity(user);
				if (
					ent instanceof Api.InputPeerUser ||
					ent instanceof Api.InputPeerChannel
				) {
					return ent;
				}
			} catch (e) {
				logger.warn(
					`⚠️ getInputEntity(${user}) не сработал: ${(e as Error).message}`,
				);
			}

			logger.warn(
				`⚠️ Не удалось получить InputPeerUser/Channel для id=${user}. Верну 0 gifts.`,
			);
			return new Api.InputPeerSelf();
		})();

		// self-peer + указанный user → не получилось резолвить
		if (peer instanceof Api.InputPeerSelf && user) {
			return { total: 0, gifts: [], users: [] };
		}

		/* ---------- 2. Пагинация ---------- */

		let offset = "";
		const gifts: Api.SavedStarGift[] = [];
		const usersMap = new Map<number, Api.TypeUser>();
		let total = 0;
		const pageSize = Math.min(limit, 100); // Telegram всё-равно режет до 100

		do {
			const res = await this.client.invoke(
				new Api.payments.GetSavedStarGifts({
					peer,
					offset,
					limit: pageSize,
				}),
			);

			if (total === 0) total = res.count;

			gifts.push(...res.gifts);
			for (const u of res.users) usersMap.set(Number(u.id), u);

			// если уже собрали нужное количество — выходим
			if (gifts.length >= limit) break;

			offset = res.nextOffset ?? "";
		} while (offset);

		logger.info(
			`📦 Loaded ${gifts.length}/${total} gifts for ${(
				user ?? "me"
			).toString()}`,
		);

		return { total, gifts, users: [...usersMap.values()] };
	}

	async transferGift(params: {
		userId: number;
		messageId: number;
		giftId: string;
		telegramGiftId: number;
		collectibleId: number;
		title: string;
		slug: string;
	}) {
		const {
			userId,
			messageId,
			giftId,
			telegramGiftId,
			collectibleId,
			title,
			slug,
		} = params;
		const peer = await this.getInputPeerForUser(userId);
		if (!peer) {
			const event = create(GiftWithdrawUserNotFoundEventSchema, {
				$typeName: GiftWithdrawUserNotFoundEventSchema.typeName,
				giftId: create(GiftIdSchema, {
					$typeName: GiftIdSchema.typeName,
					value: giftId,
				}),
				ownerTelegramId: create(TelegramUserIdSchema, {
					$typeName: TelegramUserIdSchema.typeName,
					value: BigInt(userId),
				}),
				telegramGiftId: create(GiftTelegramIdSchema, {
					$typeName: GiftTelegramIdSchema.typeName,
					value: BigInt(telegramGiftId),
				}),
				collectibleId: collectibleId,
				title: title,
				slug: slug,
			});
			await publisher.publishProto({
				routingKey: "telegram.gift.withdraw.user.not.found",
				schema: GiftWithdrawUserNotFoundEventSchema,
				msg: event,
				opts: {
					messageId: uuidv4(),
				},
			});
			logger.error({ userId, messageId }, `❌ User ${userId} not found`);
			throw new Error("User not found");
		}
		logger.info(`🎁 Transferring gift from msg ${messageId} to user ${userId}`);

		try {
			await this.client.invoke(
				new Api.payments.TransferStarGift({
					stargift: new Api.InputSavedStarGiftUser({
						msgId: messageId,
					}),
					toId: peer,
				}),
			);
			logger.info(
				`🎁 Transferred gift from msg ${messageId} to user ${userId} (free)`,
			);
			return;
		} catch (error: unknown) {
			let msg = "";
			if (error instanceof Error) {
				msg = error.message;
			} else if (typeof error === "object" && error && "message" in error) {
				msg = String(error.message);
			} else {
				msg = String(error);
			}

			if (!msg.includes("PAYMENT_REQUIRED")) {
				logger.error(
					{ err: error },
					`❌ Failed to transfer gift (msg ${messageId})`,
				);
				throw error;
			}
		}

		// Иначе — трансфер платный, генерируем инвойс
		const invoice = new Api.InputInvoiceStarGiftTransfer({
			stargift: new Api.InputSavedStarGiftUser({
				msgId: messageId,
			}),
			toId: peer,
		});

		const form = await this.client.invoke(
			new Api.payments.GetPaymentForm({ invoice }),
		);

		await this.client.invoke(
			new Api.payments.SendStarsForm({
				formId: form.formId,
				invoice,
			}),
		);

		logger.info(
			`🎁 Transferred gift from msg ${messageId} to user ${userId} (paid)`,
		);
	}
}
