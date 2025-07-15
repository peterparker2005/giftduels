import { Duel } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_pb";
import {
	GiftStatus,
	GiftView,
} from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useMemo } from "react";
import { SelectGiftCard } from "@/entities/gift/ui/SelectGiftCard";
import { SelectGiftCardSkeleton } from "@/entities/gift/ui/SelectGiftCardSkeleton";
import { useIntersectionObserver } from "@/shared/hooks/useIntersectionObserver";
import { Button } from "@/shared/ui/Button";
import { cn } from "@/shared/utils/cn";
import { useJoinDuelForm } from "../hooks/useJoinDuelForm";

interface SelectGiftsForJoinFormProps {
	duel: Duel;
	gifts: GiftView[];
	selectedGifts?: string[];
	onJoinDuel: (selectedGifts: string[]) => void;
	onGiftToggle?: (giftId: string) => void;
	onSelectAll?: () => void;
	onClearSelection?: () => void;
	isLoadingMore?: boolean;
	onLoadMore?: () => void;
	hasNextPage?: boolean;
	isPending?: boolean;
}

export function SelectGiftsForJoinForm({
	duel,
	gifts,
	selectedGifts: externalSelectedGifts,
	onJoinDuel,
	onGiftToggle,
	onSelectAll,
	onClearSelection,
	isLoadingMore = false,
	onLoadMore,
	hasNextPage = false,
	isPending = false,
}: SelectGiftsForJoinFormProps) {
	const form = useJoinDuelForm(gifts, externalSelectedGifts);

	// Use external handlers if provided, otherwise use internal form methods
	const handleToggleGift = onGiftToggle || form.toggleGift;
	const handleSelectAll = onSelectAll || form.selectAll;
	const handleClearSelection = onClearSelection || form.clearSelection;

	// Calculate total value of selected gifts
	const totalStakeValue = useMemo(() => {
		const selectedGiftObjects = gifts.filter((gift) =>
			form.selectedGifts.includes(gift.giftId?.value || ""),
		);

		return selectedGiftObjects.reduce((total, gift) => {
			const price = parseFloat(gift.price?.value || "0");
			return total + price;
		}, 0);
	}, [gifts, form.selectedGifts]);

	// Get entry price range from duel
	const entryPriceRange = useMemo(() => {
		const minPrice = parseFloat(
			duel.entryPriceRange?.minEntryPrice?.value || "0",
		);
		const maxPrice = parseFloat(
			duel.entryPriceRange?.maxEntryPrice?.value || "0",
		);
		return { minPrice, maxPrice };
	}, [duel.entryPriceRange]);

	// Check if total stake value is within the allowed range
	const isStakeInRange = useMemo(() => {
		if (entryPriceRange.minPrice === 0 && entryPriceRange.maxPrice === 0) {
			return true; // No range specified, allow any value
		}
		return (
			totalStakeValue >= entryPriceRange.minPrice &&
			totalStakeValue <= entryPriceRange.maxPrice
		);
	}, [totalStakeValue, entryPriceRange]);

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
		if (form.hasSelection && isStakeInRange) {
			onJoinDuel(form.selectedGifts);
		}
	};

	const handleSelectAllToggle = () => {
		if (form.isAllSelected) {
			handleClearSelection();
		} else {
			handleSelectAll();
		}
	};

	const availableGifts = useMemo(() => {
		return gifts.filter((gift) => gift.status === GiftStatus.OWNED);
	}, [gifts]);

	return (
		<form onSubmit={handleFormSubmit} className="flex flex-col h-full">
			<div className="flex items-center justify-between gap-2 mt-0 mb-4">
				<div
					className={cn(
						"text-muted-foreground flex items-center gap-1",
						!isStakeInRange && "text-red-500",
					)}
				>
					<p>{totalStakeValue} TON</p>
					{entryPriceRange.minPrice > 0 && entryPriceRange.maxPrice > 0 && (
						<span className="text-xs text-muted-foreground">
							({entryPriceRange.minPrice} - {entryPriceRange.maxPrice} TON
							required)
						</span>
					)}
				</div>
				<button
					type="button"
					onClick={handleSelectAllToggle}
					className="text-primary font-semibold hover:text-primary/80 transition-colors"
				>
					{form.isAllSelected ? "Deselect all" : "Select all"}
				</button>
			</div>

			{/* Show validation error if stake is out of range */}
			{form.hasSelection && !isStakeInRange && entryPriceRange.minPrice > 0 && (
				<div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
					<p className="text-destructive text-sm">
						Total stake value ({totalStakeValue} TON) must be between{" "}
						{entryPriceRange.minPrice} and {entryPriceRange.maxPrice} TON
					</p>
				</div>
			)}

			<div className="flex flex-col gap-2 flex-1 overflow-y-auto min-h-0 mb-4">
				{availableGifts.map((gift) => {
					const giftId = gift.giftId?.value || "";
					return (
						<SelectGiftCard
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
						<SelectGiftCardSkeleton />
						<SelectGiftCardSkeleton />
						<SelectGiftCardSkeleton />
						<SelectGiftCardSkeleton />
					</>
				)}

				{/* Intersection observer trigger */}
				{hasNextPage && <div ref={observerRef} className="h-4" />}
			</div>

			<div className="shrink-0 pb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
				<Button
					type="submit"
					disabled={!form.hasSelection || !isStakeInRange || isPending}
					className="w-full py-3"
				>
					{isPending
						? "Joining..."
						: `Join with ${form.hasSelection ? `${form.selectedCount} ` : ""}gifts`}
				</Button>
			</div>
		</form>
	);
}
