import { GiftCard } from "@/entities/gift/ui/GiftCard";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { EmptyInventory } from "./EmptyInventory";
import { InventoryLoading } from "./InventoryLoading";

export const Inventory = () => {
	const { data, isLoading } = useGiftsQuery();

	if (isLoading) return <InventoryLoading />;
	if (!data || data.gifts.length < 1) return <EmptyInventory />;
	return (
		<section className="grid grid-cols-2 gap-4 pb-20">
			{data.gifts.map((gift) => (
				<GiftCard key={gift.giftId?.value} gift={gift} />
			))}
		</section>
	);
};
