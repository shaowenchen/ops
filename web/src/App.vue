<script setup>
import { Nav, Alert } from '@/components';
</script>

<template>
    <div class="app-container bg-light">
        <Nav />
        <Alert />
        <div class="container pt-4 pb-4">
            <router-view />
        </div>
    </div>
</template>

<style>
@import '@/assets/base.css';
</style>