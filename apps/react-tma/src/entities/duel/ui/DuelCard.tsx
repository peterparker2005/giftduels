import { timestampDate } from "@bufbuild/protobuf/wkt";
import {
	Duel,
	DuelStatus,
} from "@giftduels/protobuf-js/giftduels/duel/v1/duel_pb";
import { useNavigate } from "@tanstack/react-router";
import { retrieveLaunchParams } from "@telegram-apps/sdk";
import { useMemo } from "react";
import { BiPlus, BiSolidGift } from "react-icons/bi";
import { DuelAvatars } from "@/entities/duel/ui/DuelAvatars";
import { GiftCardSmall } from "@/entities/gift/ui/GiftCardSmall";
import { JoinDuelDrawer } from "@/features/join-duel/ui/JoinDuelDrawer";
import { useCountdown } from "@/shared/hooks/useCountdown";
import { Button } from "@/shared/ui/Button";
import { Icon } from "@/shared/ui/Icon/Icon";
import { cn } from "@/shared/utils/cn";

export function DuelCard({ duel }: { duel: Duel }) {
	const { tgWebAppData } = retrieveLaunchParams();
	const user = tgWebAppData?.user;

	const deadlineDate = useMemo(
		() => duel.nextRollDeadline && timestampDate(duel.nextRollDeadline),
		[duel.nextRollDeadline],
	);

	const { seconds } = useCountdown(deadlineDate);

	const isUserParticipant = useMemo(
		() =>
			duel.participants.find(
				(participant) =>
					participant.telegramUserId?.value.toString() === user?.id.toString(),
			),
		[duel.participants, user],
	);

	// Create a map of participants by telegram user ID for quick lookup
	const navigate = useNavigate();
	const participantsMap = new Map(
		duel.participants.map((participant) => [
			participant.telegramUserId?.value,
			participant,
		]),
	);

	return (
		<div
			className="bg-card rounded-3xl p-2.5 flex flex-col gap-2"
			key={duel.duelId?.value}
		>
			<div className="flex items-center justify-between">
				<DuelAvatars
					winnerTelegramUserId={duel.winnerTelegramUserId?.value.toString()}
					participants={duel.participants.map((participant) => ({
						telegramUserId: {
							value: participant.telegramUserId?.value.toString() ?? "",
						},
						photoUrl: participant.photoUrl ?? undefined,
					}))}
					maxPlayers={duel.params?.maxPlayers ?? 0}
				/>
				<div
					className={cn(
						"flex items-center justify-center rounded-full px-2 py-1 text-xs font-semibold",
						duel.status === DuelStatus.WAITING_FOR_OPPONENT &&
							"bg-yellow-400/15 text-yellow-400",
						duel.status === DuelStatus.IN_PROGRESS &&
							"bg-green-400/15 text-green-400",
						duel.status === DuelStatus.COMPLETED &&
							"bg-green-400/15 text-green-400",
						duel.status === DuelStatus.CANCELLED &&
							"bg-red-400/15 text-red-400",
					)}
				>
					{duel.status === DuelStatus.WAITING_FOR_OPPONENT &&
						"Waiting for opponent"}
					{duel.status === DuelStatus.IN_PROGRESS && "In progress"}
					{duel.status === DuelStatus.COMPLETED && "Completed"}
					{duel.status === DuelStatus.CANCELLED && "Cancelled"}
				</div>
			</div>

			<div className="flex items-center gap-2">
				<div className="text-sm font-semibold px-2 py-1 rounded-full bg-card-accent flex items-center gap-1">
					<BiSolidGift className="w-3.5 h-3.5" />
					<span>{duel.stakes.length}</span>
				</div>
				<div className="text-sm font-semibold px-2 py-1 rounded-full bg-card-accent flex items-center gap-1">
					<Icon icon="TON" className="w-3.5 h-3.5" />
					<span>{duel.totalStakeValue?.value}</span>
				</div>
				{duel.nextRollDeadline && seconds > 0 && (
					<div className="text-sm font-semibold px-2 py-1 rounded-full bg-card-accent flex items-center gap-1">
						<span>{seconds}</span>
					</div>
				)}
			</div>

			<section className="flex gap-2 overflow-x-auto">
				{duel.stakes.map((stake) => {
					const { gift } = stake;
					// Find the participant who made this stake
					const participant = participantsMap.get(
						stake.participantTelegramUserId?.value,
					);

					return (
						<GiftCardSmall
							key={gift?.giftId?.value}
							gift={{
								giftId: gift?.giftId,
								slug: gift?.slug,
								title: gift?.title,
								price: gift?.price,
							}}
							participant={participant}
						/>
					);
				})}
				{duel.status === DuelStatus.WAITING_FOR_OPPONENT && (
					<JoinDuelDrawer
						displayNumber={duel.displayNumber.toString()}
						duel={duel}
					>
						<div className="rounded-3xl bg-card-muted-accent overflow-hidden w-max flex flex-col shrink-0">
							<div className="w-24 h-24 rounded-3xl border-2 border-dashed border-card-accent bg-card flex items-center justify-center">
								<BiPlus className="w-8 h-8 text-card-accent" />
							</div>
							<div className="text-center text-xs py-1" />
						</div>
					</JoinDuelDrawer>
				)}
			</section>
			{duel.status === DuelStatus.WAITING_FOR_OPPONENT &&
				!isUserParticipant && (
					<JoinDuelDrawer
						displayNumber={duel.displayNumber.toString()}
						duel={duel}
					>
						<Button
							variant={"primary"}
							className="w-full font-semibold text-base h-12 mt-2.5"
						>
							Join
						</Button>
					</JoinDuelDrawer>
				)}
			{(duel.status === DuelStatus.IN_PROGRESS || isUserParticipant) && (
				<Button
					variant={"secondary"}
					className="w-full font-semibold text-base h-12 mt-2.5 bg-card-accent"
					onClick={() => {
						navigate({
							to: "/duel/$duelId",
							params: { duelId: duel.duelId?.value ?? "" },
						});
					}}
				>
					Watch
				</Button>
			)}
		</div>
	);
}
