import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useCallback, useState } from "react";

export interface WithdrawFormData {
	selectedGifts: string[]; // giftIds
}

export const useWithdrawForm = (gifts: GiftView[] = []) => {
	const [selectedGifts, setSelectedGifts] = useState<string[]>([]);

	const isGiftSelected = useCallback(
		(giftId: string) => selectedGifts.includes(giftId),
		[selectedGifts],
	);

	const toggleGift = useCallback((giftId: string) => {
		setSelectedGifts((prev) => {
			if (prev.includes(giftId)) {
				return prev.filter((id) => id !== giftId);
			}
			return [...prev, giftId];
		});
	}, []);

	const selectAll = useCallback(() => {
		const allGiftIds = gifts
			.map((gift) => gift.giftId?.value || "")
			.filter(Boolean);
		setSelectedGifts(allGiftIds);
	}, [gifts]);

	const clearSelection = useCallback(() => {
		setSelectedGifts([]);
	}, []);

	const isAllSelected =
		gifts.length > 0 && selectedGifts.length === gifts.length;
	const hasSelection = selectedGifts.length > 0;
	const selectedCount = selectedGifts.length;

	const handleSubmit = useCallback(() => {
		// TODO: Implement withdrawal logic
		console.log("Withdrawing gifts:", selectedGifts);
	}, [selectedGifts]);

	return {
		selectedGifts,
		isGiftSelected,
		toggleGift,
		selectAll,
		clearSelection,
		isAllSelected,
		hasSelection,
		selectedCount,
		handleSubmit,
	};
};
