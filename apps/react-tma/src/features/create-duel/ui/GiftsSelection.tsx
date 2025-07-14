import { useMemo } from "react";
import { useGiftsQuery } from "@/shared/api/queries/useGiftsQuery";
import { Button } from "@/shared/ui/Button";
import { getFragmentUrl } from "@/shared/utils/getFragmentUrl";

interface GiftsSelectionProps {
	selectedGifts: string[];
	onAddGifts: () => void;
	onRemoveGift: (giftId: string) => void;
}

export function GiftsSelection({
	selectedGifts,
	onAddGifts,
	onRemoveGift,
}: GiftsSelectionProps) {
	const { data: gifts } = useGiftsQuery();

	const selectedGiftsData = useMemo(() => {
		return selectedGifts.map((giftId) =>
			gifts?.pages
				.flatMap((page) => page.gifts)
				.find((g) => g.giftId?.value === giftId),
		);
	}, [selectedGifts, gifts]);

	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between">
				<h3 className="text-base font-semibold">Gifts</h3>
				<button
					type="button"
					onClick={onAddGifts}
					className="text-base font-semibold text-primary"
				>
					Add gifts
				</button>
			</div>

			{selectedGifts.length > 0 ? (
				<div className="flex overflow-x-auto gap-2">
					{selectedGiftsData.map((gift) => (
						<div
							key={gift?.giftId?.value}
							className="flex flex-col items-center justify-between bg-card-muted-accent rounded-3xl overflow-hidden shrink-0"
						>
							<img
								src={getFragmentUrl(gift?.slug || "")}
								alt={gift?.giftId?.value}
								className="w-24 h-24 object-cover"
							/>

							<div className="text-xs py-1 font-medium">
								<span>{gift?.price?.value} TON</span>
							</div>
							{/* <button
									type="button"
									onClick={() => onRemoveGift(giftId)}
									className="text-muted-foreground hover:text-destructive text-sm"
								>
									Remove
								</button> */}
						</div>
					))}
				</div>
			) : (
				<div className="bg-card rounded-lg p-6 text-center">
					<p className="text-muted-foreground text-sm">
						No gifts selected. Add gifts to start the duel.
					</p>
				</div>
			)}
		</div>
	);
}
