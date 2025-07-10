import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useCallback, useState } from "react";

export interface UseGiftDetailsReturn {
	isOpen: boolean;
	openGiftDetails: (gift: GiftView) => void;
	closeGiftDetails: () => void;
	selectedGift: GiftView | null;
}

export const useGiftDetails = (): UseGiftDetailsReturn => {
	const [isOpen, setIsOpen] = useState(false);
	const [selectedGift, setSelectedGift] = useState<GiftView | null>(null);

	const openGiftDetails = useCallback((gift: GiftView) => {
		setSelectedGift(gift);
		setIsOpen(true);
	}, []);

	const closeGiftDetails = useCallback(() => {
		setIsOpen(false);
		// Delay clearing the gift to allow for smooth closing animation
		setTimeout(() => {
			setSelectedGift(null);
		}, 200);
	}, []);

	return {
		isOpen,
		openGiftDetails,
		closeGiftDetails,
		selectedGift,
	};
};
