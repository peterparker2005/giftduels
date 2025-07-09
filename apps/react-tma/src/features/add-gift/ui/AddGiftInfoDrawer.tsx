import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { openTelegramLink } from "@telegram-apps/sdk";
import { Button } from "@/shared/ui/Button";
import {
	Drawer,
	DrawerContent,
	DrawerTitle,
	DrawerTrigger,
} from "@/shared/ui/Drawer";

interface AddGiftInfoDrawerProps {
	children: React.ReactNode;
}

export const AddGiftInfoDrawer = ({ children }: AddGiftInfoDrawerProps) => {
	const handleOpenTelegramLink = (
		e: React.MouseEvent<HTMLAnchorElement | HTMLButtonElement>,
	) => {
		e.preventDefault();
		openTelegramLink("https://t.me/peterparkish");
	};

	return (
		<Drawer>
			<DrawerTrigger asChild>{children}</DrawerTrigger>
			<DrawerContent className="">
				<img
					src="/add-gift-info.png"
					alt="Add gift info"
					draggable={false}
					className="w-full h-[200px] object-cover select-none"
				/>
				<VisuallyHidden>
					<DrawerTitle>How to add gifts?</DrawerTitle>
				</VisuallyHidden>
				<section className="flex flex-col text-lg items-start px-4 py-5 gap-5">
					<div className="flex items-start gap-2">
						<span className="rounded-full bg-card-accent w-6 h-6 mt-0.5 flex items-center justify-center font-semibold text-sm">
							1
						</span>

						<div className="flex flex-col">
							<p className="font-semibold">
								Go to the bot using the link below
							</p>
							<p className="text-card-muted-foreground">
								Our official bot name —{" "}
								<a
									href="https://t.me/@gifts_to_duels_bot"
									onClick={handleOpenTelegramLink}
									className="text-primary"
								>
									@GiftDuels
								</a>
							</p>
						</div>
					</div>

					<div className="flex items-start gap-2">
						<span className="rounded-full bg-card-accent w-6 h-6 mt-0.5 flex items-center justify-center font-semibold text-sm">
							2
						</span>

						<div className="flex flex-col">
							<p className="font-semibold">Send your NFT gift to the bot</p>
							<p className="text-card-muted-foreground">
								No matter how many gifts you send
							</p>
						</div>
					</div>

					<div className="flex items-start gap-2">
						<span className="rounded-full bg-card-accent w-6 h-6 mt-0.5 flex items-center justify-center font-semibold text-sm">
							3
						</span>

						<div className="flex flex-col">
							<p className="font-semibold">Gifts will appear in the profile</p>
							<p className="text-card-muted-foreground">
								Processing might take 1-2 minutes
							</p>
						</div>
					</div>
				</section>
				<div className="px-4 mb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
					<Button
						variant={"default"}
						className="w-full py-3.5"
						onClick={handleOpenTelegramLink}
					>
						Add gift
					</Button>
				</div>
			</DrawerContent>
		</Drawer>
	);
};
