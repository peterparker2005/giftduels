import { cn } from "@/shared/utils/cn";

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
	className?: string;
	baseColor?: string; // CSS-переменная --base-color
	highlightColor?: string; // CSS-переменная --highlight-color
	duration?: number; // --animation-duration в секундах
	direction?: "ltr" | "rtl"; // --animation-direction
	enableAnimation?: boolean; // при false — выключает ::after
}

export const Skeleton = ({
	className,
	baseColor,
	highlightColor,
	duration,
	direction = "ltr",
	enableAnimation = true,
	...props
}: SkeletonProps) => {
	const vars: React.CSSProperties & Record<string, string> = {};
	vars["--base-color"] = baseColor || "hsl(0, 0%, 12%)";
	vars["--highlight-color"] = highlightColor || "hsl(0, 0%, 18%)";
	if (duration != null) vars["--animation-duration"] = `${duration}s`;
	if (direction === "rtl") vars["--animation-direction"] = "reverse";
	if (!enableAnimation) vars["--pseudo-element-display"] = "none";

	return (
		<div
			className={cn(
				"w-full h-4 bg-[var(--base-color)] rounded-3xl inline-flex relative select-none overflow-hidden",
				'after:content-[""] after:left-0 after:right-0 after:top-0 after:absolute',
				"after:bg-linear-90 after:bg-no-repeat after:h-full",
				"after:from-[var(--base-color)] after:via-[var(--highlight-color)] after:to-[var(--base-color)]",
				"after:animate-shimmer after:-translate-x-full",
				className,
			)}
			style={{ ...vars }}
			{...props}
		/>
	);
};
