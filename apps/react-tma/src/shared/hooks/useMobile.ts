import { usePlatform } from "./usePlatform";

export function useMobile() {
	const platform = usePlatform();

	return ["ios", "android"].includes(platform);
}
