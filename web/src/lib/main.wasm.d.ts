declare global {
  export interface Window {
    hello(): void;
    renderGoTemplate(template: string): Promise<string>;
  }
}

export {};
