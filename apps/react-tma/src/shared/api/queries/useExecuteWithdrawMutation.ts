import { ExecuteWithdrawRequest_CommissionCurrency } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_public_service_pb";
import { useMutation } from "@tanstack/react-query";
import { giftClient } from "../client";

export function useExecuteWithdrawMutation() {
	return useMutation({
		mutationFn: ({
			giftIds,
			commissionCurrency,
		}: {
			giftIds: string[];
			commissionCurrency: ExecuteWithdrawRequest_CommissionCurrency;
		}) =>
			giftClient.executeWithdraw({
				giftIds: giftIds.map((id) => ({ value: id })),
				commissionCurrency,
			}),
	});
}
