import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useState } from "react";
import { SelectGiftCard } from "@/entities/gift/ui/SelectGiftCard";
import { Button } from "@/shared/ui/Button";

interface ConfirmJoinDuelFormProps {
	selectedGifts: string[];
	gifts: GiftView[];
	duelId: string;
	onConfirm: (data: {
		selectedGifts: string[];
		entryPrice?: { from: number; to: number };
	}) => void;
	onBack: () => void;
	isPending?: boolean;
}

export function ConfirmJoinDuelForm({
	selectedGifts,
	gifts,
	duelId,
	onConfirm,
	onBack,
	isPending = false,
}: ConfirmJoinDuelFormProps) {
	const [entryPriceFrom, setEntryPriceFrom] = useState<string>("");
	const [entryPriceTo, setEntryPriceTo] = useState<string>("");

	const selectedGiftObjects = gifts.filter((gift) =>
		selectedGifts.includes(gift.giftId?.value || ""),
	);

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		const entryPrice =
			entryPriceFrom && entryPriceTo
				? {
						from: parseFloat(entryPriceFrom),
						to: parseFloat(entryPriceTo),
					}
				: undefined;

		onConfirm({
			selectedGifts,
			entryPrice,
		});
	};

	return (
		<form onSubmit={handleSubmit} className="flex flex-col h-full">
			<div className="flex flex-col gap-4 flex-1 overflow-y-auto min-h-0 mb-4">
				<div className="text-center">
					<h3 className="text-lg font-semibold">Confirm Join Duel</h3>
					<p className="text-muted-foreground">Duel #{duelId}</p>
				</div>

				{/* Entry Price Range (Future Feature) */}
				<div className="space-y-3">
					<label htmlFor="entry-price-section" className="text-sm font-medium">
						Entry Price Range (Optional)
					</label>
					<div className="flex gap-2">
						<div className="flex-1">
							<label
								htmlFor="entry-price-from"
								className="text-xs text-muted-foreground"
							>
								From (TON)
							</label>
							<input
								id="entry-price-from"
								type="number"
								step="0.01"
								min="0"
								placeholder="0.00"
								value={entryPriceFrom}
								onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
									setEntryPriceFrom(e.target.value)
								}
								className="mt-1 w-full px-3 py-2 bg-card rounded-lg border border-border focus:outline-none focus:ring-2 focus:ring-primary"
							/>
						</div>
						<div className="flex-1">
							<label
								htmlFor="entry-price-to"
								className="text-xs text-muted-foreground"
							>
								To (TON)
							</label>
							<input
								id="entry-price-to"
								type="number"
								step="0.01"
								min="0"
								placeholder="0.00"
								value={entryPriceTo}
								onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
									setEntryPriceTo(e.target.value)
								}
								className="mt-1 w-full px-3 py-2 bg-card rounded-lg border border-border focus:outline-none focus:ring-2 focus:ring-primary"
							/>
						</div>
					</div>
					<p className="text-xs text-muted-foreground">
						This feature will be available soon. Leave empty for now.
					</p>
				</div>

				{/* Selected Gifts */}
				<div className="space-y-3">
					<label
						htmlFor="selected-gifts-section"
						className="text-sm font-medium"
					>
						Selected Gifts ({selectedGiftObjects.length})
					</label>
					<div className="space-y-2">
						{selectedGiftObjects.map((gift) => {
							const giftId = gift.giftId?.value || "";
							return (
								<SelectGiftCard
									key={giftId}
									gift={gift}
									selected={true}
									onSelectionChange={() => {}} // Read-only in confirm view
								/>
							);
						})}
					</div>
				</div>
			</div>

			<div className="shrink-0 pb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)] space-y-2">
				<Button type="submit" disabled={isPending} className="w-full py-3">
					{isPending ? "Joining..." : "Join Duel"}
				</Button>
				<Button
					type="button"
					variant="secondary"
					onClick={onBack}
					disabled={isPending}
					className="w-full py-3"
				>
					Back to Selection
				</Button>
			</div>
		</form>
	);
}
