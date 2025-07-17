import { useMutation } from "@tanstack/react-query";
import { eventClient } from "../client";

export function useSubscribeDuelsMutation() {
	return useMutation({
		mutationFn: async (duelIds: string[]) =>
			eventClient.subscribeDuels({
				duelIds: duelIds.map((id) => ({ value: id })),
			}),
	});
}
