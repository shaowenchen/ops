<script setup>
import { useLoginStore } from '@/stores';
import { router } from '@/router'

const loginStore = useLoginStore();
loginStore.clear();
router.push("/");
</script>
