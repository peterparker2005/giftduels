import { useMutation } from "@tanstack/react-query";
import { paymentClient } from "../client";

export function useDepositTonMutation() {
	return useMutation({
		mutationKey: ["deposit-ton"],
		mutationFn: (amount: string) => {
			return paymentClient.depositTon({
				tonAmount: {
					value: amount,
				},
			});
		},
	});
}
