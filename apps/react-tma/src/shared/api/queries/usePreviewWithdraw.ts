import { create } from "@bufbuild/protobuf";
import { TonAmountSchema } from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { useMutation } from "@tanstack/react-query";
import { paymentClient } from "../client";

export function usePreviewWithdraw() {
	return useMutation({
		mutationKey: ["previewWithdraw"],
		mutationFn: (amount: number) =>
			paymentClient.previewWithdraw({
				tonAmount: create(TonAmountSchema, { value: amount }),
			}),
	});
}
