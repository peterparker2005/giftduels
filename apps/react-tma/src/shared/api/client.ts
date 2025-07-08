import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { GiftPublicService } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_public_service_pb";
import { IdentityPublicService } from "@giftduels/protobuf-js/giftduels/identity/v1/public_service_pb";
import { config } from "@/app/config";
import { authInterceptor } from "./auth-interceptor";

export const transport = createConnectTransport({
	baseUrl: config.apiUrl,
	interceptors: [authInterceptor()],
});

export const giftClient = createClient(GiftPublicService, transport);

export const identityClient = createClient(IdentityPublicService, transport);
