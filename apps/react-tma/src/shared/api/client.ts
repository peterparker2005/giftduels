import { giftduels } from "@giftduels/protobuf-ts";
import { retrieveRawInitData } from "@telegram-apps/sdk";
import axios, {
	AxiosError,
	AxiosInstance,
	AxiosRequestConfig,
	AxiosResponse,
} from "axios";

import { config } from "@/app/config";
import { useAuthStore } from "@/features/auth/model/store";
import { identityClient } from "./identity/v1";

export interface IHttpClient {
	get: AxiosInstance["get"];
	post: AxiosInstance["post"];
	put: AxiosInstance["put"];
	delete: AxiosInstance["delete"];
	patch: AxiosInstance["patch"];
	head: AxiosInstance["head"];
	options: AxiosInstance["options"];
}

export class HTTPClient implements IHttpClient {
	private readonly instance: AxiosInstance;
	private isRefreshing = false;
	private failedQueue: Array<{
		resolve: (value: AxiosResponse) => void;
		reject: (err: unknown) => void;
	}> = [];

	constructor(prefix = "") {
		this.instance = axios.create({
			baseURL: `${config.apiUrl}${prefix}`,
			withCredentials: true,
			headers: { "Content-Type": "application/json" },
			timeout: 10_000,
		});
		this.setupInterceptors();
	}

	/* ------------------------------------------------------------------ */
	/*  Axios-совместимые методы                                          */
	/* ------------------------------------------------------------------ */

	get: IHttpClient["get"] = (url, config?) => this.instance.get(url, config);
	post: IHttpClient["post"] = (url, data?, config?) =>
		this.instance.post(url, data, config);
	put: IHttpClient["put"] = (url, data?, config?) =>
		this.instance.put(url, data, config);
	delete: IHttpClient["delete"] = (url, config?) =>
		this.instance.delete(url, config);
	patch: IHttpClient["patch"] = (url, data?, config?) =>
		this.instance.patch(url, data, config);
	head: IHttpClient["head"] = (url, config?) => this.instance.head(url, config);
	options: IHttpClient["options"] = (url, config?) =>
		this.instance.options(url, config);

	/* ------------------------------------------------------------------ */

	private setupInterceptors() {
		/* ----- request: auth-header ------------------------------------- */
		this.instance.interceptors.request.use((cfg) => {
			const token = useAuthStore.getState().token;
			if (token) {
				cfg.headers.Authorization = `Bearer ${token}`;
			}
			return cfg;
		});

		/* ----- response: refresh-token, retry queue --------------------- */
		this.instance.interceptors.response.use(
			(res) => res,
			async (err: AxiosError) => {
				const original = err.config as AxiosRequestConfig & { _retry?: true };

				if (err.response?.status === 401 && !original._retry) {
					original._retry = true;

					if (this.isRefreshing) {
						return new Promise((resolve, reject) =>
							this.failedQueue.push({ resolve, reject }),
						).then(() => this.instance(original));
					}

					this.isRefreshing = true;
					try {
						const initData = retrieveRawInitData();
						if (!initData) throw new Error("no initData");

						const { token } = await identityClient.authorize({
							$type: giftduels.identity.v1.AuthorizeRequest.$type,
							initData,
						});
						useAuthStore.getState().setToken(token);

						// прокидываем новый токен в исходный запрос
						if (original.headers) {
							original.headers.Authorization = `Bearer ${token}`;
						} else {
							original.headers = { Authorization: `Bearer ${token}` };
						}

						const response = await this.instance(original);
						this.failedQueue.forEach((p) => p.resolve(response));
						return response;
					} catch (e) {
						this.failedQueue.forEach((p) => p.reject(e));
						throw e;
					} finally {
						this.failedQueue = [];
						this.isRefreshing = false;
					}
				}
				throw err;
			},
		);
	}
}

/** Один singletоn, которым можно делиться где угодно */
export const httpClient = new HTTPClient();
