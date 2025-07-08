import { useState } from "react";
import { Button } from "@/shared/ui/Button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogTitle,
	DialogTrigger,
} from "@/shared/ui/Dialog";
import { useDepositTonForm } from "../model/useDepositTonForm";

interface DepositTonDialogProps {
	children: React.ReactNode;
}

export const DepositTonDialog = ({ children }: DepositTonDialogProps) => {
	const [open, setOpen] = useState(false);

	const closeDialog = () => {
		setOpen(false);
	};

	const form = useDepositTonForm({ onCloseDialog: closeDialog });

	const handleOpenChange = (openState: boolean) => {
		setOpen(openState);
		if (openState) {
			// Сбрасываем форму при открытии диалога для корректной ширины
			form.reset();
		}
	};

	return (
		<Dialog open={open} onOpenChange={handleOpenChange}>
			<DialogTrigger asChild>{children}</DialogTrigger>
			<DialogContent className="container">
				<DialogTitle className="text-xl">Deposit</DialogTitle>
				<DialogDescription className="text-sm">
					Enter the amount of TON you want to deposit
				</DialogDescription>

				<form
					onSubmit={form.handleSubmit}
					className="flex flex-col h-full items-center justify-center"
				>
					<div className="flex items-center gap-2">
						<input
							type="text"
							inputMode="decimal"
							className="text-5xl font-semibold bg-transparent outline-none text-center"
							{...form.register("amount")}
							value={form.amount}
							onInput={form.onInput}
							ref={form.inputRef}
							autoComplete="off"
							spellCheck={false}
							maxLength={8} // 5 + 1 + 2
							size={Math.max(form.amount.length, 1)} // <-- вот этот атрибут
						/>
						<span className="text-muted-foreground text-2xl font-semibold">
							TON
						</span>
						{/* Hidden span for dynamic width calculation */}
						<span
							ref={form.spanRef}
							className="invisible absolute whitespace-pre text-5xl font-semibold pointer-events-none"
							style={{ left: -9999, top: -9999 }}
						/>
					</div>
					<Button type="submit" className="mt-8 w-full py-3" variant="primary">
						Deposit
					</Button>
				</form>
			</DialogContent>
		</Dialog>
	);
};
