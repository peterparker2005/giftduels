import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useCallback, useEffect, useState } from "react";

export interface WithdrawFormData {
	selectedGifts: string[]; // giftIds
}

export const useWithdrawForm = (
	gifts: GiftView[] = [],
	externalSelectedGifts?: string[],
) => {
	const [internalSelectedGifts, setInternalSelectedGifts] = useState<string[]>(
		[],
	);

	// Use external state if provided, otherwise use internal state
	const selectedGifts = externalSelectedGifts ?? internalSelectedGifts;
	// Sync internal state with external state when external state changes
	useEffect(() => {
		if (externalSelectedGifts !== undefined) {
			setInternalSelectedGifts(externalSelectedGifts);
		}
	}, [externalSelectedGifts]);

	const isGiftSelected = useCallback(
		(giftId: string) => selectedGifts.includes(giftId),
		[selectedGifts],
	);

	const toggleGift = useCallback(
		(giftId: string) => {
			if (externalSelectedGifts) {
				// Don't modify external state directly - this should be handled by parent
				return;
			}

			setInternalSelectedGifts((prev) => {
				if (prev.includes(giftId)) {
					return prev.filter((id) => id !== giftId);
				}
				return [...prev, giftId];
			});
		},
		[externalSelectedGifts],
	);

	const selectAll = useCallback(() => {
		if (externalSelectedGifts) {
			// Don't modify external state directly
			return;
		}

		const allGiftIds = gifts
			.map((gift) => gift.giftId?.value || "")
			.filter(Boolean);
		setInternalSelectedGifts(allGiftIds);
	}, [gifts, externalSelectedGifts]);

	const clearSelection = useCallback(() => {
		if (externalSelectedGifts) {
			// Don't modify external state directly
			return;
		}

		setInternalSelectedGifts([]);
	}, [externalSelectedGifts]);

	const isAllSelected =
		gifts.length > 0 && selectedGifts.length === gifts.length;
	const hasSelection = selectedGifts.length > 0;
	const selectedCount = selectedGifts.length;

	return {
		selectedGifts,
		isGiftSelected,
		toggleGift,
		selectAll,
		clearSelection,
		isAllSelected,
		hasSelection,
		selectedCount,
	};
};
