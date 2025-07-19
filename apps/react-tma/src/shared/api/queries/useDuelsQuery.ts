import { GetDuelListFilter_FilterType } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_public_service_pb";
import { useInfiniteQuery } from "@tanstack/react-query";
import { duelClient } from "../client";

export function useDuelsQuery(filterType: GetDuelListFilter_FilterType) {
	return useInfiniteQuery({
		queryKey: ["duels"],
		queryFn: () =>
			duelClient.getDuelList({
				pageRequest: {
					page: 1,
					pageSize: 10,
				},
				filter: {
					filterType,
				},
			}),
		enabled: !!filterType,
		getNextPageParam: (lastPage, pages) => {
			// Проверяем, есть ли еще страницы согласно пагинации
			const currentPage = pages.length;
			const totalPages = lastPage.pagination?.totalPages || 1;

			// Возвращаем следующую страницу только если текущая страница меньше общего количества
			return currentPage < totalPages ? currentPage + 1 : undefined;
		},
		initialPageParam: 1,
	});
}
