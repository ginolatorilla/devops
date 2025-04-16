<script context="module" lang="ts">
  declare function run(): Promise<void>;
  import "./lib/utils/main.wasm.d.ts";
</script>

<script lang="ts">
  import Footer from "./lib/components/Footer.svelte";
  import Header from "./lib/components/Header.svelte";
  import Main from "./lib/components/Main.svelte";
  const HeaderHeight = 10;
  const FooterHeight = 6;
</script>

<Header
  title="Gino's DevOps Tools"
  tools={["Tool 1", "Tool 2", "Tool 3"]}
  externalURL="https://github.com/ginolatorilla/devops"
  height={HeaderHeight}
/>
<Main
  toolName="Tool Name"
  toolDocumentation={["hi", "hello"]}
  headerHeight={HeaderHeight}
  footerHeight={FooterHeight}
/>
<Footer copyright="© 2025 Gino Latorilla" height={FooterHeight} />
