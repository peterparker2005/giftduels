import { useMutation } from "@tanstack/react-query";
import { giftClient } from "../client";

export function useExecuteWithdrawMutation() {
	return useMutation({
		mutationFn: (giftIds: string[]) =>
			giftClient.executeWithdraw({
				giftIds: giftIds.map((id) => ({ value: id })),
			}),
	});
}
