import { JoinDuelRequest } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_public_service_pb";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { duelClient } from "../client";

export function useJoinDuelMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationKey: ["joinDuel"],
		mutationFn: (request: JoinDuelRequest) => duelClient.joinDuel(request),
		onSuccess: (_, variables) => {
			// Инвалидируем кэш дуэли, чтобы получить обновленные данные
			const duelId = variables.duelId?.value;
			if (duelId) {
				queryClient.invalidateQueries({
					queryKey: ["duel", { duelId }],
				});
			}
			// Инвалидируем список дуэлей
			queryClient.invalidateQueries({
				queryKey: ["duels"],
			});
		},
	});
}
