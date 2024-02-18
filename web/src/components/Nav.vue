<script setup>

</script>

<template>
    <nav class="bg-light sidebar">
        <ul class="nav flex-column">
            <li class="nav-item">
                <router-link to="/" class="nav-link">Home</router-link>
            </li>
            <li class="nav-item">
                <router-link to="/hosts" class="nav-link">Hosts</router-link>
            </li>
            <li class="nav-item">
                <router-link to="/clusters" class="nav-link">Clusters</router-link>
            </li>
            <li class="nav-item">
                <router-link to="/tasks" class="nav-link">Tasks</router-link>
            </li>
            <li class="nav-item">
                <router-link to="/taskruns" class="nav-link">TaskRuns</router-link>
            </li>
        </ul>
    </nav>
</template>

<style scoped>
.sidebar {
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    z-index: 100;
    width: 7em;
    padding: 48px 0 0;
    box-shadow: inset -1px 0 0 rgba(0, 0, 0, .1);
}
.bg-light {
    background-color: #000000 !important;
}
.nav-link {
    color: #fff !important;
}
</style>
