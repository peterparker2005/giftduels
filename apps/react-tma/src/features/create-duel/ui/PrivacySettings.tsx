import { useFormContext } from "react-hook-form";
import { cn } from "@/shared/utils/cn";

export function PrivacySettings() {
	const { register, watch } = useFormContext();
	const privacy = watch("privacy");

	return (
		<div className="space-y-3">
			<h3 className="text-lg font-semibold">Privacy</h3>
			<div className="space-y-2">
				<label className="flex items-center space-x-3 cursor-pointer">
					<input
						type="radio"
						value="public"
						{...register("privacy")}
						className="sr-only"
					/>
					<div
						className={cn(
							"w-4 h-4 rounded-full border-2 flex items-center justify-center",
							privacy === "public"
								? "border-primary bg-primary"
								: "border-muted-foreground",
						)}
					>
						{privacy === "public" && (
							<div className="w-2 h-2 rounded-full bg-white" />
						)}
					</div>
					<span className="text-sm">Public</span>
				</label>

				<label className="flex items-center space-x-3 cursor-pointer">
					<input
						type="radio"
						value="private"
						{...register("privacy")}
						className="sr-only"
					/>
					<div
						className={cn(
							"w-4 h-4 rounded-full border-2 flex items-center justify-center",
							privacy === "private"
								? "border-primary bg-primary"
								: "border-muted-foreground",
						)}
					>
						{privacy === "private" && (
							<div className="w-2 h-2 rounded-full bg-white" />
						)}
					</div>
					<span className="text-sm">Private</span>
				</label>
			</div>
		</div>
	);
}
