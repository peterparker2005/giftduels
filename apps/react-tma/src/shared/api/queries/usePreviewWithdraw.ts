import { GiftWithdrawRequest } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import { useMutation } from "@tanstack/react-query";
import { paymentClient } from "../client";

export function usePreviewWithdraw() {
	return useMutation({
		mutationKey: ["previewWithdraw"],
		mutationFn: (gifts: GiftWithdrawRequest[]) => {
			// Не делать запрос для пустого массива
			if (gifts.length === 0) {
				throw new Error("No gifts to preview");
			}

			return paymentClient.previewWithdraw({
				gifts,
			});
		},
		// Не повторять запрос при ошибке для preview
		retry: false,
		// Не показывать ошибки для preview запросов
		onError: () => {
			// Preview ошибки не критичны, можно игнорировать
		},
	});
}
