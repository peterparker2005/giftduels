import { BiPlus } from "react-icons/bi";
import { DuelCard } from "@/entities/duel/ui/DuelCard";
import { CreateDuelDrawer } from "@/features/create-duel/ui/CreateDuelDrawer";
import { useDuelsQuery } from "@/shared/api/queries/useDuelsQuery";
import { useIntersectionObserver } from "@/shared/hooks/useIntersectionObserver";
import { Button } from "@/shared/ui/Button";
import { Skeleton } from "@/shared/ui/Skeleton";

export function Page() {
	const { data, fetchNextPage, hasNextPage, isFetching } = useDuelsQuery();

	const intersectionRef = useIntersectionObserver({
		onIntersect: () => {
			if (hasNextPage) {
				fetchNextPage();
			}
		},
		enabled: hasNextPage,
		threshold: 1,
	});

	return (
		<div className="container pb-20">
			<CreateDuelDrawer>
				<Button
					variant={"primary"}
					className="font-semibold w-full text-base h-12 flex items-center justify-center gap-1"
				>
					<BiPlus className="w-5 h-5" />
					<span>Create</span>
				</Button>
			</CreateDuelDrawer>

			<div className="flex flex-col gap-4 mt-4">
				{data?.pages.map((page, idx) => {
					return (
						// biome-ignore lint/suspicious/noArrayIndexKey: page
						<div key={idx} className="flex flex-col gap-4">
							{page.duels.map((duel) => (
								<DuelCard duel={duel} key={duel.duelId?.value} />
							))}
							<div ref={intersectionRef} />
						</div>
					);
				})}
				{isFetching && (
					<>
						<Skeleton className="h-24" />
						<Skeleton className="h-24" />
						<Skeleton className="h-24" />
					</>
				)}
			</div>
		</div>
	);
}

export default Page;
export function getAvatarSlots(
	participants: Array<{ photoUrl?: string }>,
	maxPlayers: number,
): Array<string | null> {
	const slots: Array<string | null> = Array(maxPlayers).fill(null);

	participants.forEach((p, idx) => {
		if (idx < maxPlayers) {
			slots[idx] = p.photoUrl ?? null;
		}
	});

	return slots;
}
