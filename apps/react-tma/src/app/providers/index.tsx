import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { queryClient } from "@/shared/api/query-client";
import { useMobile } from "@/shared/hooks/useMobile";
import { Toaster } from "@/shared/ui/Sonner";
import { config } from "../config";
import { TonWalletProvider } from "./TonWalletProvider";

export function Providers({ children }: { children: React.ReactNode }) {
	const mobile = useMobile();
	return (
		<QueryClientProvider client={queryClient}>
			<TonWalletProvider>
				<Toaster />
				{children}
			</TonWalletProvider>
			{!mobile && config.isDev && (
				<ReactQueryDevtools buttonPosition="top-right" />
			)}
		</QueryClientProvider>
	);
}
