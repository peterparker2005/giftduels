import { GiftView } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { BiLogoTelegram } from "react-icons/bi";
import { Button } from "@/shared/ui/Button";
import { Icon } from "@/shared/ui/Icon/Icon";
import { formatThousands } from "@/shared/utils/formatThousands";

export const GiftCard = ({ gift }: { gift: GiftView }) => {
	return (
		<div key={gift.giftId?.value} className="p-2.5 rounded-3xl bg-card">
			<div className="relative w-full h-40 rounded-2xl cursor-default">
				<img
					src={`https://nft.fragment.com/gift/${gift.slug.toLowerCase()}.large.jpg`}
					alt={gift.title}
					draggable={false}
					className="w-full h-full object-cover rounded-2xl absolute top-0 left-0"
				/>
				<div className="absolute top-0 left-0 w-full h-full bg-gradient-to-b from-transparent from-60% to-black/60 rounded-2xl" />
				<div className="flex flex-col items-center gap-0 absolute bottom-2 left-0 mx-auto w-full">
					<p className="text-lg font-semibold">{gift.title}</p>
					<span className="text-[#cccccc] text-xs font-medium">
						#{gift.collectibleId}
					</span>
				</div>
			</div>
			<div className="mt-2 flex items-center gap-2">
				<Button
					variant={"primary"}
					className="w-full font-medium flex items-center justify-center gap-1 cursor-default"
					asChild
				>
					<div>
						<Icon icon="TON" className="w-5 h-5 shrink-0" />
						<span>{formatThousands(gift.price?.value)}</span>
					</div>
				</Button>
				<Button
					variant={"secondary"}
					className="rounded-full bg-card-accent w-9 h-9 flex items-center justify-center"
				>
					<BiLogoTelegram className="w-4 h-4 shrink-0" />
				</Button>
			</div>
		</div>
	);
};
