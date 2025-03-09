import { mount } from "svelte";
import "./app.css";
import App from "./App.svelte";
import WasmLoader from "./lib/wasm-loader";

async function init() {
  const loader = new WasmLoader("./main.wasm");
  await loader.load();
  await loader.run();
  const app = mount(App, {
    target: document.getElementById("app")!,
  });
}

init();
