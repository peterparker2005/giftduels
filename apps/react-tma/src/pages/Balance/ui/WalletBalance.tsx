import { DepositTonDialog } from "@/features/deposit-ton/ui/DepositTonDialog";
import { useBalanceQuery } from "@/shared/api/queries/useBalanceQuery";
import { Button } from "@/shared/ui/Button";
import { formatTonAmount } from "@/shared/utils/formatTonAmount";

export const WalletBalance = () => {
	const { data } = useBalanceQuery();
	return (
		<div className="h-40 relative rounded-3xl overflow-hidden bg-linear-180 from-primary to-[#004999] to-150% flex flex-col items-center">
			<img
				src="/dice_wallet_pattern.png"
				alt="balance-bg"
				className="absolute top-0 left-0 w-full h-full object-cover z-0"
			/>
			<div className="py-4 flex flex-col items-center justify-between h-full z-10">
				<p className="font-medium">GiftDuels Wallet Balance</p>
				<span className="font-bold text-3xl">
					{formatTonAmount(data?.balance?.tonAmount?.value)} TON
				</span>
				<DepositTonDialog>
					<Button className="w-full text-base">
						<span>Deposit</span>
					</Button>
				</DepositTonDialog>
			</div>
		</div>
	);
};
