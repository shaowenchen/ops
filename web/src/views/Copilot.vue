<script setup>
import { ref } from "vue";
import { marked } from 'marked';
import { useCopilotStore } from "@/stores";

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
  <div class="chat-container">
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
.chat-container {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 400px;
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  border: 1px solid #ccc;
  padding: 10px;
  border-radius: 10px;
  background-color: #f9f9f9;
}

.chat-box {
  flex-grow: 1;
  overflow-y: auto;
  margin-bottom: 10px;
  padding-right: 10px;
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
  padding: 10px;
  border-radius: 10px;
  max-width: 70%;
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
}

.input-box {
  flex-grow: 1;
}

.send-button {
  white-space: nowrap;
}
</style>
