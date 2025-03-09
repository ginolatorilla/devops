import "./wasm_exec_tiny";

declare global {
  interface Window {
    Go: any;
  }
}

if (!WebAssembly.instantiateStreaming) {
  // polyfill
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer();
    return await WebAssembly.instantiate(source, importObject);
  };
}

export default class WasmLoader {
  go: any;
  mod: WebAssembly.Module | null;
  inst: WebAssembly.Instance | null;

  constructor(src: string) {
    this.go = new window.Go();
    this.mod = null;
    this.inst = null;
  }

  async load() {
    const result = await WebAssembly.instantiateStreaming(
      fetch("./main.wasm"),
      this.go.importObject
    );
    this.mod = result.module;
    this.inst = result.instance;
  }

  async run() {
    if (!this.inst) {
      console.error("Wasm module is not loaded");
      return;
    }
    await this.go.run(this.inst);
  }
}
