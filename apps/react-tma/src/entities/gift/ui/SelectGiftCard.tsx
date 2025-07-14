import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { Checkbox } from "@/shared/ui/Checkbox";
import { Icon } from "@/shared/ui/Icon/Icon";
import { getFragmentUrl } from "@/shared/utils/getFragmentUrl";

interface SelectGiftCardProps {
	gift: GiftView;
	selected?: boolean;
	onSelectionChange: (selected: boolean) => void;
}

export const SelectGiftCard = ({
	gift,
	selected = false,
	onSelectionChange,
}: SelectGiftCardProps) => {
	const handleClick = () => {
		onSelectionChange(!selected);
	};

	const handleCheckboxChange = (isSelected: boolean) => {
		onSelectionChange(isSelected);
	};

	const CardContent = () => (
		<>
			<div className="relative w-18 h-18 overflow-hidden rounded-2xl cursor-default">
				<img
					// src={`https://nft.fragment.com/gift/${gift.slug.toLowerCase()}.large.jpg`}
					src={getFragmentUrl(gift.slug, "large")}
					alt={gift.title}
				/>
			</div>
			<div className="flex flex-col gap-2">
				<div className="flex gap-2">
					<p className="text-card-foreground font-semibold text-base">
						{gift.title}
					</p>
					<p className="text-card-muted-foreground text-base">
						#{gift.collectibleId}
					</p>
				</div>
				<div className="flex items-center gap-2">
					<div className="rounded-full bg-[#2D9EED] w-5 h-5 flex items-center justify-center">
						<Icon icon="TON" className="w-3 h-3 shrink-0" />
					</div>
					<span className="text-xs font-medium">{gift.price?.value}</span>
				</div>
			</div>
			<div
				className="absolute top-1/2 -translate-y-1/2 right-4 p-1"
				onClick={(e) => e.stopPropagation()}
			>
				<Checkbox
					selected={selected}
					onChange={() => handleCheckboxChange(!selected)}
				/>
			</div>
		</>
	);

	return (
		<div
			className={`
		bg-card-muted-accent rounded-3xl p-2 flex items-center gap-2 relative
		transition-colors duration-200 w-full text-left
		cursor-pointer hover:bg-card-muted-accent/80
		`}
			onClick={handleClick}
		>
			<CardContent />
		</div>
	);
};
