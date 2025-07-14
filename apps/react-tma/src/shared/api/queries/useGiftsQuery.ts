import { useInfiniteQuery } from "@tanstack/react-query";
import { giftClient } from "../client";

export function useGiftsQuery() {
	return useInfiniteQuery({
		queryKey: ["gifts"],
		queryFn: ({ pageParam = 1 }) =>
			giftClient.getGifts({ pagination: { page: pageParam, pageSize: 10 } }),
		getNextPageParam: (lastPage, pages) => {
			// Проверяем, есть ли еще страницы согласно пагинации
			const currentPage = pages.length;
			const totalPages = lastPage.pagination?.totalPages || 1;

			// Возвращаем следующую страницу только если текущая страница меньше общего количества
			return currentPage < totalPages ? currentPage + 1 : undefined;
		},
		initialPageParam: 1,
		staleTime: 5 * 60 * 1000,
	});
}
