import { AuthorizeRequestSchema } from "@giftduels/protobuf-js/giftduels/identity/v1/public_service_pb";
import { useMutation } from "@tanstack/react-query";
import { retrieveRawInitData } from "@telegram-apps/sdk";
import { identityClient } from "../client";

export function useAuthorizeMutation() {
	return useMutation({
		mutationKey: ["authorize"],
		mutationFn: () =>
			identityClient.authorize({
				$typeName: AuthorizeRequestSchema.typeName,
				initData: retrieveRawInitData() ?? "",
			}),
	});
}
