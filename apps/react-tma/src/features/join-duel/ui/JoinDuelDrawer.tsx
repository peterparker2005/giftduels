import { create } from "@bufbuild/protobuf";
import { Duel } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_pb";
import { JoinDuelRequestSchema } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_public_service_pb";
import { GiftStatus } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { useJoinDuelMutation } from "@/shared/api/queries/useJoinDuelMutation";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";
import { JoinDuelFormData } from "../hooks/useJoinDuelForm";
import { JoinDuelForm } from "./JoinDuelForm";

interface JoinDuelDrawerProps {
	displayNumber: string;
	duel: Duel;
	children: React.ReactNode;
}

export function JoinDuelDrawer({
	displayNumber,
	duel,
	children,
}: JoinDuelDrawerProps) {
	const { data, isLoading, isFetchingNextPage, fetchNextPage, hasNextPage } =
		useGiftsQuery();
	const joinDuelMutation = useJoinDuelMutation();
	const [selectedGifts, setSelectedGifts] = useState<string[]>([]);
	const [isOpen, setIsOpen] = useState(false);

	// Flatten all pages into a single array of gifts
	const allAvailableGifts = useMemo(
		() =>
			data?.pages
				.flatMap((page) => page.gifts)
				.filter((gift) => gift.status === GiftStatus.OWNED) || [],
		[data?.pages],
	);

	const handleGiftToggle = useCallback((giftId: string) => {
		setSelectedGifts((prev) => {
			if (prev.includes(giftId)) {
				return prev.filter((id) => id !== giftId);
			}
			return [...prev, giftId];
		});
	}, []);

	const handleSelectAll = useCallback(() => {
		if (!allAvailableGifts) return;
		const allGiftIds = allAvailableGifts
			.map((gift) => gift.giftId?.value || "")
			.filter(Boolean);
		setSelectedGifts(allGiftIds);
	}, [allAvailableGifts]);

	const handleClearSelection = useCallback(() => {
		setSelectedGifts([]);
	}, []);

	const handleJoinDuel = useCallback(
		async (data: JoinDuelFormData) => {
			try {
				const stakes = data.selectedGifts.map((giftId) => ({
					giftId: {
						value: giftId,
					},
				}));

				const request = create(JoinDuelRequestSchema, {
					duelId: {
						value: duel.duelId?.value || "",
					},
					stakes,
				});

				await joinDuelMutation.mutateAsync(request);
				toast.success("Successfully joined the duel!");
				setIsOpen(false);
			} catch (error) {
				toast.error("Failed to join duel");
				console.error("Join duel error:", error);
			}
		},
		[joinDuelMutation, duel.duelId?.value],
	);

	const handleDrawerOpenChange = useCallback((open: boolean) => {
		setIsOpen(open);
		if (!open) {
			// Reset state when drawer closes
			setSelectedGifts([]);
		}
	}, []);

	const getTitle = useCallback(() => {
		return `Join Duel #${displayNumber}`;
	}, [displayNumber]);

	const renderContent = () => {
		if (isLoading) {
			return (
				<div className="flex items-center justify-center flex-1">
					<p className="text-muted-foreground">Loading gifts...</p>
				</div>
			);
		}

		if (!allAvailableGifts || allAvailableGifts.length === 0) {
			return (
				<div className="flex items-center justify-center flex-1">
					<p className="text-muted-foreground">No gifts available for duels</p>
				</div>
			);
		}

		return (
			<JoinDuelForm
				duel={duel}
				gifts={allAvailableGifts}
				selectedGifts={selectedGifts}
				onGiftToggle={handleGiftToggle}
				onSelectAll={handleSelectAll}
				onClearSelection={handleClearSelection}
				onJoinDuel={handleJoinDuel}
				isLoadingMore={isFetchingNextPage}
				onLoadMore={fetchNextPage}
				hasNextPage={hasNextPage}
				isPending={joinDuelMutation.isPending}
			/>
		);
	};

	return (
		<Drawer open={isOpen} onOpenChange={handleDrawerOpenChange}>
			<DrawerTrigger asChild>{children}</DrawerTrigger>
			<DrawerContent className="h-[90vh] px-4 pt-4 flex flex-col">
				<div className="px-0 mb-4 flex-shrink-0">
					<DrawerTitle className="text-lg">{getTitle()}</DrawerTitle>
				</div>

				<div className="flex-1 min-h-0">{renderContent()}</div>
			</DrawerContent>
		</Drawer>
	);
}
