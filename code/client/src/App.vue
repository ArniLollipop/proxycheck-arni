<template lang="pug">
   #app
    b-navbar(toggleable="lg" type="dark" variant="info")
      b-navbar-brand 
        | Proxy Management 
        b-badge v2.0
      b-navbar-nav
        b-nav-item(to="/") Proxy
        b-nav-item(to="/benchmark") Benchmark
        b-nav-item(to="/logs") Logs
    router-view
</template>

<style>

</style>
