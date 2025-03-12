declare global {
  export interface Window {
    hello(): void;
    renderGoTemplate(template: string, jsonData: string): Promise<string>;
  }
}

export {};
