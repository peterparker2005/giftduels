import { useFormContext } from "react-hook-form";
import { cn } from "@/shared/utils/cn";

export function PlayerCountSettings() {
	const { watch, setValue } = useFormContext();
	const players = watch("players");

	return (
		<div className="space-y-3">
			<h3 className="text-lg font-semibold">Number of Players</h3>
			<div className="flex items-center gap-2">
				{[2, 3, 4].map((num) => (
					<button
						key={num}
						type="button"
						onClick={() => setValue("players", num)}
						className={cn(
							"px-4 py-2 rounded-3xl bg-card-muted-accent text-center transition-colors",
							players === num ? "bg-primary" : "",
						)}
					>
						<span className="text-lg font-semibold">{num}</span>
					</button>
				))}
			</div>
		</div>
	);
}
