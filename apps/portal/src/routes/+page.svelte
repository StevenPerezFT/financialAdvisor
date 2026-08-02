<script>
  const modules = [
    { title: 'Finanzas', href: '/finanzas' },
    { title: 'Movimientos', href: '/finanzas/movimientos' },
    { title: 'Límites', href: '/finanzas/limites' }
  ];
</script>

<svelte:head>
  <title>Financial Advisor</title>
</svelte:head>

<main style="padding: 2rem; font-family: sans-serif;">
  <h1>Financial Advisor</h1>
  <p>Panel inicial para controlar finanzas y asesor fiscal.</p>

  <ul>
    {#each modules as module}
      <li><a href={module.href}>{module.title}</a></li>
    {/each}
  </ul>
</main>
