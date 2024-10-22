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
    scrollToBottom(); // 发送消息后滚动到底部
  }
}

function scrollToBottom() {
  const chatBox = document.querySelector(".chat-box");
  if (chatBox) {
    chatBox.scrollTop = chatBox.scrollHeight;
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
      <el-button type="primary" @click="sendMessage" :disabled="isSending">Send</el-button>
    </div>
  </div>
</template>

<style scoped>
.container {
  border: 1px solid #ccc;
  border-radius: 10px;
  background-color: #f9f9f9;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 600px;
  margin-top: 50px;
}

.chat-box {
  flex-grow: 1;
  overflow-y: auto;
  padding-right: 10px;
  display: flex;
  flex-direction: column; /* 正常显示消息顺序 */
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
  padding: 10px 15px;
  border-radius: 15px;
  max-width: 70%;
  word-wrap: break-word;
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
  padding-top: 10px;
}

.input-box {
  flex-grow: 1;
  border-radius: 20px;
  border: 1px solid #ccc;
  transition: border-color 0.3s;
  width: 100%;
}

.input-box:focus {
  border-color: #007bff;
  outline: none;
}

.send-button {
  white-space: nowrap;
  border-radius: 20px;
}
</style>
