import { RouterProvider } from "@tanstack/react-router";
import { ErrorBoundary } from "./ErrorBoundary";
import { Providers } from "./providers";
import { router } from "./router";

function ErrorBoundaryError({ error }: { error: unknown }) {
	return (
		<div>
			<p>An unhandled error occurred:</p>
			<blockquote>
				<code>
					{error instanceof Error
						? error.message
						: typeof error === "string"
							? error
							: JSON.stringify(error)}
				</code>
			</blockquote>
		</div>
	);
}

function App() {
	return (
		<ErrorBoundary fallback={ErrorBoundaryError}>
			<Providers>
				<RouterProvider router={router} />
			</Providers>
		</ErrorBoundary>
	);
}

export default App;
