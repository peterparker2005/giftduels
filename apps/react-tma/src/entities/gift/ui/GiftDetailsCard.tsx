import {
	GiftAttributeType,
	GiftView,
} from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useEffect, useMemo } from "react";
import { WithdrawActions } from "@/features/withdraw-gift";
import { usePreviewWithdraw } from "@/shared/api/queries/usePreviewWithdraw";
import { LottiePlayer } from "@/shared/ui/LottiePlayer";
import { formatFloat } from "@/shared/utils/formatFloat";

interface GiftDetailsCardProps {
	gift: GiftView;
	onCloseDrawer?: () => void;
}

export const GiftDetailsCard = ({
	gift,
	onCloseDrawer,
}: GiftDetailsCardProps) => {
	// Calculate total TON amount for the gift
	const totalTonAmount = useMemo(() => {
		return gift.price?.value || 0;
	}, [gift.price?.value]);

	// Preview withdraw logic
	const {
		mutate: previewWithdraw,
		data: previewData,
		isPending: isPreviewPending,
	} = usePreviewWithdraw();

	// Preview withdraw when component mounts
	useEffect(() => {
		if (totalTonAmount > 0) {
			previewWithdraw(totalTonAmount);
		}
	}, [previewWithdraw, totalTonAmount]);

	// Get gift ID for WithdrawActions
	const giftIds = useMemo(() => {
		const giftId = gift.giftId?.value;
		return giftId ? [giftId] : [];
	}, [gift.giftId?.value]);

	return (
		<div className="flex flex-col gap-6 h-full">
			{/* Gift Image */}
			<div className="relative w-40 mx-auto h-40 rounded-3xl overflow-hidden">
				<LottiePlayer
					src={`https://nft.fragment.com/gift/${gift.slug.toLowerCase()}.lottie.json`}
					autoplay
					loop
					className="w-full h-full object-cover"
				/>
				<div className="absolute top-0 left-0 w-full h-full bg-gradient-to-b from-transparent from-60% to-black/60" />
				<div className="flex flex-col items-center gap-1 absolute bottom-2 left-0 mx-auto w-full">
					<p className="text-lg text-center leading-[100%] font-semibold">
						{gift.title}
					</p>
					<span className="text-[#cccccc] text-xs font-medium">
						#{gift.collectibleId}
					</span>
				</div>
			</div>

			{/* Gift Details */}

			<div className="flex-1 h-full">
				<div className="flex flex-col rounded-2xl relative overflow-hidden">
					{gift.attributes.map((attribute) => {
						const rarity = attribute.rarityPerMille * 0.1;
						const key = `${attribute.type}-${attribute.name}`;
						return (
							<div key={key} className="grid grid-cols-12 font-medium h-10">
								<div className="col-span-4 text-sm bg-card-muted-accent text-foreground/80 w-full flex items-center justify-center h-full">
									{attribute.type === GiftAttributeType.MODEL && "Model"}
									{attribute.type === GiftAttributeType.SYMBOL && "Symbol"}
									{attribute.type === GiftAttributeType.BACKDROP && "Backdrop"}
								</div>
								<div className="col-span-8 text-sm bg-card-muted text-foreground w-full flex items-center gap-2 ml-4 h-full">
									<p className="text-sm">{attribute.name}</p>
									<div className="bg-primary/10 px-3 text-xs py-0.5 rounded-full text-primary">
										<span>{formatFloat(rarity)}%</span>
									</div>
								</div>
							</div>
						);
					})}
				</div>
			</div>

			{/* Actions */}
			<div className="flex flex-col gap-3 mt-auto">
				<WithdrawActions
					giftIds={giftIds}
					previewData={previewData}
					onSuccess={onCloseDrawer}
					disabled={isPreviewPending}
				/>
			</div>
		</div>
	);
};
