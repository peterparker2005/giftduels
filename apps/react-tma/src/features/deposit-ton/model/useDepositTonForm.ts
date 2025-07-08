import { useLayoutEffect, useRef } from "react";
import { useForm } from "react-hook-form";
import { useDepositTonMutation } from "@/shared/api/queries/useDepositTonMutation";
import { useTonTransaction } from "./useTonTransaction";

export interface DepositTonFormData {
	amount: string;
}

interface UseDepositTonFormProps {
	onCloseDialog?: () => void;
}

export const useDepositTonForm = (props?: UseDepositTonFormProps) => {
	const depositTonMutation = useDepositTonMutation();
	const { sendTransaction } = useTonTransaction();
	const { register, setValue, watch, handleSubmit, reset } =
		useForm<DepositTonFormData>({
			defaultValues: { amount: "0" },
		});

	const amount = watch("amount");
	const inputRef = useRef<HTMLInputElement>(null);
	const spanRef = useRef<HTMLSpanElement>(null);

	// Dynamic width effect
	useLayoutEffect(() => {
		if (inputRef.current && spanRef.current) {
			const text = amount || "0";
			spanRef.current.textContent = text;
			const width = spanRef.current.offsetWidth + 16;
			inputRef.current.style.width = `${width}px`;
		}
	}, [amount]);

	const onInput = (e: React.ChangeEvent<HTMLInputElement>) => {
		let raw = e.target.value
			.replace(/,/g, ".")
			.replace(/[бю]/g, ".")
			.replace(/[^0-9.]/g, "");

		const idx = raw.indexOf(".");
		if (idx >= 0) {
			raw = raw.slice(0, idx + 1) + raw.slice(idx + 1).replace(/\./g, "");
		}

		// 3) Спец-кейс: если ввели только точку
		if (raw === ".") {
			setValue("amount", "0.", { shouldValidate: false });
			return;
		}

		// 4) Спец-кейс «только нули»: 0 → 0.0 → 0.01
		if (/^0+$/.test(raw)) {
			if (raw.length === 1) {
				setValue("amount", "0", { shouldValidate: false });
			} else if (raw.length === 2) {
				setValue("amount", "0.0", { shouldValidate: false });
			} else {
				setValue("amount", "0.01", { shouldValidate: false });
			}
			return;
		}

		const [intRaw, fracRaw = ""] = raw.split(".");
		// уходим от лидирующих нулей, но чтобы при «0» оставался один ноль
		const intPart = intRaw.replace(/^0+(?!$)/, "");

		// Проверка на превышение длины - если больше 5 символов, ставим максимум
		if (intPart.length > 5) {
			setValue("amount", "99999", { shouldValidate: false });
			return;
		}

		const fracPart = fracRaw.slice(0, 2);

		// 6) Собираем
		let result = intPart;
		if (raw.includes(".")) {
			result += `.${fracPart}`;
		}

		// 7) Проверка на 0.00 - устанавливаем 0.01
		if (result === "0.00") {
			setValue("amount", "0.01", { shouldValidate: false });
			return;
		}

		// 8) Проверка минимального шага 0.01
		const num = parseFloat(result);
		if (!Number.isNaN(num) && num > 0 && num < 0.01) {
			setValue("amount", "0.01", { shouldValidate: false });
			return;
		}

		// 9) Сохраняем
		setValue("amount", result, { shouldValidate: false });
	};

	const onSubmit = (data: DepositTonFormData) => {
		depositTonMutation.mutate(data.amount, {
			onSuccess: async (depositData) => {
				console.log("Deposit created:", depositData);

				try {
					// Close dialog and reset form before opening TonConnect
					props?.onCloseDialog?.();
					reset();

					// Send TON transaction using TonConnect
					const result = await sendTransaction({
						treasuryAddress: depositData.treasuryAddress,
						nanoTonAmount: depositData.nanoTonAmount,
						payload: depositData.payload,
					});

					console.log("Transaction sent successfully:", result);
				} catch (error) {
					console.error("Failed to send transaction:", error);
					// Handle transaction error (e.g., show error message to user)
				}
			},
			onError: (error) => {
				console.error("Failed to create deposit:", error);
			},
		});
	};

	return {
		amount,
		inputRef,
		spanRef,
		register,
		onInput,
		handleSubmit: handleSubmit(onSubmit),
		reset,
	};
};
