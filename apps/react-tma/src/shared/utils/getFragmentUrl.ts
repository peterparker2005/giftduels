type FragmentUrlType = "small" | "large" | "lottie";

export const getFragmentUrl = (
	slug: string,
	type: FragmentUrlType = "small",
) => {
	const extension = type === "lottie" ? "json" : "jpg";
	return `https://nft.fragment.com/gift/${slug.toLowerCase()}.${type}.${extension}`;
};
