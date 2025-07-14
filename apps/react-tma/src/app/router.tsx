import {
	createRootRoute,
	createRoute,
	createRouter,
	Outlet,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { DefaultLayout } from "@/layouts/DefaultLayout";
import { RootLayout } from "@/layouts/RootLayout";
import Balance from "@/pages/Balance";
import Duel from "@/pages/Duel";
import Duels from "@/pages/Duels";
import Home from "@/pages/Home";
import Profile from "@/pages/Profile";
import { useMobile } from "@/shared/hooks/useMobile";
import { config } from "./config";

const rootRoute = createRootRoute({
	component: () => {
		const mobile = useMobile();
		return (
			<RootLayout>
				<Outlet />
				{config.isDev && !mobile && (
					<TanStackRouterDevtools position="bottom-right" />
				)}
			</RootLayout>
		);
	},
});

const defaultLayoutRoute = createRoute({
	getParentRoute: () => rootRoute,
	id: "default-layout",
	component: () => {
		return (
			<DefaultLayout>
				<Outlet />
			</DefaultLayout>
		);
	},
});

const duelsRoute = createRoute({
	getParentRoute: () => defaultLayoutRoute,
	path: "/duels",
	component: Duels,
});

const duelRoute = createRoute({
	getParentRoute: () => defaultLayoutRoute,
	path: "/duel/$duelId",
	component: Duel,
});

const indexRoute = createRoute({
	getParentRoute: () => defaultLayoutRoute,
	path: "/",
	component: Home,
});

const profileRoute = createRoute({
	getParentRoute: () => defaultLayoutRoute,
	path: "/profile",
	component: Profile,
});

const balanceRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/balance",
	component: Balance,
});

const defaultLayoutWithChildren = defaultLayoutRoute.addChildren([
	indexRoute,
	profileRoute,
	duelsRoute,
	duelRoute,
]);

const routeTree = rootRoute.addChildren([
	defaultLayoutWithChildren,
	balanceRoute,
]);

export const router = createRouter({
	routeTree,
	context: {},
	defaultPreload: "intent",
	scrollRestoration: true,
	defaultStructuralSharing: true,
	defaultPreloadStaleTime: 0,
});

declare module "@tanstack/react-router" {
	interface Register {
		router: typeof router;
	}
}
