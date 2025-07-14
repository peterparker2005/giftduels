import { Button } from "@/shared/ui/Button";
import { GiftsSelection } from "./GiftsSelection";
import { PlayerCountSettings } from "./PlayerCountSettings";
import { PrivacySettings } from "./PrivacySettings";

interface CreateDuelFormProps {
	selectedGifts: string[];
	onAddGifts: () => void;
	onRemoveGift: (giftId: string) => void;
	onSubmit: () => void;
}

export function CreateDuelForm({
	selectedGifts,
	onAddGifts,
	onRemoveGift,
	onSubmit,
}: CreateDuelFormProps) {
	return (
		<form
			onSubmit={(e) => {
				e.preventDefault();
				onSubmit();
			}}
			className="flex flex-col h-full"
		>
			<div className="flex-1 overflow-y-auto min-h-0 gap-4 flex flex-col">
				<GiftsSelection
					selectedGifts={selectedGifts}
					onAddGifts={onAddGifts}
					onRemoveGift={onRemoveGift}
				/>
				<PlayerCountSettings />
				<PrivacySettings />
			</div>

			{/* Submit Button */}
			<div className="shrink-0 pt-6 pb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
				<Button
					type="submit"
					disabled={selectedGifts.length === 0}
					className="w-full py-3"
				>
					Create Duel
				</Button>
			</div>
		</form>
	);
}
