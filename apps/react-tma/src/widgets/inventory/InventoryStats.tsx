import { IoIosGift } from "react-icons/io";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { Icon } from "@/shared/ui/Icon/Icon";

export const InventoryStats = () => {
	const { data } = useGiftsQuery();

	return (
		<section className="flex items-stretch justify-between">
			<div className="flex flex-col items-center gap-1">
				<div className="flex items-center gap-1.5">
					<span className="font-medium">{data?.pagination?.total}</span>
					<IoIosGift className="w-3.5 h-3.5 shrink-0" />
				</div>
				<span className="text-muted-foreground">All gifts</span>
			</div>
			<div className="w-px h-4 bg-muted-foreground" />
			<div className="flex flex-col items-center gap-1">
				<div className="flex items-center gap-1.5">
					<span className="font-medium">{data?.totalValue}</span>
					<Icon icon="TON" className="w-3.5 h-3.5" />
				</div>
				<span className="text-muted-foreground">Total value</span>
			</div>
		</section>
	);
};
