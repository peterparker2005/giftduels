import { IoIosGift } from "react-icons/io";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { Icon } from "@/shared/ui/Icon/Icon";

export const InventoryStats = () => {
	const { data } = useGiftsQuery();

	const totalGifts = data?.pages.flatMap((page) => page.gifts).length;
	const totalValue = data?.pages
		.flatMap((page) => page.gifts)
		.reduce((acc, gift) => acc + (gift.price?.value || 0), 0);

	return (
		<section className="grid grid-cols-2 place-items-center">
			<div className="flex flex-col items-center gap-0">
				<div className="flex items-center gap-1.5">
					<span className="font-medium text-lg">{totalGifts}</span>
					<IoIosGift className="w-4 h-4 shrink-0" />
				</div>
				<span className="text-muted-foreground">All gifts</span>
			</div>
			<div className="flex flex-col items-center gap-0">
				<div className="flex items-center gap-1.5">
					<span className="font-medium text-lg">{totalValue}</span>
					<Icon icon="TON" className="w-4.5 h-4.5" />
				</div>
				<span className="text-muted-foreground">Total value</span>
			</div>
		</section>
	);
};
