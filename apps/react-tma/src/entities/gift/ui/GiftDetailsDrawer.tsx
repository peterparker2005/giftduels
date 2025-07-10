import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useState } from "react";
import { Drawer, DrawerContent, DrawerTrigger } from "@/shared/ui/Drawer";
import { GiftDetailsCard } from "./GiftDetailsCard";

interface GiftDetailsDrawerProps {
	gift: GiftView;
	children: React.ReactNode;
	disabled?: boolean;
}

export const GiftDetailsDrawer = ({
	gift,
	children,
	disabled = false,
}: GiftDetailsDrawerProps) => {
	const [isOpen, setIsOpen] = useState(false);

	const handleDrawerOpenChange = (open: boolean) => {
		setIsOpen(open);
	};

	return (
		<Drawer open={isOpen} onOpenChange={handleDrawerOpenChange}>
			<DrawerTrigger asChild disabled={disabled}>
				{children}
			</DrawerTrigger>
			<DrawerContent className="px-4 pt-4">
				<GiftDetailsCard gift={gift} onCloseDrawer={() => setIsOpen(false)} />
			</DrawerContent>
		</Drawer>
	);
};
