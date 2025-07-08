import bigInt from "big-integer";
import { Api, TelegramClient } from "telegram";
import { StringSession } from "telegram/sessions";
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

	async sendGift(params: {
		userId: number;
		giftId: string | number | bigint;
		text?: string;
		hideName?: boolean;
		includeUpgrade?: boolean;
	}) {
		const peer = await this.client.getInputEntity(params.userId);

		const invoice = new Api.InputInvoiceStarGift({
			peer,
			giftId: bigInt(Number(params.giftId)),
			hideName: params.hideName ?? false,
			includeUpgrade: params.includeUpgrade ?? false,
			message: new Api.TextWithEntities({
				text: params.text ?? "",
				entities: [],
			}),
		});

		const form = await this.client.invoke(
			new Api.payments.GetPaymentForm({ invoice }),
		);

		await this.client.invoke(
			new Api.payments.SendStarsForm({ formId: form.formId, invoice }),
		);

		logger.info(`🎁 Sent gift to ${params.userId}`);
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

	async transferGift(params: { userId: number; messageId: number }) {
		const { userId, messageId } = params;
		const peer = await this.client.getInputEntity(userId);
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
