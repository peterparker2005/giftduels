import * as Slot from "@radix-ui/react-slot";
import { cva, VariantProps } from "class-variance-authority";
import { ButtonHTMLAttributes } from "react";
import { cn } from "../utils/cn";

interface ButtonProps
	extends ButtonHTMLAttributes<HTMLButtonElement>,
		VariantProps<typeof buttonVariants> {
	asChild?: boolean;
}

const buttonVariants = cva(
	"rounded-3xl px-4 py-2 font-medium disabled:cursor-default",
	{
		variants: {
			variant: {
				default: "bg-white text-background disabled:bg-white/50",
				primary: "bg-primary text-primary-foreground",
				secondary: "bg-card text-card-accent-foreground",
			},
		},
		defaultVariants: {
			variant: "default",
		},
	},
);

export const Button = ({
	children,
	className,
	variant,
	asChild,
	...props
}: ButtonProps) => {
	const Comp = asChild ? Slot.Root : "button";
	return (
		<Comp className={cn(buttonVariants({ variant }), className)} {...props}>
			{children}
		</Comp>
	);
};
