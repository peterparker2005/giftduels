import { getAvatarSlots } from "@/entities/duel/utils/getAvatarSlots"
import { cn } from "@/shared/utils/cn"
import { DuelStatus } from '@giftduels/protobuf-js/giftduels/duel/v1/duel_pb'
import { openLink } from "@telegram-apps/sdk"
import React from "react"
import { PiCrownSimpleFill } from "react-icons/pi"

interface DuelAvatarsProps {
	participants: Array<{
		telegramUserId: { value: string };
		photoUrl?: string;
	}>;
	maxPlayers: number;
	className?: string;
	winnerTelegramUserId?: string;
	status: DuelStatus
}

export const DuelAvatars: React.FC<DuelAvatarsProps> = ({
	participants,
	maxPlayers,
	className,
	winnerTelegramUserId,
	status,
}) => {
	const slots = getAvatarSlots(participants, maxPlayers);

	return (
		<div className={cn("flex items-center gap-2", className)}>
			{slots.map((url, idx) => {
				const participant = participants[idx];
				const isWinner =
					Boolean(winnerTelegramUserId) &&
					participant != null &&
					participant.telegramUserId.value === winnerTelegramUserId;
				return (
					<div
						// biome-ignore lint/suspicious/noArrayIndexKey: TODO: fix this?
						key={idx}
						role="button"
						tabIndex={0}
						aria-label={`user profile`}
						className="w-8 h-8 relative cursor-pointer"
						onClick={() => {
							openLink(
								`tg://user?id=${participant.telegramUserId.value}`,
							);
						}}
					>
						{isWinner && (
							<PiCrownSimpleFill className="w-4 h-4 rotate-30 absolute -right-1 -top-1.5 text-yellow-400 drop-shadow-md" />
						)}
						{url ? (
							<img
								src={url}
								alt={`Player avatar ${idx + 1}`}
								className={cn(
									"w-full h-full rounded-full object-cover",
									!isWinner && status === DuelStatus.COMPLETED && "opacity-50",
								)}
							/>
						) : (
							<div className="w-full h-full rounded-full bg-card-muted-accent" />
						)}
					</div>
				);
			})}
		</div>
	);
};
