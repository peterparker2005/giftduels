import type { PropsWithChildren } from "react";
import { router } from "@/app/router";
import { cn } from "@/shared/utils/cn";

interface NavItemProps extends PropsWithChildren {
	href: string;
	isActive: boolean;
}
export const NavItem = ({ children, href, isActive }: NavItemProps) => {
	return (
		<button
			type="button"
			className={cn(
				"flex flex-col items-center gap-1 px-3 cursor-pointer text-primary-foreground opacity-45",
				isActive && "opacity-100 text-primary",
			)}
			onClick={() => {
				router.navigate({ to: href });
			}}
		>
			{children}
		</button>
	);
};
