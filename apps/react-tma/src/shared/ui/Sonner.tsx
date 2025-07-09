import { Toaster as Sonner, ToasterProps } from "sonner";

const Toaster = ({ ...props }: ToasterProps) => {
	return (
		<Sonner
			theme="dark"
			className="toaster group"
			style={
				{
					"--normal-bg": "var(--color-card-accent)",
					"--normal-text": "var(--color-foreground)",
					"--normal-border": "var(--color-card-accent)",
				} as React.CSSProperties
			}
			{...props}
		/>
	);
};

export { Toaster };
