import type { PropsWithChildren } from "react";
import React from "react";
import { TonBalance } from "@/widgets/TonBalance";

export function DefaultLayout({ children }: PropsWithChildren) {
	return (
		<React.Fragment>
			<div className="flex items-center justify-end container mb-3">
				<TonBalance />
			</div>
			{children}
		</React.Fragment>
	);
}
