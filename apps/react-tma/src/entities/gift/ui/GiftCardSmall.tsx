import {
	GiftId,
	TelegramUserId,
	TonAmount,
} from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { getFragmentUrl } from "@/shared/utils/getFragmentUrl";

interface GiftCardSmallProps {
	gift: {
		giftId?: GiftId;
		slug?: string;
		title?: string;
		price?: TonAmount;
	};
	participant?: {
		telegramUserId?: TelegramUserId;
		photoUrl?: string;
		isCreator?: boolean;
	};
}

export function GiftCardSmall({ gift, participant }: GiftCardSmallProps) {
	return (
		<div
			key={gift?.giftId?.value}
			className="rounded-3xl bg-card-muted-accent w-max flex flex-col shrink-0 relative"
		>
			{/* Participant Avatar Overlay */}
			{participant?.photoUrl && (
				<div className="absolute top-2 left-2 z-10">
					<div className="relative">
						<img
							src={participant.photoUrl}
							alt="Participant"
							className="w-6 h-6 rounded-full"
						/>
					</div>
				</div>
			)}

			<img
				src={getFragmentUrl(gift?.slug || "")}
				alt={gift?.title || ""}
				className="h-24 w-24 rounded-3xl"
			/>
			<div className="text-center text-xs py-1">
				<span className="font-medium">{gift?.price?.value} TON</span>
			</div>
		</div>
	);
}
