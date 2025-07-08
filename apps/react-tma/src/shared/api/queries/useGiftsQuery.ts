import { useQuery } from "@tanstack/react-query";
import { giftClient } from "../client";

export function useGiftsQuery() {
	return useQuery({
		queryKey: ["gifts"],
		queryFn: () =>
			giftClient.getGifts({ pagination: { page: 1, pageSize: 10 } }),
	});
}
