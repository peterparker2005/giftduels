import { GiftCard } from "@/entities/gift/ui/GiftCard";
import { GiftCardSkeleton } from "@/entities/gift/ui/GiftCardSkeleton";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { useIntersectionObserver } from "@/shared/hooks/useIntersectionObserver";
import { EmptyInventory } from "./EmptyInventory";
import { InventoryLoading } from "./InventoryLoading";

export const Inventory = () => {
	const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
		useGiftsQuery();

	// Flatten all pages into a single array of gifts
	const allGifts = data?.pages.flatMap((page) => page.gifts) || [];

	// Intersection observer for infinite scrolling
	const observerRef = useIntersectionObserver({
		onIntersect: () => {
			if (hasNextPage && !isFetchingNextPage) {
				fetchNextPage();
			}
		},
		enabled: hasNextPage && !isFetchingNextPage,
		threshold: 0.1,
	});

	if (isLoading) return <InventoryLoading />;
	if (!data || allGifts.length < 1) return <EmptyInventory />;

	return (
		<section className="pb-20">
			<div className="grid grid-cols-2 gap-4">
				{allGifts.map((gift) => (
					<GiftCard key={gift.giftId?.value} gift={gift} />
				))}

				{/* Loading skeletons for next page */}
				{isFetchingNextPage && (
					<>
						<GiftCardSkeleton />
						<GiftCardSkeleton />
						<GiftCardSkeleton />
						<GiftCardSkeleton />
					</>
				)}
			</div>

			{/* Intersection observer trigger */}
			<div ref={observerRef} className="h-4" />
		</section>
	);
};
