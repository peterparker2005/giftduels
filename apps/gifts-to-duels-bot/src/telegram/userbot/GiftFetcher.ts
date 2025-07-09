import { Api, TelegramClient } from "telegram";

export class GiftFetcher {
	constructor(private client: TelegramClient) {}

	async getUserGifts(peer: Api.TypeInputPeer, limit: number) {
		const pageSize = Math.min(limit, 100);
		let offset = "";
		const gifts: Api.SavedStarGift[] = [];
		const usersMap = new Map<number, Api.TypeUser>();

		do {
			const res = await this.client.invoke(
				new Api.payments.GetSavedStarGifts({ peer, offset, limit: pageSize }),
			);
			res.gifts.forEach((g) => gifts.push(g));
			res.users.forEach((u) => usersMap.set(+u.id, u));
			offset = res.nextOffset ?? "";
		} while (offset && gifts.length < limit);

		return { gifts, users: [...usersMap.values()], total: gifts.length };
	}
}
