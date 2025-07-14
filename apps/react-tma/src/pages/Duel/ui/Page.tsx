import { useParams } from "@tanstack/react-router";
import { useDuelQuery } from "@/shared/api/queries/useDuelQuery";

export function Page() {
	const { duelId } = useParams({ from: "/default-layout/duel/$duelId" });
	const { data, isLoading } = useDuelQuery({ duelId });

	if (isLoading) return <div>Loading...</div>;

	return (
		<div>
			<div>{duelId}</div>
		</div>
	);
}

export default Page;
