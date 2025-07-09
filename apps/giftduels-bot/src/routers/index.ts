import { Composer } from "grammy";
import { ExtendedContext } from "@/types/context";

const root = new Composer<ExtendedContext>();
// root.use(userRouter);

export { root as rootRouter };
