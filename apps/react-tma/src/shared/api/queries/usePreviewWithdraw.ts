import { create } from "@bufbuild/protobuf";
import { GiftIdSchema } from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { useMutation } from "@tanstack/react-query";
import { paymentClient } from "../client";

export function usePreviewWithdraw() {
	return useMutation({
		mutationKey: ["previewWithdraw"],
		mutationFn: (giftIds: string[]) =>
			paymentClient.previewWithdraw({
				giftIds: giftIds.map((id) => create(GiftIdSchema, { value: id })),
			}),
	});
}
