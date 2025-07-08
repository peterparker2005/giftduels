import { TonWallet } from "@/features/ton-connect-wallet/ui/TonWallet";
import { WalletBalance } from "./WalletBalance";

const Page = () => {
	return (
		<div className="flex flex-col gap-4 container">
			<TonWallet />
			<WalletBalance />
		</div>
	);
};

export default Page;
