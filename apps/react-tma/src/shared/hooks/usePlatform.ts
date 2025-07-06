import { retrieveLaunchParams } from "@telegram-apps/sdk";

export function usePlatform() {
	const { tgWebAppPlatform: platform } = retrieveLaunchParams();

	return platform as "ios" | "android" | "macos" | "windows" | "linux";
}
