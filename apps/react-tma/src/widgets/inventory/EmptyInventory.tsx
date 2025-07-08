import { openTelegramLink } from "@telegram-apps/sdk";
import { AddGiftInfoDrawer } from "@/features/add-gift/ui/AddGiftInfoDrawer";
import { Button } from "@/shared/ui/Button";
import { LottiePlayer } from "@/shared/ui/LottiePlayer";

export const EmptyInventory = () => {
	const handleOpenTelegramLink = (e: React.MouseEvent<HTMLAnchorElement>) => {
		e.preventDefault();
		openTelegramLink("https://t.me/peterparkish");
	};
	return (
		<section className="flex flex-col items-center gap-2 mt-10">
			<LottiePlayer
				src="/lottie/sad-duck.json"
				autoplay
				loop
				className="w-24 h-24"
			/>
			<div className="space-y-2 text-center">
				<p className="font-semibold text-2xl">No gifts yet</p>
				<p className="text-muted-foreground text-sm">
					Send your gifts to{" "}
					<a
						href="https://t.me/giftduels"
						className="text-primary"
						onClick={handleOpenTelegramLink}
					>
						@giftduels
					</a>
				</p>
			</div>
			<AddGiftInfoDrawer>
				<Button className="px-10 font-semibold mt-4">How can I do it?</Button>
			</AddGiftInfoDrawer>
		</section>
	);
};
