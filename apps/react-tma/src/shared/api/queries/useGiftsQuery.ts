import { useInfiniteQuery } from "@tanstack/react-query";
import { giftClient } from "../client";

export function useGiftsQuery() {
	return useInfiniteQuery({
		queryKey: ["gifts"],
		queryFn: ({ pageParam = 1 }) =>
			giftClient.getGifts({ pagination: { page: pageParam, pageSize: 10 } }),
		getNextPageParam: (lastPage, pages) =>
			lastPage.gifts.length > 0 ? pages.length + 1 : undefined,
		initialPageParam: 1,
	});
}
