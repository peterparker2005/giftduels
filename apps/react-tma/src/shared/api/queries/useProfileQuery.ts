import { useQuery } from "@tanstack/react-query";
import { identityClient } from "../client";

export function useProfileQuery() {
	return useQuery({
		queryKey: ["profile"],
		queryFn: () => identityClient.getProfile({}),
	});
}
