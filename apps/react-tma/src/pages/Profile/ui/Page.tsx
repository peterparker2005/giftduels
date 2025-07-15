import { BiLogoTelegram, BiPlus } from "react-icons/bi";
import { AddGiftInfoDrawer } from "@/features/add-gift/ui/AddGiftInfoDrawer";
import { WithdrawDrawer } from "@/features/withdraw-gift/ui/WithdrawDrawer";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { useProfileQuery } from "@/shared/api/queries/useProfileQuery";
import { Avatar } from "@/shared/ui/Avatar";
import { Button } from "@/shared/ui/Button";
import { Inventory } from "@/widgets/inventory/Inventory";
import { InventoryStats } from "@/widgets/inventory/InventoryStats";

export const Page = () => {
	const { data } = useProfileQuery();

	const { data: gifts } = useGiftsQuery();

	return (
		<div className="flex-1 container">
			<div className="flex flex-col items-center gap-2">
				<Avatar className="w-20 h-20" />
				<h1 className="text-2xl font-medium">{data?.profile?.displayName}</h1>
			</div>
			{gifts?.pages[0]?.gifts && (
				<div className="mt-4">
					<InventoryStats />
				</div>
			)}
			<section className="flex flex-items-center gap-4 w-full mt-4 mb-4">
				<AddGiftInfoDrawer>
					<Button
						variant="secondary"
						className="w-full flex flex-col items-center gap-1"
					>
						<BiPlus className="w-5 h-5" />
						<span className="text-base font-medium">Add</span>
					</Button>
				</AddGiftInfoDrawer>
				<WithdrawDrawer>
					<Button
						variant="secondary"
						className="w-full flex flex-col items-center gap-1"
					>
						<BiLogoTelegram className="w-5 h-5" />
						<span className="text-base font-medium">Withdraw</span>
					</Button>
				</WithdrawDrawer>
			</section>

			{/* h-full */}
			<Inventory />
		</div>
	);
};

export default Page;
