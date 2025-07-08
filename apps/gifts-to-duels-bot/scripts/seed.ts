import { Userbot } from "@/core/userbot";

async function seed() {
	const userbot = new Userbot();

	await userbot.start();

	const { gifts } = await userbot.getUserGifts({
		limit: 1,
		user: "@GiftsToPortals",
	});

	console.log(gifts);

	await userbot.close();
}

seed().catch((err) => {
	console.error("[FATAL] Unhandled exception:", err);
	process.exit(1);
});
