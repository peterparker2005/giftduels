import type { FC, LazyExoticComponent, SVGProps } from "react";

import { lazy as _lazy } from "react";

function lazy<T extends FC<SVGProps<SVGSVGElement>>>(
	importFn: () => Promise<{ default: T }>,
): LazyExoticComponent<T> {
	return _lazy(importFn);
}

export const icons = {
	TON: lazy(() => import("@/assets/ton.svg?react")),
	Star: lazy(() => import("@/assets/star.svg?react")),
	Plus: lazy(() => import("@/assets/plus.svg?react")),
};

export type IconName = keyof typeof icons;
