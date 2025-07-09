export function setupShutdownHooks(callback: () => Promise<void>) {
	let isShuttingDown = false;

	const shutdown = async () => {
		if (isShuttingDown) return;
		isShuttingDown = true;
		console.log("Graceful shutdown initiated...");
		try {
			await callback();
		} catch (err) {
			console.error("Error during shutdown:", err);
		} finally {
			process.exit(0);
		}
	};

	process.once("SIGINT", shutdown);
	process.once("SIGTERM", shutdown);
}
