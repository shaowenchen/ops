<script setup>
import { useRouter } from 'vue-router';

const router = useRouter();
router.push('/tasks');
</script>
<template>
    <div class="card m-3">
    </div>
</template>

<style scoped>
.form-control {
    display: inline-block;
    width: 80%;
    margin-right: 10px;
}
</style>
