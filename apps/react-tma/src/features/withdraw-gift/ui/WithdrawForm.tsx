import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useEffect, useMemo } from "react";
import { usePreviewWithdraw } from "@/shared/api/queries/usePreviewWithdraw";
import { Button } from "@/shared/ui/Button";
import { useWithdrawForm } from "../hooks/useWithdrawForm";
import { TonWithdrawalCost } from "./TonWithdrawalCost";
import { WithdrawGiftCard } from "./WithdrawGiftCard";

interface WithdrawFormProps {
	gifts: GiftView[];
}

export const WithdrawForm = ({ gifts }: WithdrawFormProps) => {
	const form = useWithdrawForm(gifts);
	const {
		mutate: previewWithdraw,
		data: previewWithdrawData,
		isPending,
	} = usePreviewWithdraw();

	const handleFormSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (form.hasSelection) {
			form.handleSubmit();
		}
	};

	const handleSelectAllToggle = () => {
		if (form.isAllSelected) {
			form.clearSelection();
		} else {
			form.selectAll();
		}
	};

	// Calculate total TON amount for selected gifts
	const totalTonAmount = useMemo(() => {
		return form.selectedGifts.reduce((total, giftId) => {
			const gift = gifts.find((g) => g.giftId?.value === giftId);
			return total + (gift?.price?.value || 0);
		}, 0);
	}, [form.selectedGifts, gifts]);

	useEffect(() => {
		if (form.hasSelection) {
			previewWithdraw(totalTonAmount);
		}
	}, [form.hasSelection, previewWithdraw, totalTonAmount]);

	return (
		<form onSubmit={handleFormSubmit} className="flex flex-col h-full">
			<div className="flex items-center justify-between gap-2 mt-4 mb-4">
				<div className="text-muted-foreground flex items-center gap-1">
					<p>Withdrawal cost</p>
					<TonWithdrawalCost
						isPending={isPending}
						fee={previewWithdrawData?.totalTonFee?.value}
					/>
					<p>TON</p>
				</div>
				<button
					type="button"
					onClick={handleSelectAllToggle}
					className="text-primary font-semibold hover:text-primary/80 transition-colors"
				>
					{form.isAllSelected ? "Deselect all" : "Select all"}
				</button>
			</div>

			<div className="flex flex-col gap-2 flex-1 overflow-y-auto mb-4">
				{gifts.map((gift) => {
					const giftId = gift.giftId?.value || "";
					return (
						<WithdrawGiftCard
							key={giftId}
							gift={gift}
							selected={form.isGiftSelected(giftId)}
							onSelectionChange={() => form.toggleGift(giftId)}
						/>
					);
				})}
			</div>

			<div className="mb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
				<Button
					type="submit"
					disabled={!form.hasSelection}
					className="w-full py-3"
				>
					Withdraw {form.hasSelection ? `${form.selectedCount} ` : ""}gifts
				</Button>
			</div>
		</form>
	);
};
