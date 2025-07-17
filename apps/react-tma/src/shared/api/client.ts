import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { DuelPublicService } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_public_service_pb";
import { EventPublicService } from "@giftduels/protobuf-js/giftduels/event/v1/public_service_pb";
import { GiftPublicService } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_public_service_pb";
import { IdentityPublicService } from "@giftduels/protobuf-js/giftduels/identity/v1/public_service_pb";
import { PaymentPublicService } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import { config } from "@/app/config";
import { authInterceptor } from "./auth-interceptor";

export const transport = createConnectTransport({
	baseUrl: config.apiUrl,

	interceptors: [authInterceptor()],
});

export const giftClient = createClient(GiftPublicService, transport);

export const identityClient = createClient(IdentityPublicService, transport);

export const paymentClient = createClient(PaymentPublicService, transport);

export const duelClient = createClient(DuelPublicService, transport);

export const eventClient = createClient(EventPublicService, transport);
