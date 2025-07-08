import { RouterProvider } from "@tanstack/react-router";
import { ErrorBoundary } from "./ErrorBoundary";
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
			<RouterProvider router={router} />
		</ErrorBoundary>
	);
}

export default App;
