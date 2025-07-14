import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useMemo } from "react";
import { SelectGiftCard } from "@/entities/gift/ui/SelectGiftCard";
import { useIntersectionObserver } from "@/shared/hooks/useIntersectionObserver";
import { Button } from "@/shared/ui/Button";

interface SelectGiftsFormProps {
	gifts: GiftView[];
	selectedGifts: string[];
	onGiftToggle: (giftId: string) => void;
	onSelectAll: () => void;
	onClearSelection: () => void;
	onConfirm: () => void;
	onBack: () => void;
	isLoadingMore?: boolean;
	onLoadMore?: () => void;
	hasNextPage?: boolean;
}

export function SelectGiftsForm({
	gifts,
	selectedGifts,
	onGiftToggle,
	onSelectAll,
	onClearSelection,
	onConfirm,
	onBack,
	isLoadingMore = false,
	onLoadMore,
	hasNextPage = false,
}: SelectGiftsFormProps) {
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

	const isAllSelected =
		gifts.length > 0 && selectedGifts.length === gifts.length;
	const hasSelection = selectedGifts.length > 0;

	const handleSelectAllToggle = () => {
		if (isAllSelected) {
			onClearSelection();
		} else {
			onSelectAll();
		}
	};

	const totalValue = useMemo(() => {
		return selectedGifts.reduce((acc, giftId) => {
			const gift = gifts.find((g) => g.giftId?.value === giftId);
			return acc + Number(gift?.price?.value || 0);
		}, 0);
	}, [selectedGifts, gifts]);

	return (
		<div className="flex flex-col h-full">
			{/* Header */}
			<div className="flex items-center justify-between gap-2 mb-4">
				<div className="text-muted-foreground font-medium text-base">
					{totalValue} TON
				</div>
				<button
					type="button"
					onClick={handleSelectAllToggle}
					className="text-primary font-semibold hover:text-primary/80 transition-colors"
				>
					{isAllSelected ? "Deselect all" : "Select all"}
				</button>
			</div>

			{/* Gifts List */}
			<div className="flex flex-col gap-2 flex-1 overflow-y-auto min-h-0 mb-4">
				{gifts.map((gift) => {
					const giftId = gift.giftId?.value || "";
					const isSelected = selectedGifts.includes(giftId);

					return (
						<SelectGiftCard
							key={giftId}
							gift={gift}
							selected={isSelected}
							onSelectionChange={() => onGiftToggle(giftId)}
						/>
					);
				})}

				{hasNextPage && <div ref={observerRef} className="h-4" />}
			</div>

			{/* Action Buttons */}
			<div className="flex-shrink-0 pb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
				<Button
					onClick={onConfirm}
					disabled={!hasSelection}
					className="w-full py-3"
				>
					Select {hasSelection ? `${selectedGifts.length} ` : ""}gifts
				</Button>
			</div>
		</div>
	);
}
