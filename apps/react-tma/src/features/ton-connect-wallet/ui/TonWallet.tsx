import {
	useTonAddress,
	useTonConnectUI,
	useTonWallet,
} from "@tonconnect/ui-react";
import { AiOutlineDisconnect } from "react-icons/ai";
import { FaWallet } from "react-icons/fa";
import { formatTonAddress } from "@/shared/utils/formatTonAddress";
import { ConnectWallet } from "./ConnectWallet";
import { DisconnectWalletDrawer } from "./DisconnectWalletDrawer";

export const TonWallet = () => {
	const wallet = useTonWallet();
	const address = useTonAddress();
	const [tonConnectUI] = useTonConnectUI();

	if (!wallet) return <ConnectWallet />;

	const handleDisconnect = async () => {
		await tonConnectUI.disconnect();
	};

	return (
		<div className="bg-card rounded-3xl px-5 py-3.5">
			<div className="flex items-center gap-2.5 text-muted-foreground">
				<FaWallet className="w-3.5 h-3.5 shrink-0" />
				<span className="mr-auto">{formatTonAddress(address)}</span>
				<DisconnectWalletDrawer onDisconnect={handleDisconnect}>
					<button
						type="button"
						className="flex items-center gap-1 font-medium text-primary"
					>
						<AiOutlineDisconnect className="w-4 h-4" />
						<span>Disconnect</span>
					</button>
				</DisconnectWalletDrawer>
			</div>
		</div>
	);
};
