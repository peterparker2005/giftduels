import React from "react";
import { PiCrownSimpleFill } from "react-icons/pi";
import { getAvatarSlots } from "@/entities/duel/utils/getAvatarSlots";
import { cn } from "@/shared/utils/cn";

interface DuelAvatarsProps {
	participants: Array<{
		telegramUserId: { value: string };
		photoUrl?: string;
	}>;
	maxPlayers: number;
	className?: string;
	winnerTelegramUserId?: string;
}

export const DuelAvatars: React.FC<DuelAvatarsProps> = ({
	participants,
	maxPlayers,
	className,
	winnerTelegramUserId,
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
					// biome-ignore lint/suspicious/noArrayIndexKey: TODO: fix this?
					<div key={idx} className="w-8 h-8 relative">
						{isWinner && (
							<PiCrownSimpleFill className="w-4 h-4 rotate-30 absolute -right-1 -top-1.5 text-yellow-400 drop-shadow-md" />
						)}
						{url ? (
							<img
								src={url}
								alt={`Player avatar ${idx + 1}`}
								className="w-full h-full rounded-full object-cover"
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
