import { SendTransactionRequest, useTonConnectUI } from "@tonconnect/ui-react";

interface SendTransactionParams {
	treasuryAddress: string;
	nanoTonAmount: bigint;
	payload: string;
}

export const useTonTransaction = () => {
	const [tonConnectUI] = useTonConnectUI();

	const sendTransaction = async (params: SendTransactionParams) => {
		const { treasuryAddress, nanoTonAmount, payload } = params;

		try {
			const transaction: SendTransactionRequest = {
				validUntil: Math.floor(Date.now() / 1000) + 60, // Valid for 60 seconds
				messages: [
					{
						address: treasuryAddress,
						amount: nanoTonAmount.toString(),
						payload: payload,
					},
				],
			};

			const result = await tonConnectUI.sendTransaction(transaction);
			console.log("Transaction sent:", result);
			return result;
		} catch (error) {
			console.error("Failed to send transaction:", error);
			throw error;
		}
	};

	return {
		sendTransaction,
	};
};
