import { giftduels } from "@giftduels/protobuf-ts";
import { httpClient } from "@/shared/api/client";

/**
 * Клиент публичных методов identity-сервиса
 */
export class IdentityClient {
	authorize(data: giftduels.identity.v1.AuthorizeRequest) {
		return httpClient
			.post<giftduels.identity.v1.AuthorizeResponse>(
				"/giftduels.identity.v1.IdentityPublicService/Authorize",
				data,
			)
			.then((r) => r.data);
	}

	validateToken(data: giftduels.identity.v1.ValidateTokenRequest) {
		return httpClient
			.post<giftduels.identity.v1.ValidateTokenResponse>(
				"/giftduels.identity.v1.IdentityPublicService/ValidateToken",
				data,
			)
			.then((r) => r.data);
	}

	getProfile() {
		return httpClient
			.post<giftduels.identity.v1.GetProfileResponse>(
				"/giftduels.identity.v1.IdentityPublicService/GetProfile",
				{}, // пустое тело
			)
			.then((r) => r.data);
	}
}

/* единый экспорт */
export const identityClient = new IdentityClient();
