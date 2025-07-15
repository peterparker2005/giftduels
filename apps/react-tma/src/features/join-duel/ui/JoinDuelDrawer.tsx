import { Duel } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_pb";
import { GiftStatus } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
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
	onJoinDuel?: (data: JoinDuelFormData) => Promise<void>;
}

export function JoinDuelDrawer({
	displayNumber,
	duel,
	children,
	onJoinDuel,
}: JoinDuelDrawerProps) {
	const { data, isLoading, isFetchingNextPage, fetchNextPage, hasNextPage } =
		useGiftsQuery();
	const [selectedGifts, setSelectedGifts] = useState<string[]>([]);
	const [isOpen, setIsOpen] = useState(false);
	const [isPending, setIsPending] = useState(false);

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
			if (!onJoinDuel) {
				console.log("Join duel data:", data);
				toast.success("Join duel functionality will be implemented soon!");
				return;
			}

			setIsPending(true);
			try {
				await onJoinDuel(data);
				toast.success("Successfully joined the duel!");
				setIsOpen(false);
			} catch (error) {
				toast.error("Failed to join duel");
				console.error("Join duel error:", error);
			} finally {
				setIsPending(false);
			}
		},
		[onJoinDuel],
	);

	const handleDrawerOpenChange = useCallback((open: boolean) => {
		setIsOpen(open);
		if (!open) {
			// Reset state when drawer closes
			setSelectedGifts([]);
			setIsPending(false);
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
				isPending={isPending}
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
