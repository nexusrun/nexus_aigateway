<!-- NEXUS AI Gateway brand artwork. The parent owns sizing and can choose
     between the compact gateway mark and the complete logo lockup. -->
<script>
  import { aigatewayPath } from "$lib/api/paths.js";

  let { variant = "mark" } = $props();

  const assetBase = import.meta.env.BASE_URL.endsWith("/")
    ? import.meta.env.BASE_URL
    : `${import.meta.env.BASE_URL}/`;

  const source = $derived(
    variant === "full"
      ? `${assetBase}nexus-ai-gateway-logo.png`
      : `${assetBase}logo.png`,
  );
</script>

<img
  src={aigatewayPath(source)}
  alt={variant === "full" ? "NEXUS AI Gateway" : ""}
  aria-hidden={variant === "full" ? undefined : "true"}
/>
