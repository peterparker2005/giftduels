import {
	Data,
	DotLottieReact,
	DotLottieReactProps,
} from "@lottiefiles/dotlottie-react";
import { cn } from "../utils/cn";

interface LottiePlayerProps extends Omit<DotLottieReactProps, "src"> {
	src: string | Data;
	className?: string;
}

export function LottiePlayer({ src, className, ...props }: LottiePlayerProps) {
	return (
		<div className={cn(className)}>
			<DotLottieReact
				data={typeof src === "string" ? undefined : src}
				src={typeof src === "string" ? src : undefined}
				useFrameInterpolation={false}
				{...props}
			/>
		</div>
	);
}
