import {
	GiftId,
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
}

export function GiftCardSmall({ gift }: GiftCardSmallProps) {
	return (
		<div
			key={gift?.giftId?.value}
			className="rounded-3xl bg-card-muted-accent overflow-hidden w-max flex flex-col shrink-0"
		>
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
