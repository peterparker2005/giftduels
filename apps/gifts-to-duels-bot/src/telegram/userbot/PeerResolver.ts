import { Api, TelegramClient } from "telegram";

type PeerResolverFn = (
	user: string | number,
) => Promise<Api.TypeInputPeer | undefined>;

export class PeerResolver {
	constructor(private client: TelegramClient) {}

	private resolvers: PeerResolverFn[] = [
		this.tryGetInputEntity.bind(this),
		this.tryGetEntity.bind(this),
		this.tryDialogs.bind(this),
	];

	/**
	 * Попытаться последовательно промапить вход (username или userId)
	 * в TypeInputPeer.
	 * Бросает ошибку, если ни один из способов не сработал.
	 */
	async resolve(user: string | number): Promise<Api.TypeInputPeer> {
		for (const fn of this.resolvers) {
			const peer = await fn(user);
			if (peer) {
				return peer;
			}
		}
		throw new Error(`Не удалось разрешить peer для ${user}`);
	}

	/**
	 * 1. Прямой вызов client.getInputEntity(user)
	 */
	private async tryGetInputEntity(
		user: string | number,
	): Promise<Api.TypeInputPeer | undefined> {
		try {
			const inputPeer = await this.client.getInputEntity(user);
			if (
				inputPeer instanceof Api.InputPeerUser ||
				inputPeer instanceof Api.InputPeerChannel
			) {
				return inputPeer;
			}
		} catch {
			// игнорируем, попробуем следующий способ
		}
		return undefined;
	}

	/**
	 * 2. Получить полную entity, затем из неё inputPeer
	 */
	private async tryGetEntity(
		user: string | number,
	): Promise<Api.TypeInputPeer | undefined> {
		try {
			const entity = await this.client.getEntity(user);
			const inputPeer = await this.client.getInputEntity(entity);
			if (
				inputPeer instanceof Api.InputPeerUser ||
				inputPeer instanceof Api.InputPeerChannel
			) {
				return inputPeer;
			}
		} catch {
			// игнорируем
		}
		return undefined;
	}

	/**
	 * 3. Перебрать последние диалоги и найти там пользователя
	 */
	private async tryDialogs(
		user: string | number,
	): Promise<Api.TypeInputPeer | undefined> {
		try {
			const dialogs = await this.client.getDialogs({ limit: 100 });
			for (const dialog of dialogs) {
				const entity = dialog.entity;
				if (entity instanceof Api.User && Number(entity.id) === Number(user)) {
					const inputPeer = await this.client.getInputEntity(entity);
					if (
						inputPeer instanceof Api.InputPeerUser ||
						inputPeer instanceof Api.InputPeerChannel
					) {
						return inputPeer;
					}
				}
			}
		} catch {
			// игнорируем
		}
		return undefined;
	}
}
