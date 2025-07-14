import { CreateDuelRequest } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_public_service_pb";
import { useMutation } from "@tanstack/react-query";
import { duelClient } from "../client";

export function useCreateDuelMutation() {
	return useMutation({
		mutationKey: ["createDuel"],
		mutationFn: (duel: CreateDuelRequest) => duelClient.createDuel(duel),
	});
}
