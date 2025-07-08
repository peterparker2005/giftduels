import { useTonConnectUI } from "@tonconnect/ui-react";
import { Icon } from "@/shared/ui/Icon/Icon";

export const ConnectWallet = () => {
	const [tonConnectUI] = useTonConnectUI();

	const openModal = () => {
		tonConnectUI.openModal();
	};

	return (
		<button
			type="button"
			className="bg-card rounded-3xl px-5 py-3.5 flex items-center justify-center font-medium gap-2"
			onClick={openModal}
		>
			<Icon icon="TON" className="w-4 h-4" />
			<span>Connect wallet</span>
		</button>
	);
};
