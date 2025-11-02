<template lang="pug">
   #app
    b-navbar(toggleable="lg" type="dark" variant="info")
      b-navbar-brand 
        | Proxy Management 
        b-badge v2.0
      b-navbar-nav
        b-nav-item(to="/") Proxy
        b-nav-item(to="/ip_logs") IP History
        b-nav-item(to="/speed_logs") Speed History
        b-nav-item(to="/visit_logs") Visit History
    router-view
</template>

<style>

</style>
