import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { ExecuteWithdrawRequest_CommissionCurrency } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_public_service_pb";
import { PreviewWithdrawResponse } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import { useMemo } from "react";
import { Icon } from "@/shared/ui/Icon/Icon";
import { TonWithdrawalCost } from "./TonWithdrawalCost";
import { WithdrawActions } from "./WithdrawActions";
import { WithdrawSummaryCard } from "./WithdrawSummaryCard";

interface WithdrawSummaryProps {
	gifts: GiftView[];
	selectedGiftIds: string[];
	selectedCommissionCurrency: ExecuteWithdrawRequest_CommissionCurrency;
	previewData?: PreviewWithdrawResponse;
	onRemoveGift: (giftId: string) => void;
	onBack: () => void;
	onSuccess?: () => void; // Optional callback for successful withdrawal
	onCommissionCurrencyChange: (
		currency: ExecuteWithdrawRequest_CommissionCurrency,
	) => void;
}

export const WithdrawSummary = ({
	gifts,
	selectedGiftIds,
	previewData,
	onRemoveGift,
	onBack,
	onSuccess,
}: WithdrawSummaryProps) => {
	// Filter selected gifts
	const selectedGifts = useMemo(() => {
		return gifts.filter((gift) =>
			selectedGiftIds.includes(gift.giftId?.value || ""),
		);
	}, [gifts, selectedGiftIds]);

	// Create a map of gift fees for quick lookup
	const giftFeesMap = useMemo(() => {
		const map = new Map<string, { starsFee: number; tonFee: number }>();
		if (previewData?.fees) {
			previewData.fees.forEach((fee) => {
				map.set(fee.giftId?.value || "", {
					starsFee: Number(fee.starsFee?.value || 0),
					tonFee: Number(fee.tonFee?.value || 0),
				});
			});
		}
		return map;
	}, [previewData?.fees]);

	// Handle success with navigation back
	const handleWithdrawSuccess = () => {
		if (onSuccess) {
			onSuccess();
		} else {
			onBack();
		}
	};

	if (selectedGifts.length === 0) {
		return (
			<div className="flex flex-col h-full">
				<div className="flex-1 flex items-center justify-center">
					<p className="text-muted-foreground">No gifts selected</p>
				</div>
			</div>
		);
	}

	return (
		<div className="flex flex-col h-full">
			<div className="flex items-center justify-between gap-2 mb-4">
				<div className="text-muted-foreground flex items-center gap-1">
					<p>Total cost</p>
					<TonWithdrawalCost
						isPending={false} // Preview data is already loaded
						fee={Number(previewData?.totalTonFee?.value || 0)}
					/>
					<p>TON</p>
				</div>
				<div className="text-sm text-muted-foreground">
					{selectedGifts.length} gift{selectedGifts.length !== 1 ? "s" : ""}{" "}
					selected
				</div>
			</div>

			<div className="flex flex-col gap-2 flex-1 overflow-y-auto mb-4">
				{selectedGifts.map((gift) => {
					const giftId = gift.giftId?.value || "";
					const giftFee = giftFeesMap.get(giftId);
					return (
						<WithdrawSummaryCard
							key={giftId}
							gift={gift}
							fee={giftFee?.tonFee || 0}
							onRemove={() => onRemoveGift(giftId)}
						/>
					);
				})}
				<button
					type="button"
					className="text-primary text-left w-max text-base font-medium flex items-center gap-2"
					onClick={onBack}
				>
					<Icon icon="Plus" className="w-4 h-4 shrink-0" />
					<span>Add more gifts</span>
				</button>
			</div>

			<WithdrawActions
				giftIds={selectedGiftIds}
				previewData={previewData}
				onSuccess={handleWithdrawSuccess}
				disabled={selectedGifts.length === 0}
			/>
		</div>
	);
};
