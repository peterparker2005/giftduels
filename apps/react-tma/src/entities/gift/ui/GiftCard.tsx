import {
	GiftStatus,
	GiftView,
} from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import { useNavigate } from "@tanstack/react-router";
import { BiLogoTelegram } from "react-icons/bi";
import { IoLogoGameControllerB } from "react-icons/io";
import { Button } from "@/shared/ui/Button";
import { Icon } from "@/shared/ui/Icon/Icon";
import { cn } from "@/shared/utils/cn";
import { getFragmentUrl } from "@/shared/utils/getFragmentUrl";
import { GiftDetailsDrawer } from "./GiftDetailsDrawer";

interface GiftCardProps {
	gift: GiftView;
}

export const GiftCard = ({ gift }: GiftCardProps) => {
	const navigate = useNavigate();

	return (
		<GiftDetailsDrawer gift={gift}>
			<div
				key={gift.giftId?.value}
				className={cn(
					"p-2.5 rounded-3xl bg-card cursor-pointer hover:bg-card/80 transition-colors relative",
				)}
			>
				<div className="relative w-full h-40 rounded-2xl">
					<img
						// src={`https://nft.fragment.com/gift/${gift.slug.toLowerCase()}.large.jpg`}
						src={getFragmentUrl(gift.slug, "large")}
						alt={gift.title}
						draggable={false}
						className="w-full h-full object-cover rounded-2xl absolute top-0 left-0"
					/>
					<div className="absolute top-0 left-0 w-full h-full bg-gradient-to-b from-transparent from-60% to-black/60 rounded-2xl" />
					<div className="flex flex-col items-center gap-1 absolute bottom-2 left-0 mx-auto w-full">
						<p className="text-lg text-center leading-[100%] font-semibold">
							{gift.title}
						</p>
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
							<span>{gift.price?.value}</span>
						</div>
					</Button>
					<Button
						variant={"secondary"}
						className="rounded-full bg-card-accent w-9 h-9 flex items-center justify-center"
						onClick={() => {
							if (!gift.relatedDuelId?.value) return;
							navigate({
								to: "/duel/$duelId",
								params: { duelId: gift.relatedDuelId?.value },
							});
						}}
					>
						{gift.status === GiftStatus.IN_GAME ? (
							<IoLogoGameControllerB className="w-4 h-4 shrink-0" />
						) : (
							<BiLogoTelegram className="w-4 h-4 shrink-0" />
						)}
					</Button>
				</div>
			</div>
		</GiftDetailsDrawer>
	);
};
