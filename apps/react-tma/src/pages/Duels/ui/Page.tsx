import { DuelStatus } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_pb";
import { BiPlus } from "react-icons/bi";
import { GiftCardSmall } from "@/entities/gift/ui/GiftCardSmall";
import { CreateDuelDrawer } from "@/features/create-duel/ui/CreateDuelDrawer";
import { useDuelsQuery } from "@/shared/api/queries/useDuelsQuery";
import { useIntersectionObserver } from "@/shared/hooks/useIntersectionObserver";
import { Button } from "@/shared/ui/Button";
import { Skeleton } from "@/shared/ui/Skeleton";
import { getFragmentUrl } from "@/shared/utils/getFragmentUrl";
import { DuelAvatars } from "./DuelAvatar";

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
		<div className="container">
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
						<div key={idx} className="flex flex-col">
							{page.duels.map((duel) => (
								<div
									className="bg-card rounded-3xl p-2.5 flex flex-col gap-2"
									key={duel.duelId?.value}
								>
									<div className="flex items-center justify-between">
										<DuelAvatars
											participants={duel.participants.map((participant) => ({
												telegramUserId: {
													value:
														participant.telegramUserId?.value.toString() ?? "",
												},
												photoUrl: participant.photoUrl ?? undefined,
											}))}
											maxPlayers={duel.params?.maxPlayers ?? 0}
										/>
										<div className="flex items-center justify-center bg-yellow-400/15 rounded-full px-2 py-1 text-yellow-400 text-sm font-semibold">
											{duel.status === DuelStatus.WAITING_FOR_OPPONENT &&
												"Waiting for opponent"}
											{duel.status === DuelStatus.IN_PROGRESS && "In progress"}
											{duel.status === DuelStatus.COMPLETED && "Completed"}
											{duel.status === DuelStatus.CANCELLED && "Cancelled"}
										</div>
									</div>

									<section className="flex gap-2 overflow-x-auto">
										{duel.stakes.map((stake) => {
											const { gift } = stake;
											return (
												<GiftCardSmall
													key={gift?.giftId?.value}
													gift={{
														giftId: gift?.giftId,
														slug: gift?.slug,
														title: gift?.title,
														price: gift?.price,
													}}
												/>
											);
										})}
									</section>
									<Button
										variant={"primary"}
										className="w-full font-semibold text-base h-12 mt-2.5"
									>
										Join
									</Button>
								</div>
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
						{/* <div className="flex justify-center items-center h-24 animate-pulse bg-card rounded-3xl" />
						<div className="flex justify-center items-center h-24 animate-pulse bg-card rounded-3xl" />
						<div className="flex justify-center items-center h-24 animate-pulse bg-card rounded-3xl" /> */}
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
