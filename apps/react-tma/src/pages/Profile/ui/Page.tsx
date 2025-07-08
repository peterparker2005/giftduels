import { BiLogoTelegram, BiPlus } from "react-icons/bi";
import { AddGiftInfoDrawer } from "@/features/add-gift/ui/AddGiftInfoDrawer";
import { WithdrawDrawer } from "@/features/withdraw-gift/ui/WithdrawDrawer";
import { useProfileQuery } from "@/shared/api/queries/useProfileQuery";
import { Avatar } from "@/shared/ui/Avatar";
import { Button } from "@/shared/ui/Button";
import { Inventory } from "@/widgets/inventory/Inventory";

export const Page = () => {
	const { data } = useProfileQuery();

	return (
		<div className="container pt-4 relative flex-1">
			<div className="flex flex-col items-center gap-2">
				<Avatar className="w-20 h-20" />
				<h1 className="text-2xl font-medium">{data?.profile?.displayName}</h1>
			</div>
			<section className="flex flex-items-center gap-2 w-full mt-4 mb-4">
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
