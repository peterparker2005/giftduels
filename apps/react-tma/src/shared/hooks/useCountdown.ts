import { useEffect, useState } from "react";

export function useCountdown(deadline?: Date) {
	const calculate = () => {
		if (!deadline) return { days: 0, hours: 0, minutes: 0, seconds: 0 };
		const now = Date.now();
		const diff = deadline.getTime() - now;
		const seconds = Math.max(0, Math.floor(diff / 1000) % 60);
		const minutes = Math.max(0, Math.floor(diff / 1000 / 60) % 60);
		const hours = Math.max(0, Math.floor(diff / (1000 * 60 * 60)) % 24);
		const days = Math.max(0, Math.floor(diff / (1000 * 60 * 60 * 24)));
		return { days, hours, minutes, seconds };
	};

	const [time, setTime] = useState(calculate);

	// biome-ignore lint/correctness/useExhaustiveDependencies: calculate re-renders every second
	useEffect(() => {
		if (!deadline) return;
		const timer = setInterval(() => {
			setTime(calculate());
		}, 1000);
		return () => clearInterval(timer);
	}, [deadline]);

	return time;
}
