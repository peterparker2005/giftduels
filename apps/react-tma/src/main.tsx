import "./app/sentry";
import "./app/styles/main.css";

import { retrieveLaunchParams } from "@telegram-apps/sdk-react";
import { StrictMode } from "react";
import ReactDOM from "react-dom/client";
import { AppLoader } from "./app/AppLoader.tsx";
import { init } from "./app/init.ts";
import reportWebVitals from "./reportWebVitals.ts";

// biome-ignore lint/style/noNonNullAssertion: idgaf
const root = ReactDOM.createRoot(document.getElementById("root")!);

(async () => {
	try {
		const launchParams = retrieveLaunchParams();
		const { tgWebAppPlatform: platform } = launchParams;
		const debug =
			(launchParams.tgWebAppStartParam || "").includes("platformer_debug") ||
			import.meta.env.DEV;

		await init({
			debug,
			eruda: debug && ["ios", "android"].includes(platform),
			mockForMacOS: platform === "macos",
		}).then(() => {
			root.render(
				<StrictMode>
					<AppLoader />
				</StrictMode>,
			);
		});
	} catch (e) {
		root.render(<div>Error: {JSON.stringify(e)}</div>);
	}
})();

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
