import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { PreviewWithdrawResponse } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import { useIntersectionObserver } from "@/shared/hooks/useIntersectionObserver";
import { Button } from "@/shared/ui/Button";
import { useWithdrawForm } from "../hooks/useWithdrawForm";
import { TonWithdrawalCost } from "./TonWithdrawalCost";
import { WithdrawGiftCard } from "./WithdrawGiftCard";
import { WithdrawGiftCardSkeleton } from "./WithdrawGiftCardSkeleton";

interface WithdrawFormProps {
	gifts: GiftView[];
	selectedGifts?: string[];
	onProceedToConfirm: (selectedGifts: string[]) => void;
	onGiftToggle?: (giftId: string) => void;
	onSelectAll?: () => void;
	onClearSelection?: () => void;
	previewData?: PreviewWithdrawResponse;
	isPreviewPending?: boolean;
	isLoadingMore?: boolean;
	onLoadMore?: () => void;
	hasNextPage?: boolean;
}

export const WithdrawForm = ({
	gifts,
	selectedGifts: externalSelectedGifts,
	onProceedToConfirm,
	onGiftToggle,
	onSelectAll,
	onClearSelection,
	previewData,
	isPreviewPending = false,
	isLoadingMore = false,
	onLoadMore,
	hasNextPage = false,
}: WithdrawFormProps) => {
	const form = useWithdrawForm(gifts, externalSelectedGifts);

	// Use external handlers if provided, otherwise use internal form methods
	const handleToggleGift = onGiftToggle || form.toggleGift;
	const handleSelectAll = onSelectAll || form.selectAll;
	const handleClearSelection = onClearSelection || form.clearSelection;

	// Intersection observer for infinite scrolling
	const observerRef = useIntersectionObserver({
		onIntersect: () => {
			if (hasNextPage && !isLoadingMore && onLoadMore) {
				onLoadMore();
			}
		},
		enabled: hasNextPage && !isLoadingMore && !!onLoadMore,
		threshold: 0.1,
	});

	const handleFormSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (form.hasSelection) {
			onProceedToConfirm(form.selectedGifts);
		}
	};

	const handleSelectAllToggle = () => {
		if (form.isAllSelected) {
			handleClearSelection();
		} else {
			handleSelectAll();
		}
	};

	return (
		<form onSubmit={handleFormSubmit} className="flex flex-col h-full">
			<div className="flex items-center justify-between gap-2 mt-0 mb-4">
				<div className="text-muted-foreground flex items-center gap-1">
					<p>Withdrawal cost</p>
					<TonWithdrawalCost
						isPending={isPreviewPending}
						fee={previewData?.totalTonFee?.value}
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

			<div className="flex flex-col gap-2 flex-1 overflow-y-auto min-h-0 mb-4">
				{gifts.map((gift) => {
					const giftId = gift.giftId?.value || "";
					return (
						<WithdrawGiftCard
							key={giftId}
							gift={gift}
							selected={form.isGiftSelected(giftId)}
							onSelectionChange={() => handleToggleGift(giftId)}
						/>
					);
				})}

				{/* Loading skeletons for next page */}
				{isLoadingMore && (
					<>
						<WithdrawGiftCardSkeleton />
						<WithdrawGiftCardSkeleton />
						<WithdrawGiftCardSkeleton />
						<WithdrawGiftCardSkeleton />
					</>
				)}

				{/* Intersection observer trigger */}
				{hasNextPage && <div ref={observerRef} className="h-4" />}
			</div>

			<div className="flex-shrink-0 pb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
				<Button
					type="submit"
					disabled={!form.hasSelection}
					className="w-full py-3"
				>
					Select {form.hasSelection ? `${form.selectedCount} ` : ""}gifts
				</Button>
			</div>
		</form>
	);
};
