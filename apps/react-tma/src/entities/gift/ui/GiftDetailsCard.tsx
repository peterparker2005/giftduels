import { create } from "@bufbuild/protobuf";
import {
	GiftAttributeType,
	GiftStatus,
	GiftView,
} from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { GiftWithdrawRequestSchema } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import {
	GiftIdSchema,
	TonAmountSchema,
} from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useMemo } from "react";
import { WithdrawActions } from "@/features/withdraw-gift";
import { usePreviewWithdraw } from "@/shared/api/queries/usePreviewWithdraw";
import { Button } from "@/shared/ui/Button";
import { LottiePlayer } from "@/shared/ui/LottiePlayer";
import { formatFloat } from "@/shared/utils/formatFloat";
import { getFragmentUrl } from "@/shared/utils/getFragmentUrl";

interface GiftDetailsCardProps {
	gift: GiftView;
	onCloseDrawer?: () => void;
}

export const GiftDetailsCard = ({
	gift,
	onCloseDrawer,
}: GiftDetailsCardProps) => {
	const navigate = useNavigate();

	// Preview withdraw logic
	const {
		mutate: previewWithdraw,
		data: previewData,
		isPending: isPreviewPending,
	} = usePreviewWithdraw();

	// Preview withdraw when component mounts
	useEffect(() => {
		if (gift.price?.value && Number(gift.price.value) > 0) {
			const gifts = [
				create(GiftWithdrawRequestSchema, {
					giftId: create(GiftIdSchema, { value: gift.giftId?.value || "" }),
					price: create(TonAmountSchema, { value: gift.price.value }),
				}),
			];
			previewWithdraw(gifts);
		}
	}, [previewWithdraw, gift.giftId?.value, gift.price?.value]);

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
					// src={`https://nft.fragment.com/gift/${gift.slug.toLowerCase()}.lottie.json`}
					src={getFragmentUrl(gift.slug, "lottie")}
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
								<div className="col-span-8 bg-card-muted-accent/50 text-sm bg-card-muted text-foreground w-full flex items-center gap-2 pl-4 h-full">
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
			{gift.status !== GiftStatus.IN_GAME ? (
				<div className="flex flex-col gap-3 mt-auto">
					<WithdrawActions
						giftIds={giftIds}
						previewData={previewData}
						onSuccess={onCloseDrawer}
						disabled={isPreviewPending}
					/>
				</div>
			) : (
				<div className="flex flex-col gap-3 mt-auto mb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
					<Button
						variant="default"
						className="w-full py-3 flex items-center justify-center gap-2 font-semibold"
						onClick={() => {
							if (!gift.relatedDuelId?.value) return;
							navigate({
								to: "/duel/$duelId",
								params: { duelId: gift.relatedDuelId?.value },
							});
						}}
					>
						Open Game
					</Button>
				</div>
			)}
		</div>
	);
};
