import { create } from "@bufbuild/protobuf";
import { GiftStatus } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { ExecuteWithdrawRequest_CommissionCurrency } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_public_service_pb";
import { GiftWithdrawRequestSchema } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import {
	GiftIdSchema,
	TonAmountSchema,
} from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { openTelegramLink } from "@telegram-apps/sdk";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { usePreviewWithdraw } from "@/shared/api/queries/usePreviewWithdraw";
import { Button } from "@/shared/ui/Button";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";
import { LottiePlayer } from "@/shared/ui/LottiePlayer";
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
	const allGifts = useMemo(
		() =>
			data?.pages
				.flatMap((page) => page.gifts)
				.filter((gift) => gift.status !== GiftStatus.IN_GAME) || [],
		[data?.pages],
	);

	// Preview withdraw logic moved here for caching
	const {
		mutate: previewWithdraw,
		data: previewWithdrawData,
		isPending: isPreviewPending,
		reset: resetPreviewWithdraw,
	} = usePreviewWithdraw();

	// Create a stable reference for gift data
	const giftDataMap = useMemo(() => {
		const map = new Map();
		allGifts.forEach((gift) => {
			if (gift.giftId?.value) {
				map.set(gift.giftId.value, gift);
			}
		});
		return map;
	}, [allGifts]);

	// Preview withdraw when selected gifts change
	useEffect(() => {
		// Не делать preview запрос для пустого массива
		if (selectedGifts.length === 0) {
			resetPreviewWithdraw();
			return;
		}

		const gifts = selectedGifts.map((giftId) => {
			const gift = giftDataMap.get(giftId);
			return create(GiftWithdrawRequestSchema, {
				giftId: create(GiftIdSchema, { value: giftId }),
				price: create(TonAmountSchema, {
					value: gift?.price?.value || "0",
				}),
			});
		});
		previewWithdraw(gifts);
	}, [previewWithdraw, selectedGifts, giftDataMap, resetPreviewWithdraw]);

	const handleProceedToConfirm = useCallback((giftIds: string[]) => {
		setSelectedGifts(giftIds);
		setStep("confirm");
	}, []);

	const handleBackToSelect = useCallback(() => {
		setStep("select");
	}, []);

	const handleToggleGift = useCallback((giftId: string) => {
		setSelectedGifts((prev) => {
			if (prev.includes(giftId)) {
				return prev.filter((id) => id !== giftId);
			}
			return [...prev, giftId];
		});
	}, []);

	const handleSelectAll = useCallback(() => {
		if (!allGifts) return;
		const allGiftIds = allGifts
			.map((gift) => gift.giftId?.value || "")
			.filter(Boolean);
		setSelectedGifts(allGiftIds);
	}, [allGifts]);

	const handleClearSelection = useCallback(() => {
		setSelectedGifts([]);
	}, []);

	const handleRemoveGift = useCallback((giftId: string) => {
		setSelectedGifts((prev) => prev.filter((id) => id !== giftId));
	}, []);

	const handleWithdrawSuccess = useCallback(() => {
		// Close the drawer and reset state
		setIsOpen(false);
		resetPreviewWithdraw();
	}, [resetPreviewWithdraw]);

	const handleCommissionCurrencyChange = useCallback(
		(currency: ExecuteWithdrawRequest_CommissionCurrency) => {
			setSelectedCommissionCurrency(currency);
		},
		[],
	);

	const handleDrawerOpenChange = useCallback((open: boolean) => {
		setIsOpen(open);
		if (!open) {
			// Reset state when drawer closes
			setStep("select");
			setSelectedGifts([]);
			setSelectedCommissionCurrency(
				ExecuteWithdrawRequest_CommissionCurrency.TON,
			);
		}
	}, []);

	// Auto-navigate back to select screen if all gifts are removed
	useEffect(() => {
		if (step === "confirm" && selectedGifts.length === 0) {
			setStep("select");
		}
	}, [step, selectedGifts.length]);

	const getTitle = useCallback(() => {
		switch (step) {
			case "select":
				return "Select gifts for withdrawal";
			case "confirm":
				return "Confirm withdrawal";
			default:
				return "Withdraw gifts";
		}
	}, [step]);

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
				<div className="flex items-center justify-center flex-1 h-[calc(100%-28px-20px)]">
					<div className="flex flex-col items-center gap-4">
						<LottiePlayer
							src={"/lottie/plushdog.json"}
							loop={true}
							className="w-40 h-40"
						/>
						<p className="text-muted-foreground text-center font-medium text-base">
							No gifts available for withdrawal
						</p>
						<Button
							variant="default"
							className=""
							onClick={() => {
								openTelegramLink("https://t.me/devpp2_bot");
							}}
						>
							Add gifts
						</Button>
					</div>
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
