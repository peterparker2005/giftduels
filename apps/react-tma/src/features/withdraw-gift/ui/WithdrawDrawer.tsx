import { ExecuteWithdrawRequest_CommissionCurrency } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_public_service_pb";
import { useEffect, useMemo, useState } from "react";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { usePreviewWithdraw } from "@/shared/api/queries/usePreviewWithdraw";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";
import { WithdrawForm } from "./WithdrawForm";
import { WithdrawSummary } from "./WithdrawSummary";

interface WithdrawDrawerProps {
	children: React.ReactNode;
	disabled?: boolean;
}

type WithdrawStep = "select" | "confirm";

export const WithdrawDrawer = ({ children, disabled }: WithdrawDrawerProps) => {
	const { data, isLoading, isFetchingNextPage, fetchNextPage, hasNextPage } =
		useGiftsQuery();
	const [step, setStep] = useState<WithdrawStep>("select");
	const [selectedGifts, setSelectedGifts] = useState<string[]>([]);
	const [selectedCommissionCurrency, setSelectedCommissionCurrency] =
		useState<ExecuteWithdrawRequest_CommissionCurrency>(
			ExecuteWithdrawRequest_CommissionCurrency.TON,
		);
	const [isOpen, setIsOpen] = useState(false);

	// Flatten all pages into a single array of gifts
	const allGifts = data?.pages.flatMap((page) => page.gifts) || [];

	// Preview withdraw logic moved here for caching
	const {
		mutate: previewWithdraw,
		data: previewWithdrawData,
		isPending: isPreviewPending,
	} = usePreviewWithdraw();

	// Calculate total TON amount for selected gifts
	const totalTonAmount = useMemo(() => {
		if (!allGifts || selectedGifts.length === 0) return 0;

		return selectedGifts.reduce((total, giftId) => {
			const gift = allGifts.find((g) => g.giftId?.value === giftId);
			return total + (gift?.price?.value || 0);
		}, 0);
	}, [selectedGifts, allGifts]);

	// Preview withdraw when selected gifts change
	useEffect(() => {
		if (totalTonAmount > 0) {
			previewWithdraw(totalTonAmount);
		}
	}, [previewWithdraw, totalTonAmount]);

	const handleProceedToConfirm = (giftIds: string[]) => {
		setSelectedGifts(giftIds);
		setStep("confirm");
	};

	const handleBackToSelect = () => {
		setStep("select");
	};

	const handleToggleGift = (giftId: string) => {
		setSelectedGifts((prev) => {
			if (prev.includes(giftId)) {
				return prev.filter((id) => id !== giftId);
			}
			return [...prev, giftId];
		});
	};

	const handleSelectAll = () => {
		if (!allGifts) return;
		const allGiftIds = allGifts
			.map((gift) => gift.giftId?.value || "")
			.filter(Boolean);
		setSelectedGifts(allGiftIds);
	};

	const handleClearSelection = () => {
		setSelectedGifts([]);
	};

	const handleRemoveGift = (giftId: string) => {
		setSelectedGifts((prev) => prev.filter((id) => id !== giftId));
	};

	const handleWithdrawSuccess = () => {
		// Close the drawer and reset state
		setIsOpen(false);
	};

	const handleCommissionCurrencyChange = (
		currency: ExecuteWithdrawRequest_CommissionCurrency,
	) => {
		setSelectedCommissionCurrency(currency);
	};

	const handleDrawerOpenChange = (open: boolean) => {
		setIsOpen(open);
		if (!open) {
			// Reset state when drawer closes
			setStep("select");
			setSelectedGifts([]);
			setSelectedCommissionCurrency(
				ExecuteWithdrawRequest_CommissionCurrency.TON,
			);
		}
	};

	// Auto-navigate back to select screen if all gifts are removed
	useEffect(() => {
		if (step === "confirm" && selectedGifts.length === 0) {
			setStep("select");
		}
	}, [step, selectedGifts.length]);

	const getTitle = () => {
		switch (step) {
			case "select":
				return "Select gifts for withdrawal";
			case "confirm":
				return "Confirm withdrawal";
			default:
				return "Withdraw gifts";
		}
	};

	const renderContent = () => {
		if (isLoading) {
			return (
				<div className="flex items-center justify-center flex-1">
					<p className="text-muted-foreground">Loading gifts...</p>
				</div>
			);
		}

		if (!allGifts || allGifts.length === 0) {
			return (
				<div className="flex items-center justify-center flex-1">
					<p className="text-muted-foreground">
						No gifts available for withdrawal
					</p>
				</div>
			);
		}

		switch (step) {
			case "select":
				return (
					<WithdrawForm
						gifts={allGifts}
						selectedGifts={selectedGifts}
						onProceedToConfirm={handleProceedToConfirm}
						onGiftToggle={handleToggleGift}
						onSelectAll={handleSelectAll}
						onClearSelection={handleClearSelection}
						previewData={previewWithdrawData}
						isPreviewPending={isPreviewPending}
						isLoadingMore={isFetchingNextPage}
						onLoadMore={fetchNextPage}
						hasNextPage={hasNextPage}
					/>
				);
			case "confirm":
				return (
					<WithdrawSummary
						gifts={allGifts}
						selectedGiftIds={selectedGifts}
						selectedCommissionCurrency={selectedCommissionCurrency}
						previewData={previewWithdrawData}
						onRemoveGift={handleRemoveGift}
						onBack={handleBackToSelect}
						onSuccess={handleWithdrawSuccess}
						onCommissionCurrencyChange={handleCommissionCurrencyChange}
					/>
				);
			default:
				return null;
		}
	};

	return (
		<Drawer open={isOpen} onOpenChange={handleDrawerOpenChange}>
			<DrawerTrigger asChild disabled={disabled}>
				{children}
			</DrawerTrigger>
			<DrawerContent className="h-[90vh] px-4 pt-4 flex flex-col">
				<div className="px-0 mb-4 flex-shrink-0">
					<DrawerTitle className="text-lg">{getTitle()}</DrawerTitle>
				</div>

				<div className="flex-1 min-h-0">{renderContent()}</div>
			</DrawerContent>
		</Drawer>
	);
};
