import {
	createRootRoute,
	createRoute,
	createRouter,
	Outlet,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { RootLayout } from "@/layouts/RootLayout";
import Balance from "@/pages/Balance";
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

const indexRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/",
	component: Home,
});

const profileRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/profile",
	component: Profile,
});

const balanceRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/balance",
	component: Balance,
});

const routeTree = rootRoute.addChildren([
	indexRoute,
	profileRoute,
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
