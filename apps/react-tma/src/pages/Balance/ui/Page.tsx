import { TonWallet } from "@/features/ton-connect-wallet/ui/TonWallet";
import { TonTransactionHistory } from "./TonTransactionHistory";
import { WalletBalance } from "./WalletBalance";

const Page = () => {
	return (
		<div className="flex flex-col gap-4 container">
			<TonWallet />
			<WalletBalance />
			<h2 className="text-2xl font-bold">Recent actions</h2>
			<TonTransactionHistory />
		</div>
	);
};

export default Page;
