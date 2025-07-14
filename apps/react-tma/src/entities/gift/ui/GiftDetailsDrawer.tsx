import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { useState } from "react";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";
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
				<VisuallyHidden>
					<DrawerTitle>{gift.title} details</DrawerTitle>
				</VisuallyHidden>
				<GiftDetailsCard gift={gift} onCloseDrawer={() => setIsOpen(false)} />
			</DrawerContent>
		</Drawer>
	);
};
