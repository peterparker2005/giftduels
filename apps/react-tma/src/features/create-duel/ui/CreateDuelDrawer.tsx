import { create } from "@bufbuild/protobuf";
import { CreateDuelRequestSchema } from "@giftduels/protobuf-js/giftduels/duel/v1/duel_public_service_pb";
import { useRouter } from "@tanstack/react-router";
import { useCallback, useMemo, useState } from "react";
import { FormProvider } from "react-hook-form";
import { toast } from "sonner";
import { useCreateDuelMutation } from "@/shared/api/queries/useCreateDuelMutation";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";
import { useCreateDuelForm } from "../hooks/useCreateDuelForm";
import { CreateDuelForm } from "./CreateDuelForm";
import { SelectGiftsForm } from "./SelectGiftsForm";

type CreateDuelDrawerProps = {
	children: React.ReactNode;
};

type CreateDuelStep = "form" | "select-gifts";

export function CreateDuelDrawer({ children }: CreateDuelDrawerProps) {
	const { data, isLoading, isFetchingNextPage, fetchNextPage, hasNextPage } =
		useGiftsQuery();
	const [step, setStep] = useState<CreateDuelStep>("form");
	const [selectedGifts, setSelectedGifts] = useState<string[]>([]);
	const [isOpen, setIsOpen] = useState(false);

	const form = useCreateDuelForm();

	const router = useRouter();

	const { mutate: createDuel, isPending: isCreateDuelPending } =
		useCreateDuelMutation();

	// Flatten all pages into a single array of gifts
	const allGifts = useMemo(
		() => data?.pages.flatMap((page) => page.gifts) || [],
		[data?.pages],
	);

	const handleAddGifts = useCallback(() => {
		setStep("select-gifts");
	}, []);

	const handleBackToForm = useCallback(() => {
		setStep("form");
	}, []);

	const handleGiftToggle = useCallback(
		(giftId: string) => {
			setSelectedGifts((prev) => {
				if (prev.includes(giftId)) {
					const newSelection = prev.filter((id) => id !== giftId);
					form.setValue("gifts", newSelection);
					return newSelection;
				}
				const newSelection = [...prev, giftId];
				form.setValue("gifts", newSelection);
				return newSelection;
			});
		},
		[form],
	);

	const handleSelectAll = useCallback(() => {
		if (!allGifts) return;
		const allGiftIds = allGifts
			.map((gift) => gift.giftId?.value || "")
			.filter(Boolean);
		setSelectedGifts(allGiftIds);
		form.setValue("gifts", allGiftIds);
	}, [allGifts, form]);

	const handleClearSelection = useCallback(() => {
		setSelectedGifts([]);
		form.setValue("gifts", []);
	}, [form]);

	const handleRemoveGift = useCallback(
		(giftId: string) => {
			setSelectedGifts((prev) => {
				const newSelection = prev.filter((id) => id !== giftId);
				form.setValue("gifts", newSelection);
				return newSelection;
			});
		},
		[form],
	);

	const handleConfirmGifts = useCallback(() => {
		setStep("form");
	}, []);

	const handleSubmit = useCallback(() => {
		const formData = form.getValues();
		console.log("Creating duel with data:", {
			...formData,
			gifts: selectedGifts,
		});

		const stakes = selectedGifts.map((giftId) => ({
			giftId: {
				value: giftId,
			},
		}));

		const request = create(CreateDuelRequestSchema, {
			params: {
				isPrivate: formData.privacy === "private",
				maxPlayers: formData.players,
				maxGifts: selectedGifts.length,
			},
			stakes,
		});

		createDuel(request, {
			onSuccess: (data) => {
				router.navigate({
					to: "/duel/$duelId",
					params: {
						duelId: data.duelId?.value || "",
					},
				});
			},
			onError: () => {
				toast.error("Failed to create duel");
			},
		});
	}, [form, selectedGifts, createDuel, router]);

	const handleDrawerOpenChange = useCallback(
		(open: boolean) => {
			setIsOpen(open);
			if (!open) {
				// Reset state when drawer closes
				setStep("form");
				setSelectedGifts([]);
				form.reset();
			} else {
				// Sync form gifts with selected gifts when opening
				form.setValue("gifts", selectedGifts);
			}
		},
		[form, selectedGifts],
	);

	const getTitle = useCallback(() => {
		switch (step) {
			case "form":
				return "Create Duel";
			case "select-gifts":
				return "Select Gifts";
			default:
				return "Create Duel";
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
				<div className="flex items-center justify-center flex-1">
					<p className="text-muted-foreground">No gifts available for duels</p>
				</div>
			);
		}

		switch (step) {
			case "form":
				return (
					<FormProvider {...form}>
						<CreateDuelForm
							selectedGifts={selectedGifts}
							onAddGifts={handleAddGifts}
							onRemoveGift={handleRemoveGift}
							onSubmit={handleSubmit}
						/>
					</FormProvider>
				);
			case "select-gifts":
				return (
					<SelectGiftsForm
						gifts={allGifts}
						selectedGifts={selectedGifts}
						onGiftToggle={handleGiftToggle}
						onSelectAll={handleSelectAll}
						onClearSelection={handleClearSelection}
						onConfirm={handleConfirmGifts}
						onBack={handleBackToForm}
						isLoadingMore={isFetchingNextPage}
						onLoadMore={fetchNextPage}
						hasNextPage={hasNextPage}
					/>
				);
			default:
				return null;
		}
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
