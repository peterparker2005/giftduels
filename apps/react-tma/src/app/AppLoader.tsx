import { useEffect } from "react";
import { useAuthStore } from "@/features/auth/model/store";
import { useAuthorizeMutation } from "@/shared/api/queries/useAuthorizeMutation";
import { DiceLoader } from "@/shared/ui/DiceLoader";
import App from "./App";

export function AppLoader() {
	const { mutate, isPending, isError, error, data } = useAuthorizeMutation();

	useEffect(() => {
		mutate(undefined, {
			onSuccess: ({ token }) => {
				useAuthStore.getState().setToken(token);
			},
		});
	}, [mutate]);

	if (isPending) {
		return <DiceLoader />;
	}

	if (isError) {
		return (
			<div style={{ padding: 24 }}>
				<h2>Что-то пошло не так</h2>
				<pre>{JSON.stringify(error, null, 2)}</pre>
			</div>
		);
	}

	if (data) {
		return <App />;
	}

	return null;
}
