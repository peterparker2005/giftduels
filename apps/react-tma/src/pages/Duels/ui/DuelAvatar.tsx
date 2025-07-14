import React from "react";
import { getAvatarSlots } from "@/entities/duel/utils/getAvatarSlots";
import { cn } from "@/shared/utils/cn";

interface DuelAvatarsProps {
	participants: Array<{
		telegramUserId: { value: string };
		photoUrl?: string;
	}>;
	maxPlayers: number;
	className?: string;
}

export const DuelAvatars: React.FC<DuelAvatarsProps> = ({
	participants,
	maxPlayers,
	className,
}) => {
	const slots = getAvatarSlots(participants, maxPlayers);

	return (
		<div className={cn("flex items-center gap-2", className)}>
			{slots.map((url, idx) => (
				// biome-ignore lint/suspicious/noArrayIndexKey: TODO: fix this?
				<div key={idx} className="w-10 h-10">
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
			))}
		</div>
	);
};
