import { useEffect, useRef } from "react";

interface UseIntersectionObserverOptions {
	onIntersect: () => void;
	threshold?: number;
	rootMargin?: string;
	enabled?: boolean;
}

export function useIntersectionObserver({
	onIntersect,
	threshold = 0.1,
	rootMargin = "0px",
	enabled = true,
}: UseIntersectionObserverOptions) {
	const targetRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		if (!enabled) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting) {
					onIntersect();
				}
			},
			{
				threshold,
				rootMargin,
			},
		);

		if (targetRef.current) {
			observer.observe(targetRef.current);
		}

		return () => observer.disconnect();
	}, [onIntersect, threshold, rootMargin, enabled]);

	return targetRef;
}
