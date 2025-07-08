import lottie, { AnimationItem } from "lottie-web";
import { useEffect, useRef } from "react";
import { cn } from "../utils/cn";

interface LottiePlayerProps {
	src: string | object; // можно .json URL или готовый JSON
	className?: string;
	loop?: boolean;
	autoplay?: boolean;
	renderer?: "svg" | "canvas" | "html";
	onComplete?: () => void;
	onLoopComplete?: () => void;
	onEnterFrame?: (frame: number) => void;
}

export function LottiePlayer({
	src,
	className,
	loop = true,
	autoplay = true,
	renderer = "svg",
	onComplete,
	onLoopComplete,
	onEnterFrame,
}: LottiePlayerProps) {
	const containerRef = useRef<HTMLDivElement>(null);
	const animRef = useRef<AnimationItem>(null);

	useEffect(() => {
		if (!containerRef.current) return;

		const isJsonUrl = typeof src === "string";
		const anim = lottie.loadAnimation({
			container: containerRef.current,
			renderer,
			loop,
			autoplay,
			...(isJsonUrl ? { path: src } : { animationData: src }),
		});

		anim.setSubframe(false);

		animRef.current = anim;

		if (onComplete) anim.addEventListener("complete", onComplete);
		if (onLoopComplete) anim.addEventListener("loopComplete", onLoopComplete);
		if (onEnterFrame) {
			anim.addEventListener("enterFrame", (e) =>
				// biome-ignore lint/suspicious/noExplicitAny: i cant explain
				onEnterFrame((e as any).currentTime),
			);
		}

		return () => {
			anim.destroy();
		};
	}, [src, loop, autoplay, renderer, onComplete, onLoopComplete, onEnterFrame]);

	return <div ref={containerRef} className={cn(className)} />;
}
