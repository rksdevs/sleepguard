import { mountApp } from "./app";
import { registerServiceWorker } from "./push";

const root = document.querySelector<HTMLElement>("#app");
if (!root) {
  throw new Error("#app not found");
}

void registerServiceWorker();
mountApp(root);
