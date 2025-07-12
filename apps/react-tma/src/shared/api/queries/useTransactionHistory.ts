import { useInfiniteQuery } from "@tanstack/react-query";
import { paymentClient } from "../client";

export function useTransactionHistory() {
	return useInfiniteQuery({
		queryKey: ["transaction-history"],
		queryFn: ({ pageParam = 1 }) =>
			paymentClient.getTransactionHistory({
				pagination: {
					page: pageParam,
					pageSize: 10,
				},
			}),
		getNextPageParam: (lastPage, pages) =>
			lastPage.transactions.length > 0 ? pages.length + 1 : undefined,
		initialPageParam: 1,
	});
}
