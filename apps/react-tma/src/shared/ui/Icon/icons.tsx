import type { FC, LazyExoticComponent, SVGProps } from "react";

import { lazy as _lazy } from "react";

function lazy<T extends FC<SVGProps<SVGSVGElement>>>(
	importFn: () => Promise<{ default: T }>,
): LazyExoticComponent<T> {
	return _lazy(importFn);
}

export const icons = {
	TON: lazy(() => import("@/assets/ton.svg?react")),
};

export type IconName = keyof typeof icons;
