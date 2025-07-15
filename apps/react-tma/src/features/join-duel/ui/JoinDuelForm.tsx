import { Duel } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_pb";
import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useState } from "react";
import { JoinDuelFormData } from "../hooks/useJoinDuelForm";
import { SelectGiftsForJoinForm } from "./SelectGiftsForJoinForm";

interface JoinDuelFormProps {
	duel: Duel;
	gifts: GiftView[];
	selectedGifts?: string[];
	onGiftToggle?: (giftId: string) => void;
	onSelectAll?: () => void;
	onClearSelection?: () => void;
	onJoinDuel: (data: JoinDuelFormData) => void;
	isLoadingMore?: boolean;
	onLoadMore?: () => void;
	hasNextPage?: boolean;
	isPending?: boolean;
}

export function JoinDuelForm({
	duel,
	gifts,
	selectedGifts: externalSelectedGifts,
	onGiftToggle,
	onSelectAll,
	onClearSelection,
	onJoinDuel,
	isLoadingMore = false,
	onLoadMore,
	hasNextPage = false,
	isPending = false,
}: JoinDuelFormProps) {
	const [selectedGifts, setSelectedGifts] = useState<string[]>(
		externalSelectedGifts || [],
	);

	const handleJoinDuel = (gifts: string[]) => {
		onJoinDuel({
			selectedGifts: gifts,
		});
	};

	const handleGiftToggle = (giftId: string) => {
		if (onGiftToggle) {
			onGiftToggle(giftId);
		} else {
			setSelectedGifts((prev) => {
				if (prev.includes(giftId)) {
					return prev.filter((id) => id !== giftId);
				}
				return [...prev, giftId];
			});
		}
	};

	const handleSelectAll = () => {
		if (onSelectAll) {
			onSelectAll();
		} else {
			const allGiftIds = gifts
				.map((gift) => gift.giftId?.value || "")
				.filter(Boolean);
			setSelectedGifts(allGiftIds);
		}
	};

	const handleClearSelection = () => {
		if (onClearSelection) {
			onClearSelection();
		} else {
			setSelectedGifts([]);
		}
	};

	const currentSelectedGifts = externalSelectedGifts || selectedGifts;

	return (
		<div className="flex flex-col h-full">
			<SelectGiftsForJoinForm
				gifts={gifts}
				duel={duel}
				selectedGifts={currentSelectedGifts}
				onJoinDuel={handleJoinDuel}
				onGiftToggle={handleGiftToggle}
				onSelectAll={handleSelectAll}
				onClearSelection={handleClearSelection}
				isLoadingMore={isLoadingMore}
				onLoadMore={onLoadMore}
				hasNextPage={hasNextPage}
				isPending={isPending}
			/>
		</div>
	);
}
