import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";
import { WithdrawForm } from "./WithdrawForm";

interface WithdrawDrawerProps {
	children: React.ReactNode;
	disabled?: boolean;
}

export const WithdrawDrawer = ({ children, disabled }: WithdrawDrawerProps) => {
	const { data, isLoading } = useGiftsQuery();

	return (
		<Drawer>
			<DrawerTrigger asChild disabled={disabled}>
				{children}
			</DrawerTrigger>
			<DrawerContent className="h-[90vh] px-4 pt-4">
				<DrawerTitle>Select gifts for withdrawal</DrawerTitle>

				{isLoading ? (
					<div className="flex items-center justify-center flex-1">
						<p className="text-muted-foreground">Loading gifts...</p>
					</div>
				) : data?.gifts && data.gifts.length > 0 ? (
					<WithdrawForm gifts={data.gifts} />
				) : (
					<div className="flex items-center justify-center flex-1">
						<p className="text-muted-foreground">
							No gifts available for withdrawal
						</p>
					</div>
				)}
			</DrawerContent>
		</Drawer>
	);
};
