import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { useState } from "react";
import { Button } from "@/shared/ui/Button";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";
import { Icon } from "@/shared/ui/Icon/Icon";

interface DisconnectWalletDrawerProps {
	children: React.ReactNode;
	onDisconnect: () => void;
}

export const DisconnectWalletDrawer = ({
	children,
	onDisconnect,
}: DisconnectWalletDrawerProps) => {
	const [isOpen, setIsOpen] = useState(false);

	const handleDisconnect = () => {
		onDisconnect();
		setIsOpen(false);
	};

	return (
		<Drawer open={isOpen} onOpenChange={setIsOpen}>
			<DrawerTrigger asChild>{children}</DrawerTrigger>
			<DrawerContent className="container">
				<VisuallyHidden>
					<DrawerTitle>Disconnect Wallet</DrawerTitle>
				</VisuallyHidden>
				<section className="flex flex-col h-full items-center mt-10 mb-10">
					<div className="flex items-center justify-center rounded-full bg-card-muted-accent size-16">
						<Icon icon="TON" className="size-8" />
					</div>
					<span className="text-2xl font-semibold mt-4">Disconnect Wallet</span>
					<p className="text-card-muted-foreground mt-4 text-center">
						You can disconnect your wallet at any time. However, deposits in TON
						will not be available while it's disconnected.
					</p>
				</section>
				<div className="mb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
					<Button
						className="text-lg font-medium w-full h-12"
						onClick={handleDisconnect}
					>
						Disconnect
					</Button>
				</div>
			</DrawerContent>
		</Drawer>
	);
};
