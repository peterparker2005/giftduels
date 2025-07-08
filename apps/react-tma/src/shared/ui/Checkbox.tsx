import { BiCheck } from "react-icons/bi";
import { cn } from "../utils/cn";

interface CheckboxProps {
	selected?: boolean;
	onChange?: (selected: boolean) => void;
	disabled?: boolean;
	className?: string;
}

export const Checkbox = ({
	selected = false,
	onChange,
	disabled = false,
	className = "",
}: CheckboxProps) => {
	const handleClick = () => {
		if (!disabled && onChange) {
			onChange(!selected);
		}
	};

	return (
		<button
			type="button"
			onClick={handleClick}
			disabled={disabled}
			className={cn(
				"w-5 h-5 rounded-md border-1 flex items-center justify-center transition-all duration-200 ease-in-out",
				selected
					? "bg-primary border-primary"
					: "bg-transparent border-muted-foreground",
				disabled
					? "opacity-50 cursor-not-allowed"
					: "cursor-pointer hover:border-primary",
				className,
			)}
		>
			{selected && <BiCheck />}
		</button>
	);
};
