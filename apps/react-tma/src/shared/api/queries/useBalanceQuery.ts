import { useQuery } from "@tanstack/react-query";
import { paymentClient } from "../client";

export function useBalanceQuery() {
	return useQuery({
		queryKey: ["balance"],
		queryFn: () => paymentClient.getBalance({}),
	});
}
