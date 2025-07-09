import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import { useExecuteWithdrawMutation } from "@/shared/api/queries/useExecuteWithdrawMutation";

export interface WithdrawFormData {
	selectedGifts: string[]; // giftIds
}

export const useWithdrawForm = (gifts: GiftView[] = []) => {
	const [selectedGifts, setSelectedGifts] = useState<string[]>([]);
	const { mutate } = useExecuteWithdrawMutation();
	const queryClient = useQueryClient();
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
		mutate(selectedGifts, {
			onSuccess: (data) => {
				console.log("Withdrawal successful", data);
				clearSelection();
				queryClient.invalidateQueries({ queryKey: ["gifts"] });
			},
			onError: (error) => {
				toast.error(error.message, { position: "top-center" });
				console.error("Withdrawal failed", error);
			},
		});
	}, [mutate, selectedGifts, clearSelection, queryClient]);

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
