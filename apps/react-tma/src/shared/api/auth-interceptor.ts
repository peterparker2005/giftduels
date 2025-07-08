import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";

import { AuthorizeRequestSchema } from "@giftduels/protobuf-js/giftduels/identity/v1/public_service_pb";
import { retrieveRawInitData } from "@telegram-apps/sdk";
import { useAuthStore } from "@/features/auth/model/store";
import { identityClient } from "./client";

class AuthManager {
	private refreshing?: Promise<string>;

	get token(): string | null {
		return useAuthStore.getState().token;
	}
	set token(v: string) {
		useAuthStore.getState().setToken(v);
	}

	async refresh(): Promise<string> {
		if (this.refreshing) return this.refreshing;

		this.refreshing = (async () => {
			const initData = retrieveRawInitData();
			if (!initData) throw new Error("Telegram initData missing");

			const { token } = await identityClient.authorize({
				$typeName: AuthorizeRequestSchema.typeName,
				initData,
			});

			this.token = token; // saved
			this.refreshing = undefined; // reset flag
			return token;
		})();

		return this.refreshing;
	}
}
const auth = new AuthManager();

export function authInterceptor(): Interceptor {
	return (next) => async (req) => {
		if (auth.token) {
			req.header.set("Authorization", `Bearer ${auth.token}`);
		}

		try {
			return await next(req);
		} catch (e) {
			const err = ConnectError.from(e);

			if (err.code === Code.Unauthenticated) {
				const newToken = await auth.refresh();
				req.header.set("Authorization", `Bearer ${newToken}`);
				return await next(req);
			}
			throw err;
		}
	};
}
