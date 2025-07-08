import type { PropsWithChildren } from "react";
import { DiceBackground } from "@/shared/ui/DiceBackground";
import { Navbar } from "@/shared/ui/Navbar";
import { TonBalance } from "@/widgets/TonBalance";

export function RootLayout({ children }: PropsWithChildren) {
	return (
		<>
			<DiceBackground />
			<div className="relative text-sm pt-[calc(var(--tg-viewport-safe-area-inset-top)+30px)] pb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)] overflow-y-auto w-full overflow-x-hidden scrollable h-screen flex flex-col">
				<div className="relative flex flex-col flex-1">
					<div className="mb-3 relative pt-4">
						<div className="flex items-center justify-end container">
							<TonBalance />
						</div>

						{children}
					</div>
					<Navbar />
				</div>
			</div>
		</>
	);
}
