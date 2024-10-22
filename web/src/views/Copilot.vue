<script setup>
import { ref } from "vue";
import { marked } from 'marked';
import { useCopilotStore } from "@/stores";
import { useLoginStore } from "@/stores";

var loginStore = useLoginStore();
loginStore.check();

var messageList = ref([]);
var currentMessage = ref("");
var isSending = ref(false);

async function sendMessage() {
  if (currentMessage.value.trim() === "" || isSending.value) return;

  messageList.value.push({
    sender: "user",
    content: currentMessage.value,
    timestamp: new Date().toLocaleTimeString(),
  });

  isSending.value = true;
  const userMessage = currentMessage.value;
  currentMessage.value = "";

  try {
    const store = useCopilotStore();
    const res = await store.post(userMessage);

    messageList.value.push({
      sender: "bot",
      content: marked(res),
      timestamp: new Date().toLocaleTimeString(),
    });
  } catch (error) {
    messageList.value.push({
      sender: "bot",
      content: "Oops, something went wrong. Please try again later.",
      timestamp: new Date().toLocaleTimeString(),
    });
  } finally {
    isSending.value = false;
  }
}

</script>

<template>
  <div class="container">
    <div class="chat-box">
      <div v-for="(message, index) in messageList" :key="index" class="message">
        <div :class="message.sender === 'user' ? 'user-message' : 'bot-message'">
          <p v-html="message.content"></p>
          <span class="timestamp">{{ message.timestamp }}</span>
        </div>
      </div>
    </div>
    <div class="input-container">
      <el-input v-model="currentMessage" placeholder="输入消息..." class="input-box" @keyup.enter="sendMessage"
        :disabled="isSending">
      </el-input>
      <el-button type="primary" @click="sendMessage" :disabled="isSending"> Send
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.container {
  margin-left: 7em;
  border: 1px solid #ccc;
  padding: 10px;
  border-radius: 10px;
  background-color: #f9f9f9;
  max-width: 500px; /* 限制容器的最大宽度 */
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1); /* 添加阴影效果 */
}

.chat-box {
  flex-grow: 1;
  overflow-y: auto;
  margin-bottom: 10px;
  padding-right: 10px;
  max-height: 400px; /* 设置最大高度以防止过多内容溢出 */
}

.message {
  margin: 10px 0;
}

.user-message {
  text-align: right;
}

.bot-message {
  text-align: left;
}

.user-message p,
.bot-message p {
  display: inline-block;
  padding: 10px 15px; /* 增加左右内边距 */
  border-radius: 15px; /* 加大圆角 */
  max-width: 70%;
  word-wrap: break-word; /* 允许长单词换行 */
}

.user-message p {
  background-color: #007bff;
  color: white;
}

.bot-message p {
  background-color: #f1f1f1;
  color: black;
}

.timestamp {
  display: block;
  font-size: 0.8em;
  margin-top: 5px;
  color: gray;
}

.input-container {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: 10px; /* 增加顶部内边距以分隔输入框 */
}

.input-box {
  flex-grow: 1;
  border-radius: 20px; /* 圆角 */
  border: 1px solid #ccc; /* 边框颜色 */
  transition: border-color 0.3s; /* 平滑过渡 */
}

.input-box:focus {
  border-color: #007bff; /* 聚焦时的边框颜色 */
  outline: none; /* 去除默认轮廓 */
}

.send-button {
  white-space: nowrap;
  border-radius: 20px; /* 圆角 */
}
</style>
