import { useEffect, useState } from "react";
import { DiceLoader } from "@/shared/ui/DiceLoader";
import App from "./App";

export function AppLoader() {
	const [ready, setReady] = useState(false);
	const [error, _setError] = useState<unknown>(null);

	useEffect(() => {
		const timeout = setTimeout(() => {
			setReady(true);
		}, 10 * 1000);

		return () => clearTimeout(timeout); // чистка на размонтировании
	}, []);

	if (error) {
		return (
			<div style={{ padding: 24 }}>
				<h2>Something went wrong</h2>
				<pre>{JSON.stringify(error, null, 2)}</pre>
			</div>
		);
	}

	if (!ready) {
		return <DiceLoader />;
	}

	return <App />;
}
