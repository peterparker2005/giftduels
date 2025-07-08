import { HTMLAttributes, Suspense, useMemo } from "react";

import { icons } from "./icons";

export type IconName = keyof typeof icons;

// Props can vary from project to project, some will require to have some specific variant passed for styling,
// others will extend base css classes with custom prop class etc

interface Props extends HTMLAttributes<HTMLDivElement> {
	icon: IconName;
	className?: string;
	lazy?: boolean;
}

/**
 *
 * @param icon string key icon name
 * @param className string classes for styling
 * @param rotate optional number rotation of the icon
 * @returns Icon react component
 */
export const Icon = ({ icon, className, lazy = true, ...rest }: Props) => {
	const SvgIcon = useMemo(() => icons[icon], [icon]);

	if (!SvgIcon) return null;

	return (
		<div
			className={className}
			aria-label={icon}
			role="img"
			style={{
				display: "flex",
				justifyContent: "center",
				alignItems: "center",
			}}
			{...rest}
		>
			{lazy ? (
				<Suspense fallback={null}>
					<SvgIcon style={{ width: "100%", height: "100%" }} />
				</Suspense>
			) : (
				<SvgIcon style={{ width: "100%", height: "100%" }} />
			)}
		</div>
	);
};
