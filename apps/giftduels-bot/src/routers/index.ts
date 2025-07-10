import { Composer } from "grammy";
import { ExtendedContext } from "@/types/context";
import { paymentsRouter } from "./payments";

const root = new Composer<ExtendedContext>();
root.use(paymentsRouter);

export { root as rootRouter };
