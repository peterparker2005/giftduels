import { GiftWithdrawRequest } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import { useMutation } from "@tanstack/react-query";
import { paymentClient } from "../client";

export function usePreviewWithdraw() {
	return useMutation({
		mutationKey: ["previewWithdraw"],
		mutationFn: (gifts: GiftWithdrawRequest[]) => {
			return paymentClient.previewWithdraw({
				gifts,
			});
		},
	});
}
