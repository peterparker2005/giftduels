import { useQuery } from "@tanstack/react-query";
import { duelClient } from "../client";

export function useDuelQuery({ duelId }: { duelId: string }) {
	return useQuery({
		queryKey: ["duel", { duelId }],
		queryFn: () => duelClient.getDuel({ id: { value: duelId } }),
	});
}
