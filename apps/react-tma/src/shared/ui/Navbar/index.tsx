import { useLocation } from "@tanstack/react-router";
import { BiSolidDice3, BiSolidHome, BiSolidUserCircle } from "react-icons/bi";
import { NavItem } from "./NavItem";

export const Navbar = () => {
	const { pathname } = useLocation();
	const isActive = (href: string) => pathname === href;
	return (
		<div className="fixed left-0 bottom-0 z-20 w-full blur-background">
			<div className="pt-3 w-full pb-[calc(var(--tg-viewport-safe-area-inset-bottom)+10px)]">
				<div className="grid grid-cols-3 w-full place-items-center">
					<NavItem href="/" isActive={isActive("/")}>
						<BiSolidHome className="size-6" />
						<span className="text-xs font-medium">Home</span>
					</NavItem>
					<NavItem href="/duels" isActive={isActive("/duels")}>
						<BiSolidDice3 className="size-6" />
						<span className="text-xs font-medium">Duels</span>
					</NavItem>
					<NavItem href="/profile" isActive={isActive("/profile")}>
						<BiSolidUserCircle className="size-6" />
						<span className="text-xs font-medium">Profile</span>
					</NavItem>
				</div>
			</div>
		</div>
	);
};
